package mock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
)

const Version = "mock-1"
const Name = "mock"

// Collector emits a deterministic Kyiv (UA) sale catalog in UAH.
// Wave 2+ applies price cuts and removes mock-030.
type Collector struct {
	wave atomic.Int32
}

func New() *Collector { return &Collector{} }

// SetWave prepares the next FetchNewListings call to emit the given wave (1-based).
func (c *Collector) SetWave(n int) {
	if n < 1 {
		n = 1
	}
	c.wave.Store(int32(n - 1))
}

func (c *Collector) Name() string { return Name }

func (c *Collector) HealthCheck(context.Context) error { return nil }

func (c *Collector) FetchNewListings(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	run := c.wave.Add(1)
	items := fixtures()
	if run >= 2 {
		applySecondRun(items)
	}
	raw := make([]domain.RawListing, 0, len(items))
	for _, p := range items {
		r, err := toRaw(p)
		if err != nil {
			return nil, cursor, err
		}
		raw = append(raw, r)
	}
	cursor.Page = 1
	return raw, cursor, nil
}

func (c *Collector) FetchListingDetails(_ context.Context, sourceListingID string) (*domain.RawListing, error) {
	for _, p := range fixtures() {
		if p.SourceListingID == sourceListingID {
			r, err := toRaw(p)
			if err != nil {
				return nil, err
			}
			return &r, nil
		}
	}
	return nil, fmt.Errorf("listing %s not found", sourceListingID)
}

func (c *Collector) FetchUpdates(context.Context, domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, domain.CrawlCursor{}, nil
}

func applySecondRun(items []Payload) {
	for i := range items {
		switch items[i].SourceListingID {
		case "mock-001":
			items[i].Price = 2470000
			items[i].Description = "Світла 3-кімнатна біля Контрактової. Ціну знижено повторно."
		case "mock-008":
			items[i].Price = 1990000
		case "mock-018":
			items[i].Price = 1590000
		case "mock-030":
			items[i].Status = "removed"
		}
	}
}

func toRaw(p Payload) (domain.RawListing, error) {
	body, err := json.Marshal(p)
	if err != nil {
		return domain.RawListing{}, err
	}
	sum := sha256.Sum256(body)
	status := 200
	return domain.RawListing{
		ID:               uuid.New(),
		Source:           Name,
		SourceListingID:  p.SourceListingID,
		SourceURL:        p.SourceURL,
		FetchedAt:        time.Now().UTC(),
		RawPayload:       body,
		ContentHash:      hex.EncodeToString(sum[:]),
		HTTPStatus:       &status,
		CollectorVersion: Version,
	}, nil
}

var _ collectors.SourceCollector = (*Collector)(nil)
