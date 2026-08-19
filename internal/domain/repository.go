package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PropertyRepository interface {
	Create(ctx context.Context, p *Property) error
	Get(ctx context.Context, id uuid.UUID) (*Property, error)
	FindMatch(ctx context.Context, candidate Property) (*Property, error)
	Search(ctx context.Context, filter SearchFilter) ([]PropertySearchResult, int, error)
}

type ListingRepository interface {
	Get(ctx context.Context, id uuid.UUID) (*Listing, error)
	GetBySourceID(ctx context.Context, source, sourceListingID string) (*Listing, error)
	ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]Listing, error)
	Upsert(ctx context.Context, listing *Listing) (created bool, previous *Listing, err error)
}

type HistoryRepository interface {
	Append(ctx context.Context, h ListingHistory) error
	ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]ListingHistory, error)
}

type RawListingRepository interface {
	InsertIfNew(ctx context.Context, raw RawListing) (inserted bool, err error)
}

type SourceRepository interface {
	GetByName(ctx context.Context, name string) (*Source, error)
	List(ctx context.Context) ([]Source, error)
	TouchCrawl(ctx context.Context, name string, success bool, errMsg string) error
	InsertRun(ctx context.Context, run *CollectorRun) error
	FinishRun(ctx context.Context, run CollectorRun) error
}

type Repositories interface {
	Properties() PropertyRepository
	Listings() ListingRepository
	History() HistoryRepository
	Raw() RawListingRepository
	Sources() SourceRepository
	FX() FXRepository
	MarketStats() MarketStatsRepository
	Auctions() AuctionRepository
	WithTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type FXRepository interface {
	Upsert(ctx context.Context, rate FxRate) error
	Get(ctx context.Context, date time.Time, ccy string) (*FxRate, error)
}

type MarketStatsRepository interface {
	Upsert(ctx context.Context, stat MarketStatistic) error
}

type AuctionRepository interface {
	Upsert(ctx context.Context, lot AuctionLot) error
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }
