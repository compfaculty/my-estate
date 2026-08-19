package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rawRepo struct{ q querier }

func (r *rawRepo) InsertIfNew(ctx context.Context, raw domain.RawListing) (bool, error) {
	if raw.ID == uuid.Nil {
		raw.ID = uuid.New()
	}
	if raw.FetchedAt.IsZero() {
		raw.FetchedAt = time.Now().UTC()
	}
	var id uuid.UUID
	err := r.q.QueryRow(ctx, `
		INSERT INTO raw_listings (
			id, source, source_listing_id, source_url, fetched_at,
			raw_payload, content_hash, http_status, collector_version
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (source, source_listing_id, content_hash) DO NOTHING
		RETURNING id`,
		raw.ID, raw.Source, raw.SourceListingID, raw.SourceURL, raw.FetchedAt,
		raw.RawPayload, raw.ContentHash, raw.HTTPStatus, raw.CollectorVersion,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type sourceRepo struct{ q querier }

func (r *sourceRepo) GetByName(ctx context.Context, name string) (*domain.Source, error) {
	row := r.q.QueryRow(ctx, `
		SELECT id, name, enabled, crawl_interval_seconds, concurrency, auth_mode, fetch_mode,
		       last_success_at, last_error, last_crawl_at, created_at, updated_at
		FROM sources WHERE name = $1`, name)
	s, err := scanSource(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

func (r *sourceRepo) List(ctx context.Context) ([]domain.Source, error) {
	rows, err := r.q.Query(ctx, `
		SELECT s.id, s.name, s.enabled, s.crawl_interval_seconds, s.concurrency, s.auth_mode, s.fetch_mode,
		       s.last_success_at, s.last_error, s.last_crawl_at, s.created_at, s.updated_at,
		       coalesce(r.listings_discovered, 0), coalesce(r.listings_updated, 0), coalesce(r.status, '')
		FROM sources s
		LEFT JOIN LATERAL (
			SELECT listings_discovered, listings_updated, status
			FROM collector_runs
			WHERE source_id = s.id
			ORDER BY started_at DESC
			LIMIT 1
		) r ON TRUE
		ORDER BY s.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Source, 0)
	for rows.Next() {
		var s domain.Source
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Enabled, &s.CrawlIntervalSeconds, &s.Concurrency, &s.AuthMode, &s.FetchMode,
			&s.LastSuccessAt, &s.LastError, &s.LastCrawlAt, &s.CreatedAt, &s.UpdatedAt,
			&s.ListingsDiscovered, &s.ListingsUpdated, &s.LastRunStatus,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *sourceRepo) TouchCrawl(ctx context.Context, name string, success bool, errMsg string) error {
	if success {
		_, err := r.q.Exec(ctx, `
			UPDATE sources SET last_success_at = now(), last_crawl_at = now(), last_error = NULL, updated_at = now()
			WHERE name = $1`, name)
		return err
	}
	_, err := r.q.Exec(ctx, `
		UPDATE sources SET last_crawl_at = now(), last_error = $2, updated_at = now()
		WHERE name = $1`, name, errMsg)
	return err
}

func (r *sourceRepo) InsertRun(ctx context.Context, run *domain.CollectorRun) error {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	_, err := r.q.Exec(ctx, `
		INSERT INTO collector_runs (id, source_id, started_at, status, cursor)
		VALUES ($1,$2,$3,$4,$5)`,
		run.ID, run.SourceID, run.StartedAt, run.Status, run.Cursor,
	)
	return err
}

func (r *sourceRepo) FinishRun(ctx context.Context, run domain.CollectorRun) error {
	now := time.Now().UTC()
	_, err := r.q.Exec(ctx, `
		UPDATE collector_runs SET
			finished_at = $2, status = $3, listings_discovered = $4,
			listings_updated = $5, listings_removed = $6, error = $7, cursor = $8
		WHERE id = $1`,
		run.ID, now, run.Status, run.ListingsDiscovered, run.ListingsUpdated,
		run.ListingsRemoved, run.Error, run.Cursor,
	)
	return err
}

func scanSource(row pgx.Row) (*domain.Source, error) {
	var s domain.Source
	err := row.Scan(
		&s.ID, &s.Name, &s.Enabled, &s.CrawlIntervalSeconds, &s.Concurrency, &s.AuthMode, &s.FetchMode,
		&s.LastSuccessAt, &s.LastError, &s.LastCrawlAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
