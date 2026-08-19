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

type fxRepo struct{ q querier }

func (r *fxRepo) Upsert(ctx context.Context, rate domain.FxRate) error {
	if rate.Source == "" {
		rate.Source = "nbu"
	}
	_, err := r.q.Exec(ctx, `
		INSERT INTO fx_rates (rate_date, ccy, rate_uah, fetched_at, source)
		VALUES ($1,$2,$3,now(),$4)
		ON CONFLICT (rate_date, ccy) DO UPDATE SET
			rate_uah = EXCLUDED.rate_uah,
			fetched_at = now(),
			source = EXCLUDED.source`,
		rate.RateDate, rate.CCY, rate.RateUAH, rate.Source,
	)
	return err
}

func (r *fxRepo) Get(ctx context.Context, date time.Time, ccy string) (*domain.FxRate, error) {
	row := r.q.QueryRow(ctx, `
		SELECT rate_date, ccy, rate_uah, fetched_at, source
		FROM fx_rates WHERE rate_date = $1::date AND ccy = $2`, date, ccy)
	var out domain.FxRate
	err := row.Scan(&out.RateDate, &out.CCY, &out.RateUAH, &out.FetchedAt, &out.Source)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type marketRepo struct{ q querier }

func (r *marketRepo) Upsert(ctx context.Context, stat domain.MarketStatistic) error {
	if stat.ID == uuid.Nil {
		stat.ID = uuid.New()
	}
	payload := metadataJSON(stat.Payload)
	_, err := r.q.Exec(ctx, `
		INSERT INTO market_statistics (id, source, region, market, period, index_value, payload, fetched_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,now())
		ON CONFLICT (source, region, market, period) DO UPDATE SET
			index_value = EXCLUDED.index_value,
			payload = EXCLUDED.payload,
			fetched_at = now()`,
		stat.ID, stat.Source, stat.Region, stat.Market, stat.Period, stat.IndexValue, payload,
	)
	return err
}

type auctionRepo struct{ q querier }

func (r *auctionRepo) Upsert(ctx context.Context, lot domain.AuctionLot) error {
	if lot.ID == uuid.Nil {
		lot.ID = uuid.New()
	}
	if lot.Source == "" {
		lot.Source = "prozorro"
	}
	payload := lot.RawPayload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	_, err := r.q.Exec(ctx, `
		INSERT INTO auction_lots (
			id, source, source_lot_id, source_url, title, description, city, oblast,
			start_price, currency, guarantee_deposit, auction_stage, deadline_at,
			organizer, protocol, final_price, raw_payload, last_seen_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,now()
		)
		ON CONFLICT (source, source_lot_id) DO UPDATE SET
			source_url = EXCLUDED.source_url,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			city = EXCLUDED.city,
			oblast = EXCLUDED.oblast,
			start_price = EXCLUDED.start_price,
			currency = EXCLUDED.currency,
			guarantee_deposit = EXCLUDED.guarantee_deposit,
			auction_stage = EXCLUDED.auction_stage,
			deadline_at = EXCLUDED.deadline_at,
			organizer = EXCLUDED.organizer,
			protocol = EXCLUDED.protocol,
			final_price = EXCLUDED.final_price,
			raw_payload = EXCLUDED.raw_payload,
			last_seen_at = now()`,
		lot.ID, lot.Source, lot.SourceLotID, lot.SourceURL, lot.Title, lot.Description, lot.City, lot.Oblast,
		lot.StartPrice, lot.Currency, lot.GuaranteeDeposit, lot.AuctionStage, lot.DeadlineAt,
		lot.Organizer, lot.Protocol, lot.FinalPrice, payload,
	)
	return err
}
