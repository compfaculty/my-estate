package collectors

import (
	"context"

	"github.com/alexg/my-estate/internal/domain"
)

// SourceCollector retrieves listings from one external (or mock) source.
// Implementations must not contain investment or risk business logic.
type SourceCollector interface {
	Name() string
	FetchNewListings(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error)
	FetchListingDetails(ctx context.Context, sourceListingID string) (*domain.RawListing, error)
	FetchUpdates(ctx context.Context, cursor domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error)
	HealthCheck(ctx context.Context) error
}
