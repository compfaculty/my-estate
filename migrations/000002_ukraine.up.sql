ALTER TABLE properties
    ALTER COLUMN country SET DEFAULT 'UA',
    ADD COLUMN IF NOT EXISTS oblast TEXT,
    ADD COLUMN IF NOT EXISTS raion TEXT,
    ADD COLUMN IF NOT EXISTS hromada TEXT,
    ADD COLUMN IF NOT EXISTS cadastral_number TEXT;

UPDATE properties SET country = 'UA' WHERE country = 'IL';

ALTER TABLE listings
    ALTER COLUMN currency SET DEFAULT 'UAH',
    ADD COLUMN IF NOT EXISTS price_original NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS currency_original TEXT,
    ADD COLUMN IF NOT EXISTS exchange_rate NUMERIC(18, 8),
    ADD COLUMN IF NOT EXISTS exchange_rate_date DATE,
    ADD COLUMN IF NOT EXISTS fx_source TEXT,
    ADD COLUMN IF NOT EXISTS price_uah NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS price_usd NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS price_per_m2_uah NUMERIC(12, 2);

UPDATE listings
SET
    price_original = COALESCE(price_original, asking_price),
    currency_original = COALESCE(currency_original, currency),
    price_uah = COALESCE(price_uah, CASE WHEN currency = 'UAH' THEN asking_price ELSE NULL END);

ALTER TABLE listing_history
    ADD COLUMN IF NOT EXISTS price_original NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS currency_original TEXT,
    ADD COLUMN IF NOT EXISTS price_uah NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS price_usd NUMERIC(14, 2),
    ADD COLUMN IF NOT EXISTS exchange_rate NUMERIC(18, 8),
    ADD COLUMN IF NOT EXISTS exchange_rate_date DATE;

CREATE TABLE IF NOT EXISTS fx_rates (
    rate_date DATE NOT NULL,
    ccy TEXT NOT NULL,
    rate_uah NUMERIC(18, 8) NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source TEXT NOT NULL DEFAULT 'nbu',
    PRIMARY KEY (rate_date, ccy)
);

CREATE TABLE IF NOT EXISTS market_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    region TEXT,
    market TEXT,
    period TEXT NOT NULL,
    index_value NUMERIC(18, 6),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source, region, market, period)
);

CREATE TABLE IF NOT EXISTS auction_lots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL DEFAULT 'prozorro',
    source_lot_id TEXT NOT NULL,
    source_url TEXT,
    title TEXT,
    description TEXT,
    city TEXT,
    oblast TEXT,
    start_price NUMERIC(18, 2),
    currency TEXT,
    guarantee_deposit NUMERIC(18, 2),
    auction_stage TEXT,
    deadline_at TIMESTAMPTZ,
    organizer TEXT,
    protocol TEXT,
    final_price NUMERIC(18, 2),
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source, source_lot_id)
);

CREATE TABLE IF NOT EXISTS legal_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    edrpou TEXT UNIQUE,
    name TEXT,
    status TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS risk_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID REFERENCES properties (id),
    listing_id UUID REFERENCES listings (id),
    legal_entity_id UUID REFERENCES legal_entities (id),
    finding_type TEXT NOT NULL,
    classification TEXT NOT NULL DEFAULT 'unverified_signal',
    confidence NUMERIC(4, 3),
    description TEXT,
    source_url TEXT,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO sources (name, enabled, crawl_interval_seconds, fetch_mode) VALUES
    ('dimria', TRUE, 600, 'api'),
    ('nbu', TRUE, 86400, 'api'),
    ('derzhstat', TRUE, 604800, 'api'),
    ('prozorro', TRUE, 900, 'api'),
    ('rieltor', TRUE, 900, 'http'),
    ('olx', TRUE, 900, 'chrome'),
    ('lun', FALSE, 3600, 'stub'),
    ('flatfy', FALSE, 3600, 'stub'),
    ('address_ua', FALSE, 3600, 'stub'),
    ('realty100', FALSE, 3600, 'stub'),
    ('m2bomber', FALSE, 3600, 'stub'),
    ('opendatabot', FALSE, 86400, 'stub'),
    ('youcontrol', FALSE, 86400, 'stub'),
    ('facebook', FALSE, 3600, 'stub'),
    ('telegram', FALSE, 3600, 'stub')
ON CONFLICT (name) DO NOTHING;
