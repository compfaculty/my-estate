package postgres

import (
	"context"
	"time"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type historyRepo struct{ q querier }

func (r *historyRepo) Append(ctx context.Context, h domain.ListingHistory) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	if h.RecordedAt.IsZero() {
		h.RecordedAt = time.Now().UTC()
	}
	_, err := r.q.Exec(ctx, `
		INSERT INTO listing_history (
			id, listing_id, recorded_at, price, currency, description_hash, status, seller_type, photos_hash,
			price_original, currency_original, price_uah, price_usd, exchange_rate, exchange_rate_date
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		h.ID, h.ListingID, h.RecordedAt, h.Price, nullString(h.Currency),
		h.DescriptionHash, h.Status, h.SellerType, h.PhotosHash,
		h.PriceOriginal, nullString(h.CurrencyOriginal), h.PriceUAH, h.PriceUSD, h.ExchangeRate, h.ExchangeRateDate,
	)
	return err
}

func (r *historyRepo) ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]domain.ListingHistory, error) {
	rows, err := r.q.Query(ctx, `
		SELECT h.id, h.listing_id, h.recorded_at, h.price, h.currency,
		       h.description_hash, h.status, h.seller_type, h.photos_hash,
		       h.price_original, h.currency_original, h.price_uah, h.price_usd,
		       h.exchange_rate, h.exchange_rate_date
		FROM listing_history h
		JOIN listings l ON l.id = h.listing_id
		WHERE l.property_id = $1
		ORDER BY h.recorded_at DESC`, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanHistoryRows(rows)
}

func scanHistoryRows(rows pgx.Rows) ([]domain.ListingHistory, error) {
	out := make([]domain.ListingHistory, 0)
	for rows.Next() {
		var h domain.ListingHistory
		var currency, curOrig *string
		if err := rows.Scan(
			&h.ID, &h.ListingID, &h.RecordedAt, &h.Price, &currency,
			&h.DescriptionHash, &h.Status, &h.SellerType, &h.PhotosHash,
			&h.PriceOriginal, &curOrig, &h.PriceUAH, &h.PriceUSD,
			&h.ExchangeRate, &h.ExchangeRateDate,
		); err != nil {
			return nil, err
		}
		h.Currency = deref(currency)
		h.CurrencyOriginal = deref(curOrig)
		out = append(out, h)
	}
	return out, rows.Err()
}
