package ingest_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/alexg/my-estate/internal/collectors/mock"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/ingest"
	"github.com/alexg/my-estate/internal/repository/postgres"
	"github.com/alexg/my-estate/migrations"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestIngestMockPipeline(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resetSchema(t, dsn)

	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := postgres.New(pool)
	svc := ingest.New(store, nil)
	svc.SetNBU(nil)
	if err := store.FX().Upsert(ctx, domain.FxRate{
		RateDate: time.Now().UTC(),
		CCY:      "USD",
		RateUAH:  41.5,
		Source:   "nbu",
	}); err != nil {
		t.Fatal(err)
	}
	col := mock.New()

	r1, err := svc.Run(ctx, col)
	if err != nil {
		t.Fatal(err)
	}
	if r1.Discovered == 0 {
		t.Fatal("expected discoveries on first run")
	}

	items, total, err := store.Properties().Search(ctx, domain.SearchFilter{City: "Київ", Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatal("expected searchable properties")
	}

	var listingCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM listings`).Scan(&listingCount); err != nil {
		t.Fatal(err)
	}
	if listingCount <= total {
		t.Fatalf("expected more listings than properties (dedup); listings=%d properties=%d", listingCount, total)
	}

	r2, err := svc.Run(ctx, col)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Updated == 0 && r2.Removed == 0 {
		t.Fatalf("expected updates on second wave, got %+v", r2)
	}

	var hist int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM listing_history`).Scan(&hist); err != nil {
		t.Fatal(err)
	}
	if hist <= listingCount {
		t.Fatalf("expected history rows beyond initial snapshots; history=%d listings=%d", hist, listingCount)
	}

	found := false
	for _, p := range items {
		if p.Neighborhood == "Поділ" && p.Rooms != nil && *p.Rooms == 3 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected Поділ 3-room in search")
	}
}

type staticCollector struct {
	name string
	raws []domain.RawListing
}

func (s staticCollector) Name() string                      { return s.name }
func (s staticCollector) HealthCheck(context.Context) error { return nil }
func (s staticCollector) FetchListingDetails(context.Context, string) (*domain.RawListing, error) {
	return nil, nil
}
func (s staticCollector) FetchUpdates(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, cursor, nil
}
func (s staticCollector) FetchNewListings(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return s.raws, cursor, nil
}

func TestHistoricalFXDoesNotMove(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resetSchema(t, dsn)

	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := postgres.New(pool)
	svc := ingest.New(store, nil)
	svc.SetNBU(nil)

	day1 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := store.FX().Upsert(ctx, domain.FxRate{RateDate: day1, CCY: "USD", RateUAH: 40, Source: "nbu"}); err != nil {
		t.Fatal(err)
	}
	if err := store.FX().Upsert(ctx, domain.FxRate{RateDate: day2, CCY: "USD", RateUAH: 42, Source: "nbu"}); err != nil {
		t.Fatal(err)
	}

	raw := func(id string, fetched time.Time) domain.RawListing {
		body := []byte(`{"source_listing_id":"` + id + `","source_url":"https://mock.local/` + id + `","city":"Київ","neighborhood":"Печерськ","street":"Липська","house_number":"1","lat":50.44,"lng":30.53,"property_type":"apartment","rooms":2,"bedrooms":1,"bathrooms":1,"floor":3,"total_floors":9,"area_m2":50,"listing_type":"sale","price":100000,"currency":"USD","description":"fx","seller_name":"A","seller_type":"private","phone":"0500000000","photos":[],"status":"active"}`)
		st := 200
		return domain.RawListing{
			Source:           mock.Name,
			SourceListingID:  id,
			SourceURL:        "https://mock.local/" + id,
			FetchedAt:        fetched,
			RawPayload:       body,
			ContentHash:      id + fetched.Format("20060102"),
			HTTPStatus:       &st,
			CollectorVersion: "test",
		}
	}

	if _, err := svc.Run(ctx, staticCollector{name: mock.Name, raws: []domain.RawListing{raw("fx-1", day1)}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Run(ctx, staticCollector{name: mock.Name, raws: []domain.RawListing{raw("fx-2", day2)}}); err != nil {
		t.Fatal(err)
	}

	l1, err := store.Listings().GetBySourceID(ctx, mock.Name, "fx-1")
	if err != nil || l1 == nil || l1.PriceUAH == nil {
		t.Fatalf("fx-1: %+v %v", l1, err)
	}
	l2, err := store.Listings().GetBySourceID(ctx, mock.Name, "fx-2")
	if err != nil || l2 == nil || l2.PriceUAH == nil {
		t.Fatalf("fx-2: %+v %v", l2, err)
	}
	if *l1.PriceUAH != 4_000_000 || *l2.PriceUAH != 4_200_000 {
		t.Fatalf("uah %v vs %v", *l1.PriceUAH, *l2.PriceUAH)
	}

	if err := store.FX().Upsert(ctx, domain.FxRate{RateDate: time.Now().UTC(), CCY: "USD", RateUAH: 50, Source: "nbu"}); err != nil {
		t.Fatal(err)
	}
	l1b, _ := store.Listings().GetBySourceID(ctx, mock.Name, "fx-1")
	if *l1b.PriceUAH != *l1.PriceUAH {
		t.Fatal("stored UAH must not follow today's rate")
	}
	hist, err := store.History().ListByProperty(ctx, l1.PropertyID)
	if err != nil || len(hist) == 0 {
		t.Fatal(err)
	}
	if hist[0].PriceUAH != nil && *hist[0].PriceUAH != *l1.PriceUAH && *hist[0].PriceUAH != *l2.PriceUAH {
		// either listing's snapshot is fine; none should be 5e6
		t.Fatalf("history moved with today's FX: %v", *hist[0].PriceUAH)
	}
	for _, h := range hist {
		if h.PriceUAH != nil && *h.PriceUAH == 5_000_000 {
			t.Fatal("history recomputed with new rate")
		}
	}
}

func TestCrossSourceDedupKyiv(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resetSchema(t, dsn)
	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := postgres.New(pool)
	svc := ingest.New(store, nil)
	svc.SetNBU(nil)
	_ = store.FX().Upsert(ctx, domain.FxRate{RateDate: time.Now().UTC(), CCY: "USD", RateUAH: 41, Source: "nbu"})

	mockRaw, err := json.Marshal(map[string]any{
		"source_listing_id": "m-cross",
		"source_url":        "https://mock.local/m-cross",
		"city":              "Київ",
		"neighborhood":      "Поділ",
		"street":            "Хорива",
		"house_number":      "12",
		"lat":               50.4691,
		"lng":               30.5134,
		"property_type":     "apartment",
		"rooms":             3,
		"bedrooms":          2,
		"bathrooms":         1,
		"floor":             2,
		"total_floors":      5,
		"area_m2":           75,
		"listing_type":      "sale",
		"price":             2680000,
		"currency":          "UAH",
		"description":       "a",
		"seller_name":       "A",
		"seller_type":       "private",
		"phone":             "0501111001",
		"photos":            []string{},
		"status":            "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	dimRaw := []byte(`{"realty_id":"d-cross","price":2700000,"currency_type":"UAH","rooms_count":3,"total_square_meters":75,"street_name":"Хорива","building_number_str":"12","floor":2,"floors_count":5,"latitude":50.4691,"longitude":30.5134}`)
	st := 200
	if _, err := svc.Run(ctx, staticCollector{name: mock.Name, raws: []domain.RawListing{{
		Source: mock.Name, SourceListingID: "m-cross", FetchedAt: time.Now().UTC(),
		RawPayload: mockRaw, ContentHash: "m-cross", HTTPStatus: &st, CollectorVersion: "t",
	}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Run(ctx, staticCollector{name: "dimria", raws: []domain.RawListing{{
		Source: "dimria", SourceListingID: "d-cross", FetchedAt: time.Now().UTC(),
		RawPayload: dimRaw, ContentHash: "d-cross", HTTPStatus: &st, CollectorVersion: "t",
	}}}); err != nil {
		t.Fatal(err)
	}
	var props, listings int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM properties`).Scan(&props); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM listings`).Scan(&listings); err != nil {
		t.Fatal(err)
	}
	if props != 1 {
		t.Fatalf("expected 1 property after cross-source dedup, got %d", props)
	}
	if listings != 2 {
		t.Fatalf("expected 2 listings, got %d", listings)
	}
}

func resetSchema(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatal(err)
	}

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		t.Fatal(err)
	}
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "pgx5", driver)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}
}
