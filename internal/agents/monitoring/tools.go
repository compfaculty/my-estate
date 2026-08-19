package monitoring

import (
	"context"
	"time"

	"github.com/alexg/my-estate/internal/domain"
)

// Typed tools exposed to source agents. No raw SQL, no shell.
type Tools struct {
	FetchSearchPage      func(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error)
	FetchListingDetails  func(ctx context.Context, sourceListingID string) (*domain.RawListing, error)
	SaveRawAdvertisement func(ctx context.Context, raw domain.RawListing) error
	GetFxRate            func(ctx context.Context, day time.Time, ccy string) (*domain.FxRate, error)
	HealthCheck          func(ctx context.Context) error
}
