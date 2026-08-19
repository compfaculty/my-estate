package rieltor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/httpx"
	"github.com/google/uuid"
)

const Name = "rieltor"
const Version = "rieltor-1"

var (
	hrefRE  = regexp.MustCompile(`href="(https://rieltor\.ua/[^"]+/[0-9]+/?)"`)
	jsonLD  = regexp.MustCompile(`<script type="application/ld\+json">([\s\S]*?)</script>`)
	priceRE = regexp.MustCompile(`"price"\s*:\s*"?([0-9.]+)"?`)
)

type Collector struct {
	HTTP *httpx.Client
	URL  string
}

func New(httpClient *httpx.Client) *Collector {
	if httpClient == nil {
		httpClient = httpx.New()
	}
	return &Collector{
		HTTP: httpClient,
		URL:  "https://rieltor.ua/kiev/flats-sale/",
	}
}

func (c *Collector) Name() string { return Name }

func (c *Collector) HealthCheck(context.Context) error { return nil }

func (c *Collector) FetchUpdates(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, cursor, nil
}

func (c *Collector) FetchNewListings(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	body, status, err := c.HTTP.Get(ctx, c.URL)
	if err != nil {
		return nil, cursor, err
	}
	if status >= 400 {
		return nil, cursor, fmt.Errorf("rieltor status %d", status)
	}
	hrefs := hrefRE.FindAllStringSubmatch(string(body), 15)
	out := make([]domain.RawListing, 0, len(hrefs))
	seen := map[string]bool{}
	for _, h := range hrefs {
		u := h[1]
		if seen[u] {
			continue
		}
		seen[u] = true
		raw, err := c.FetchListingDetails(ctx, u)
		if err != nil {
			continue
		}
		out = append(out, *raw)
		select {
		case <-ctx.Done():
			return out, cursor, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	cursor.Page++
	return out, cursor, nil
}

func (c *Collector) FetchListingDetails(ctx context.Context, sourceListingID string) (*domain.RawListing, error) {
	pageURL := sourceListingID
	if !strings.HasPrefix(pageURL, "http") {
		pageURL = "https://rieltor.ua" + sourceListingID
	}
	body, status, err := c.HTTP.Get(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("rieltor listing status %d", status)
	}
	payload := ParseListingHTML(pageURL, body)
	rawJSON, _ := json.Marshal(payload)
	sum := sha256.Sum256(rawJSON)
	st := status
	id := payload.SourceListingID
	return &domain.RawListing{
		ID:               uuid.New(),
		Source:           Name,
		SourceListingID:  id,
		SourceURL:        pageURL,
		FetchedAt:        time.Now().UTC(),
		RawPayload:       rawJSON,
		ContentHash:      hex.EncodeToString(sum[:]),
		HTTPStatus:       &st,
		CollectorVersion: Version,
	}, nil
}

type Payload struct {
	SourceListingID string   `json:"source_listing_id"`
	SourceURL       string   `json:"source_url"`
	City            string   `json:"city"`
	Oblast          string   `json:"oblast"`
	District        string   `json:"district"`
	Street          string   `json:"street"`
	HouseNumber     string   `json:"house_number"`
	Lat             float64  `json:"lat"`
	Lng             float64  `json:"lng"`
	Rooms           float64  `json:"rooms"`
	AreaM2          float64  `json:"area_m2"`
	Floor           int      `json:"floor"`
	TotalFloors     int      `json:"total_floors"`
	Price           float64  `json:"price"`
	Currency        string   `json:"currency"`
	Description     string   `json:"description"`
	Photos          []string `json:"photos"`
	Status          string   `json:"status"`
	PropertyType    string   `json:"property_type"`
	ListingType     string   `json:"listing_type"`
}

func ParseListingHTML(pageURL string, html []byte) Payload {
	s := string(html)
	p := Payload{
		SourceURL:       pageURL,
		City:            "Київ",
		Oblast:          "Київ",
		Currency:        "UAH",
		Status:          "active",
		PropertyType:    "apartment",
		ListingType:     "sale",
		SourceListingID: pageURL,
	}
	if m := jsonLD.FindStringSubmatch(s); len(m) == 2 {
		var obj map[string]any
		if json.Unmarshal([]byte(m[1]), &obj) == nil {
			if n, ok := obj["name"].(string); ok {
				p.Description = n
			}
		}
	}
	if m := priceRE.FindStringSubmatch(s); len(m) == 2 {
		var f float64
		fmt.Sscanf(m[1], "%f", &f)
		p.Price = f
	}
	parts := strings.Split(strings.Trim(pageURL, "/"), "/")
	if len(parts) > 0 {
		p.SourceListingID = parts[len(parts)-1]
	}
	return p
}

var _ collectors.SourceCollector = (*Collector)(nil)
