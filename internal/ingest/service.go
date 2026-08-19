package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/collectors/nbu"
	"github.com/alexg/my-estate/internal/dedup"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/normalize"
)

type Result struct {
	Discovered int
	Updated    int
	Removed    int
	Unchanged  int
	RawStored  int
}

type Service struct {
	repos    domain.Repositories
	resolver *dedup.Resolver
	nbu      *nbu.Client
	log      *slog.Logger
}

// SetNBU replaces the NBU client (pass nil to disable live FX fetch; used in tests).
func (s *Service) SetNBU(c *nbu.Client) {
	s.nbu = c
}

func New(repos domain.Repositories, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		repos:    repos,
		resolver: dedup.NewResolver(),
		nbu:      nbu.New(nil),
		log:      log,
	}
}

func (s *Service) Run(ctx context.Context, collector collectors.SourceCollector) (Result, error) {
	src, err := s.repos.Sources().GetByName(ctx, collector.Name())
	if err != nil {
		return Result{}, fmt.Errorf("load source: %w", err)
	}
	if src == nil {
		return Result{}, fmt.Errorf("unknown source %q", collector.Name())
	}

	run := domain.CollectorRun{
		SourceID: src.ID,
		Status:   "running",
	}
	if err := s.repos.Sources().InsertRun(ctx, &run); err != nil {
		return Result{}, fmt.Errorf("insert run: %w", err)
	}

	raws, cursor, err := collector.FetchNewListings(ctx, domain.CrawlCursor{})
	if err != nil {
		msg := err.Error()
		run.Status = "error"
		run.Error = &msg
		_ = s.repos.Sources().FinishRun(ctx, run)
		_ = s.repos.Sources().TouchCrawl(ctx, collector.Name(), false, msg)
		return Result{}, fmt.Errorf("fetch listings: %w", err)
	}

	updates, cursor, err := collector.FetchUpdates(ctx, cursor)
	if err != nil {
		msg := err.Error()
		run.Status = "error"
		run.Error = &msg
		_ = s.repos.Sources().FinishRun(ctx, run)
		_ = s.repos.Sources().TouchCrawl(ctx, collector.Name(), false, msg)
		return Result{}, fmt.Errorf("fetch updates: %w", err)
	}
	raws = append(raws, updates...)

	var result Result
	for _, raw := range raws {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		item, err := s.process(ctx, raw)
		if err != nil {
			s.log.Error("ingest listing", "source", raw.Source, "id", raw.SourceListingID, "err", err)
			continue
		}
		result.RawStored += item.rawStored
		result.Discovered += item.discovered
		result.Updated += item.updated
		result.Removed += item.removed
		result.Unchanged += item.unchanged
	}

	cursorJSON, _ := json.Marshal(cursor)
	run.Cursor = cursorJSON
	run.Status = "success"
	run.ListingsDiscovered = result.Discovered
	run.ListingsUpdated = result.Updated
	run.ListingsRemoved = result.Removed
	if err := s.repos.Sources().FinishRun(ctx, run); err != nil {
		return result, fmt.Errorf("finish run: %w", err)
	}
	if err := s.repos.Sources().TouchCrawl(ctx, collector.Name(), true, ""); err != nil {
		return result, err
	}
	s.log.Info("ingest complete",
		"source", collector.Name(),
		"discovered", result.Discovered,
		"updated", result.Updated,
		"removed", result.Removed,
		"unchanged", result.Unchanged,
	)
	return result, nil
}

type itemResult struct {
	rawStored  int
	discovered int
	updated    int
	removed    int
	unchanged  int
}

func (s *Service) process(ctx context.Context, raw domain.RawListing) (itemResult, error) {
	var out itemResult
	parsed, err := normalize.Parse(raw.Source, raw.RawPayload)
	if err != nil {
		return out, fmt.Errorf("normalize: %w", err)
	}
	s.applyFX(ctx, raw.FetchedAt, &parsed)

	err = s.repos.WithTx(ctx, func(ctx context.Context, repos domain.Repositories) error {
		inserted, err := repos.Raw().InsertIfNew(ctx, raw)
		if err != nil {
			return fmt.Errorf("raw insert: %w", err)
		}
		if inserted {
			out.rawStored = 1
		}

		prop, confidence, err := s.resolver.Resolve(ctx, repos.Properties(), parsed.Property)
		if err != nil {
			return fmt.Errorf("dedup: %w", err)
		}

		listing := domain.Listing{
			PropertyID:       prop.ID,
			Source:           parsed.Source,
			SourceListingID:  parsed.SourceListingID,
			SourceURL:        parsed.SourceURL,
			ListingType:      parsed.ListingType,
			AskingPrice:      parsed.AskingPrice,
			Currency:         parsed.Currency,
			PricePerM2:       parsed.PricePerM2(),
			PriceOriginal:    parsed.PriceOriginal,
			CurrencyOriginal: parsed.CurrencyOriginal,
			ExchangeRate:     parsed.ExchangeRate,
			ExchangeRateDate: parsed.ExchangeRateDate,
			FxSource:         parsed.FxSource,
			PriceUAH:         parsed.PriceUAH,
			PriceUSD:         parsed.PriceUSD,
			PricePerM2UAH:    parsed.PricePerM2UAH,
			Description:      parsed.Description,
			DescriptionHash:  parsed.DescriptionHash(),
			SellerName:       parsed.SellerName,
			SellerType:       parsed.SellerType,
			SellerPhoneHash:  parsed.SellerPhoneHash,
			Status:           parsed.Status,
			Photos:           parsed.Photos,
			PhotosHash:       parsed.PhotosHash(),
			Metadata:         parsed.Metadata,
			MatchConfidence:  &confidence,
		}
		if listing.Status == domain.ListingStatusRemoved {
			now := time.Now().UTC()
			listing.RemovedAt = &now
		}

		created, previous, err := repos.Listings().Upsert(ctx, &listing)
		if err != nil {
			return fmt.Errorf("upsert listing: %w", err)
		}

		eventType := domain.EventListingUpdated
		writeHistory := false
		switch {
		case created:
			eventType = domain.EventListingDiscovered
			writeHistory = true
			out.discovered = 1
		case previous != nil && listing.ChangedFrom(*previous):
			if listing.Status == domain.ListingStatusRemoved && previous.Status != domain.ListingStatusRemoved {
				eventType = domain.EventListingRemoved
				out.removed = 1
			} else {
				out.updated = 1
			}
			writeHistory = true
		default:
			out.unchanged = 1
		}

		if writeHistory {
			snap := listing.Snapshot()
			if err := repos.History().Append(ctx, snap); err != nil {
				return fmt.Errorf("history: %w", err)
			}
			s.log.Info("listing event",
				"type", eventType,
				"listing_id", listing.ID,
				"property_id", listing.PropertyID,
				"source", listing.Source,
				"source_listing_id", listing.SourceListingID,
			)
		}
		return nil
	})
	return out, err
}

func (s *Service) applyFX(ctx context.Context, fetched time.Time, parsed *normalize.Listing) {
	if parsed.AskingPrice == nil {
		return
	}
	orig := *parsed.AskingPrice
	parsed.PriceOriginal = &orig
	if parsed.Currency == "" {
		parsed.Currency = "UAH"
	}
	parsed.CurrencyOriginal = parsed.Currency
	day := fetched
	if day.IsZero() {
		day = time.Now().UTC()
	}
	rate, err := s.repos.FX().Get(ctx, day, "USD")
	if err != nil {
		s.log.Warn("fx get", "err", err)
	}
	if rate == nil && s.nbu != nil {
		rates, ferr := s.nbu.FetchDate(ctx, day)
		if ferr != nil {
			s.log.Warn("nbu fetch", "err", ferr)
		} else {
			for _, r := range rates {
				_ = s.repos.FX().Upsert(ctx, r)
				if r.CCY == "USD" {
					cp := r
					rate = &cp
				}
			}
		}
	}
	if parsed.CurrencyOriginal == "UAH" {
		parsed.PriceUAH = &orig
	}
	if rate != nil && rate.RateUAH > 0 {
		uah, usd := nbu.Convert(orig, parsed.CurrencyOriginal, rate.RateUAH)
		parsed.PriceUAH = &uah
		parsed.PriceUSD = &usd
		parsed.ExchangeRate = &rate.RateUAH
		d := rate.RateDate
		parsed.ExchangeRateDate = &d
		parsed.FxSource = "nbu"
	}
	if parsed.PriceUAH != nil && parsed.Property.AreaM2 != nil && *parsed.Property.AreaM2 > 0 {
		v := *parsed.PriceUAH / *parsed.Property.AreaM2
		parsed.PricePerM2UAH = &v
	}
}
