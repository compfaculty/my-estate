package postgres

import (
	"context"
	"fmt"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store is a PostgreSQL-backed implementation of domain.Repositories.
type Store struct {
	pool *pgxpool.Pool
	q    querier
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: pool}
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

func (s *Store) Properties() domain.PropertyRepository { return &propertyRepo{q: s.q} }
func (s *Store) Listings() domain.ListingRepository    { return &listingRepo{q: s.q} }
func (s *Store) History() domain.HistoryRepository     { return &historyRepo{q: s.q} }
func (s *Store) Raw() domain.RawListingRepository      { return &rawRepo{q: s.q} }
func (s *Store) Sources() domain.SourceRepository      { return &sourceRepo{q: s.q} }
func (s *Store) FX() domain.FXRepository               { return &fxRepo{q: s.q} }
func (s *Store) MarketStats() domain.MarketStatsRepository {
	return &marketRepo{q: s.q}
}
func (s *Store) Auctions() domain.AuctionRepository { return &auctionRepo{q: s.q} }

func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context, repos domain.Repositories) error) error {
	if s.pool == nil {
		return fmt.Errorf("WithTx requires a connection pool")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	inner := &Store{pool: s.pool, q: tx}
	if err := fn(ctx, inner); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Ping(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("no pool")
	}
	return s.pool.Ping(ctx)
}

func (s *Store) Pool() *pgxpool.Pool { return s.pool }
