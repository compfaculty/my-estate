package stub

import (
	"context"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/domain"
)

type Collector struct {
	SourceName string
}

func New(name string) *Collector { return &Collector{SourceName: name} }

func (c *Collector) Name() string { return c.SourceName }

func (c *Collector) HealthCheck(context.Context) error { return nil }

// Status is the ADK health token for unimplemented sources.
func (c *Collector) Status() string { return "stub" }

func (c *Collector) FetchNewListings(context.Context, domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, domain.CrawlCursor{}, nil
}

func (c *Collector) FetchListingDetails(context.Context, string) (*domain.RawListing, error) {
	return nil, nil
}

func (c *Collector) FetchUpdates(context.Context, domain.CrawlCursor) ([]domain.RawListing, domain.CrawlCursor, error) {
	return nil, domain.CrawlCursor{}, nil
}

var _ collectors.SourceCollector = (*Collector)(nil)
