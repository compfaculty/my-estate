package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const listingCols = `
	id, property_id, source, source_listing_id, source_url, listing_type,
	asking_price, currency, price_per_m2, description, description_hash,
	seller_name, seller_type, seller_phone_hash,
	first_seen_at, last_seen_at, removed_at, status, photos, photos_hash,
	metadata, match_confidence,
	price_original, currency_original, exchange_rate, exchange_rate_date, fx_source,
	price_uah, price_usd, price_per_m2_uah`

func scanListing(row pgx.Row) (*domain.Listing, error) {
	var (
		l                                     domain.Listing
		photos                                []byte
		metadata                              []byte
		sellerName, phoneHash, desc, descHash *string
		url, photosHash, curOrig, fxSrc       *string
	)
	err := row.Scan(
		&l.ID, &l.PropertyID, &l.Source, &l.SourceListingID, &url, &l.ListingType,
		&l.AskingPrice, &l.Currency, &l.PricePerM2, &desc, &descHash,
		&sellerName, &l.SellerType, &phoneHash,
		&l.FirstSeenAt, &l.LastSeenAt, &l.RemovedAt, &l.Status, &photos, &photosHash,
		&metadata, &l.MatchConfidence,
		&l.PriceOriginal, &curOrig, &l.ExchangeRate, &l.ExchangeRateDate, &fxSrc,
		&l.PriceUAH, &l.PriceUSD, &l.PricePerM2UAH,
	)
	if err != nil {
		return nil, err
	}
	l.SourceURL = deref(url)
	l.Description = deref(desc)
	l.DescriptionHash = deref(descHash)
	l.SellerName = deref(sellerName)
	l.SellerPhoneHash = deref(phoneHash)
	l.PhotosHash = deref(photosHash)
	l.CurrencyOriginal = deref(curOrig)
	l.FxSource = deref(fxSrc)
	if len(photos) > 0 {
		_ = json.Unmarshal(photos, &l.Photos)
	}
	l.Metadata = metadata
	return &l, nil
}

type listingRepo struct{ q querier }

func (r *listingRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	row := r.q.QueryRow(ctx, `SELECT `+listingCols+` FROM listings WHERE id = $1`, id)
	l, err := scanListing(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return l, err
}

func (r *listingRepo) GetBySourceID(ctx context.Context, source, sourceListingID string) (*domain.Listing, error) {
	row := r.q.QueryRow(ctx,
		`SELECT `+listingCols+` FROM listings WHERE source = $1 AND source_listing_id = $2`,
		source, sourceListingID,
	)
	l, err := scanListing(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return l, err
}

func (r *listingRepo) ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]domain.Listing, error) {
	rows, err := r.q.Query(ctx,
		`SELECT `+listingCols+` FROM listings WHERE property_id = $1 ORDER BY last_seen_at DESC`,
		propertyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Listing, 0)
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *listingRepo) Upsert(ctx context.Context, listing *domain.Listing) (bool, *domain.Listing, error) {
	existing, err := r.GetBySourceID(ctx, listing.Source, listing.SourceListingID)
	if err != nil {
		return false, nil, err
	}
	now := time.Now().UTC()
	if existing == nil {
		if listing.ID == uuid.Nil {
			listing.ID = uuid.New()
		}
		if listing.FirstSeenAt.IsZero() {
			listing.FirstSeenAt = now
		}
		listing.LastSeenAt = now
		const q = `
			INSERT INTO listings (
				id, property_id, source, source_listing_id, source_url, listing_type,
				asking_price, currency, price_per_m2, description, description_hash,
				seller_name, seller_type, seller_phone_hash,
				first_seen_at, last_seen_at, removed_at, status, photos, photos_hash,
				metadata, match_confidence,
				price_original, currency_original, exchange_rate, exchange_rate_date, fx_source,
				price_uah, price_usd, price_per_m2_uah
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,
				$23,$24,$25,$26,$27,$28,$29,$30
			)`
		_, err := r.q.Exec(ctx, q,
			listing.ID, listing.PropertyID, listing.Source, listing.SourceListingID, listing.SourceURL, listing.ListingType,
			listing.AskingPrice, listing.Currency, listing.PricePerM2, nullString(listing.Description), listing.DescriptionHash,
			nullString(listing.SellerName), listing.SellerType, nullString(listing.SellerPhoneHash),
			listing.FirstSeenAt, listing.LastSeenAt, listing.RemovedAt, listing.Status,
			photosJSON(listing.Photos), listing.PhotosHash, metadataJSON(listing.Metadata), listing.MatchConfidence,
			listing.PriceOriginal, nullString(listing.CurrencyOriginal), listing.ExchangeRate, listing.ExchangeRateDate, nullString(listing.FxSource),
			listing.PriceUAH, listing.PriceUSD, listing.PricePerM2UAH,
		)
		return true, nil, err
	}

	prev := *existing
	listing.ID = existing.ID
	listing.FirstSeenAt = existing.FirstSeenAt
	listing.LastSeenAt = now
	if listing.PropertyID == uuid.Nil {
		listing.PropertyID = existing.PropertyID
	}
	const q = `
		UPDATE listings SET
			property_id = $2, source_url = $3, listing_type = $4,
			asking_price = $5, currency = $6, price_per_m2 = $7,
			description = $8, description_hash = $9, seller_name = $10,
			seller_type = $11, seller_phone_hash = $12, last_seen_at = $13,
			removed_at = $14, status = $15, photos = $16, photos_hash = $17,
			metadata = $18, match_confidence = $19,
			price_original = $20, currency_original = $21, exchange_rate = $22,
			exchange_rate_date = $23, fx_source = $24, price_uah = $25, price_usd = $26,
			price_per_m2_uah = $27
		WHERE id = $1`
	_, err = r.q.Exec(ctx, q,
		listing.ID, listing.PropertyID, listing.SourceURL, listing.ListingType,
		listing.AskingPrice, listing.Currency, listing.PricePerM2,
		nullString(listing.Description), listing.DescriptionHash, nullString(listing.SellerName),
		listing.SellerType, nullString(listing.SellerPhoneHash), listing.LastSeenAt,
		listing.RemovedAt, listing.Status, photosJSON(listing.Photos), listing.PhotosHash,
		metadataJSON(listing.Metadata), listing.MatchConfidence,
		listing.PriceOriginal, nullString(listing.CurrencyOriginal), listing.ExchangeRate,
		listing.ExchangeRateDate, nullString(listing.FxSource), listing.PriceUAH, listing.PriceUSD,
		listing.PricePerM2UAH,
	)
	return false, &prev, err
}
