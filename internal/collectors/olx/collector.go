package olx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
)

const Name = "olx"
const Version = "olx-chrome-1"

type CaptchaError struct{ Msg string }

func (e CaptchaError) Error() string { return e.Msg }

type Collector struct {
	Chrome *Chrome
}

func New() *Collector {
	return &Collector{Chrome: NewChrome()}
}

func (c *Collector) Name() string { return Name }

func (c *Collector) HealthCheck(context.Context) error {
	if FindChrome() == "" {
		return fmt.Errorf("chrome/chromium not found; set CHROME_PATH")
	}
	return nil
}

func (c *Collector) FetchUpdates(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, cursor, nil
}

func (c *Collector) FetchNewListings(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	search := "https://www.olx.ua/uk/nedvizhimost/kvartiry/prodazha/kiev/"
	extracted, err := c.extract(ctx, search)
	if err != nil {
		return nil, cursor, err
	}
	ids := listingIDsFromExtract(extracted)
	out := make([]domain.RawListing, 0, len(ids))
	for i, id := range ids {
		if i >= 8 {
			break
		}
		raw, err := c.FetchListingDetails(ctx, id)
		if err != nil {
			if _, ok := err.(CaptchaError); ok {
				return out, cursor, err
			}
			continue
		}
		out = append(out, *raw)
	}
	cursor.Page++
	return out, cursor, nil
}

func (c *Collector) FetchListingDetails(ctx context.Context, sourceListingID string) (*domain.RawListing, error) {
	pageURL := sourceListingID
	if !strings.HasPrefix(pageURL, "http") {
		pageURL = "https://www.olx.ua/d/uk/obyavlenie/" + sourceListingID
	}
	extracted, err := c.extract(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	payload := ParseExtract(pageURL, extracted)
	rawJSON, _ := json.Marshal(payload)
	sum := sha256.Sum256(rawJSON)
	st := 200
	return &domain.RawListing{
		ID:               uuid.New(),
		Source:           Name,
		SourceListingID:  payload.SourceListingID,
		SourceURL:        pageURL,
		FetchedAt:        time.Now().UTC(),
		RawPayload:       rawJSON,
		ContentHash:      hex.EncodeToString(sum[:]),
		HTTPStatus:       &st,
		CollectorVersion: Version,
	}, nil
}

type Extract struct {
	Blocked bool           `json:"blocked"`
	HTML    string         `json:"html"`
	JSON    map[string]any `json:"json"`
}

func (c *Collector) extract(ctx context.Context, target string) (Extract, error) {
	u, err := url.Parse(target)
	if err != nil || !AllowedHost(u.Hostname()) {
		return Extract{}, fmt.Errorf("url not allowlisted")
	}
	if c.Chrome == nil {
		c.Chrome = NewChrome()
	}
	out, err := c.Chrome.Extract(ctx, target)
	if err != nil {
		return out, err
	}
	if out.Blocked {
		return out, CaptchaError{Msg: "olx captcha or waf"}
	}
	return out, nil
}

func AllowedHost(h string) bool {
	h = strings.ToLower(h)
	return h == "olx.ua" || h == "www.olx.ua"
}

type Payload struct {
	SourceListingID string   `json:"source_listing_id"`
	SourceURL       string   `json:"source_url"`
	City            string   `json:"city"`
	Oblast          string   `json:"oblast"`
	Street          string   `json:"street"`
	HouseNumber     string   `json:"house_number"`
	Rooms           float64  `json:"rooms"`
	AreaM2          float64  `json:"area_m2"`
	Floor           int      `json:"floor"`
	Price           float64  `json:"price"`
	Currency        string   `json:"currency"`
	Description     string   `json:"description"`
	Photos          []string `json:"photos"`
	Status          string   `json:"status"`
	PropertyType    string   `json:"property_type"`
	ListingType     string   `json:"listing_type"`
}

func ParseExtract(pageURL string, ext Extract) Payload {
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
	if ext.JSON != nil {
		if s, ok := ext.JSON["id"].(string); ok {
			p.SourceListingID = s
		}
		if s, ok := ext.JSON["title"].(string); ok {
			p.Description = s
		}
		if n, ok := ext.JSON["price"].(float64); ok {
			p.Price = n
		}
		if s, ok := ext.JSON["currency"].(string); ok && s != "" {
			p.Currency = s
		}
	}
	if p.SourceListingID == pageURL {
		parts := strings.Split(strings.Trim(pageURL, "/"), "-")
		if len(parts) > 0 {
			p.SourceListingID = strings.TrimSuffix(parts[len(parts)-1], ".html")
		}
	}
	return p
}

func listingIDsFromExtract(ext Extract) []string {
	ids := make([]string, 0)
	if ext.JSON != nil {
		if arr, ok := ext.JSON["hrefs"].([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					ids = append(ids, s)
				}
			}
		}
		if arr, ok := ext.JSON["hrefs"].([]string); ok {
			ids = append(ids, arr...)
		}
	}
	return ids
}

var _ collectors.SourceCollector = (*Collector)(nil)
