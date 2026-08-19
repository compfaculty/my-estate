package dimria

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/httpx"
	"github.com/google/uuid"
)

const Name = "dimria"
const Version = "dimria-1"

// Kyiv on DIM.RIA: state_id=10, city_id=10.
const KyivStateID = "10"
const KyivCityID = "10"

type Collector struct {
	HTTP   *httpx.Client
	APIKey string
	Limit  int
}

func New(apiKey string, httpClient *httpx.Client) *Collector {
	if httpClient == nil {
		httpClient = httpx.New()
	}
	return &Collector{HTTP: httpClient, APIKey: apiKey, Limit: 20}
}

func (c *Collector) Name() string { return Name }

func (c *Collector) HealthCheck(ctx context.Context) error {
	if c.APIKey == "" {
		return fmt.Errorf("DIMRIA_API_KEY not set")
	}
	return nil
}

func (c *Collector) FetchNewListings(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	if c.APIKey == "" {
		return nil, cursor, fmt.Errorf("DIMRIA_API_KEY not set")
	}
	page := cursor.Page
	if page < 1 {
		page = 1
	}
	q := url.Values{}
	q.Set("api_key", c.APIKey)
	q.Set("category", "1")
	q.Set("realty_type", "2")
	q.Set("operation_type", "1")
	q.Set("state_id", KyivStateID)
	q.Set("city_id", KyivCityID)
	q.Set("page", strconv.Itoa(page))
	searchURL := "https://developers.ria.com/dom/search?" + q.Encode()
	body, status, err := c.HTTP.Get(ctx, searchURL)
	if err != nil {
		return nil, cursor, err
	}
	if status != 200 {
		return nil, cursor, fmt.Errorf("dimria search status %d", status)
	}
	ids, err := parseSearchIDs(body)
	if err != nil {
		return nil, cursor, err
	}
	out := make([]domain.RawListing, 0, len(ids))
	lim := c.Limit
	if lim <= 0 {
		lim = 20
	}
	for i, id := range ids {
		if i >= lim {
			break
		}
		raw, err := c.FetchListingDetails(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, *raw)
		select {
		case <-ctx.Done():
			return out, cursor, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	cursor.Page = page + 1
	return out, cursor, nil
}

func (c *Collector) FetchUpdates(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, cursor, nil
}

func (c *Collector) FetchListingDetails(ctx context.Context, sourceListingID string) (*domain.RawListing, error) {
	infoURL := fmt.Sprintf("https://developers.ria.com/dom/info/%s?api_key=%s", url.PathEscape(sourceListingID), url.QueryEscape(c.APIKey))
	body, status, err := c.HTTP.Get(ctx, infoURL)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("dimria info status %d", status)
	}
	return ToRaw(body, sourceListingID, status)
}

func parseSearchIDs(body []byte) ([]string, error) {
	return ParseSearchIDs(body)
}

// ParseSearchIDs extracts DIM.RIA realty IDs from a search JSON body.
func ParseSearchIDs(body []byte) ([]string, error) {
	var wrap struct {
		Items []any `json:"items"`
	}
	if err := json.Unmarshal(body, &wrap); err == nil && len(wrap.Items) > 0 {
		return idsFromAny(wrap.Items), nil
	}
	var arr []any
	if err := json.Unmarshal(body, &arr); err == nil {
		return idsFromAny(arr), nil
	}
	return nil, fmt.Errorf("dimria search: unexpected json")
}

func idsFromAny(items []any) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		switch v := it.(type) {
		case float64:
			out = append(out, strconv.FormatInt(int64(v), 10))
		case string:
			out = append(out, v)
		case json.Number:
			out = append(out, v.String())
		case map[string]any:
			if id, ok := v["realty_id"]; ok {
				out = append(out, fmt.Sprint(id))
			}
		}
	}
	return out
}

func ToRaw(body []byte, id string, status int) (*domain.RawListing, error) {
	sum := sha256.Sum256(body)
	st := status
	return &domain.RawListing{
		ID:               uuid.New(),
		Source:           Name,
		SourceListingID:  id,
		SourceURL:        fmt.Sprintf("https://dom.ria.com/uk/%s", id),
		FetchedAt:        time.Now().UTC(),
		RawPayload:       body,
		ContentHash:      hex.EncodeToString(sum[:]),
		HTTPStatus:       &st,
		CollectorVersion: Version,
	}, nil
}

var _ collectors.SourceCollector = (*Collector)(nil)
