CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    crawl_interval_seconds INTEGER NOT NULL DEFAULT 600,
    concurrency INTEGER NOT NULL DEFAULT 1,
    auth_mode TEXT NOT NULL DEFAULT 'none',
    fetch_mode TEXT NOT NULL DEFAULT 'mock',
    last_success_at TIMESTAMPTZ,
    last_error TEXT,
    last_crawl_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE collector_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES sources (id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'running',
    listings_discovered INTEGER NOT NULL DEFAULT 0,
    listings_updated INTEGER NOT NULL DEFAULT 0,
    listings_removed INTEGER NOT NULL DEFAULT 0,
    error TEXT,
    cursor JSONB
);

CREATE INDEX collector_runs_source_started_idx ON collector_runs (source_id, started_at DESC);

CREATE TABLE raw_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    source_listing_id TEXT NOT NULL,
    source_url TEXT,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    raw_payload JSONB NOT NULL,
    content_hash TEXT NOT NULL,
    http_status INTEGER,
    collector_version TEXT NOT NULL,
    UNIQUE (source, source_listing_id, content_hash)
);

CREATE INDEX raw_listings_source_listing_idx ON raw_listings (source, source_listing_id, fetched_at DESC);

CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country TEXT NOT NULL DEFAULT 'IL',
    city TEXT NOT NULL,
    district TEXT,
    neighborhood TEXT,
    street TEXT,
    house_number TEXT,
    location GEOGRAPHY (POINT, 4326),
    property_type TEXT NOT NULL,
    rooms NUMERIC(4, 1),
    bedrooms INTEGER,
    bathrooms NUMERIC(3, 1),
    floor INTEGER,
    total_floors INTEGER,
    area_m2 NUMERIC(8, 2),
    balcony_area_m2 NUMERIC(8, 2),
    lot_area_m2 NUMERIC(8, 2),
    parking_spaces INTEGER,
    construction_year INTEGER,
    renovation_state TEXT,
    elevator BOOLEAN,
    parking BOOLEAN,
    storage BOOLEAN,
    balcony BOOLEAN,
    air_conditioning BOOLEAN,
    safe_room BOOLEAN,
    accessibility BOOLEAN,
    field_provenance JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX properties_location_gix ON properties USING gist (location);
CREATE INDEX properties_city_trgm_idx ON properties USING gin (city gin_trgm_ops);
CREATE INDEX properties_neighborhood_trgm_idx ON properties USING gin (neighborhood gin_trgm_ops);
CREATE INDEX properties_street_trgm_idx ON properties USING gin (street gin_trgm_ops);
CREATE INDEX properties_type_rooms_idx ON properties (property_type, rooms);
CREATE INDEX properties_dedup_idx ON properties (lower(city), lower(street), lower(house_number), property_type);

CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties (id),
    source TEXT NOT NULL,
    source_listing_id TEXT NOT NULL,
    source_url TEXT,
    listing_type TEXT NOT NULL DEFAULT 'sale',
    asking_price NUMERIC(14, 2),
    currency TEXT NOT NULL DEFAULT 'ILS',
    price_per_m2 NUMERIC(12, 2),
    description TEXT,
    description_hash TEXT,
    seller_name TEXT,
    seller_type TEXT NOT NULL DEFAULT 'unknown',
    seller_phone_hash TEXT,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    removed_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'active',
    photos JSONB NOT NULL DEFAULT '[]'::jsonb,
    photos_hash TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    match_confidence NUMERIC(4, 3),
    UNIQUE (source, source_listing_id)
);

CREATE INDEX listings_property_idx ON listings (property_id);
CREATE INDEX listings_status_seen_idx ON listings (status, last_seen_at DESC);
CREATE INDEX listings_price_idx ON listings (asking_price);
CREATE INDEX listings_description_trgm_idx ON listings USING gin (description gin_trgm_ops);
CREATE INDEX listings_first_seen_idx ON listings (first_seen_at DESC);

CREATE TABLE listing_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL REFERENCES listings (id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    price NUMERIC(14, 2),
    currency TEXT,
    description_hash TEXT,
    status TEXT,
    seller_type TEXT,
    photos_hash TEXT
);

CREATE INDEX listing_history_listing_idx ON listing_history (listing_id, recorded_at DESC);

-- Future-phase stubs (no application code yet).
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT
);

CREATE INDEX jobs_claim_idx ON jobs (status, priority DESC, scheduled_at)
    WHERE status = 'pending';

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE,
    display_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE saved_searches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users (id),
    name TEXT NOT NULL,
    criteria JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_run_at TIMESTAMPTZ
);

CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saved_search_id UUID REFERENCES saved_searches (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status TEXT NOT NULL DEFAULT 'new'
);

CREATE TABLE alert_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saved_search_id UUID NOT NULL REFERENCES saved_searches (id),
    property_id UUID NOT NULL REFERENCES properties (id),
    listing_id UUID REFERENCES listings (id),
    matched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    notification_status TEXT NOT NULL DEFAULT 'pending',
    UNIQUE (saved_search_id, property_id, listing_id)
);

CREATE TABLE property_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties (id),
    score_type TEXT NOT NULL,
    total_score NUMERIC(5, 2),
    valuation_score NUMERIC(5, 2),
    location_score NUMERIC(5, 2),
    yield_score NUMERIC(5, 2),
    liquidity_score NUMERIC(5, 2),
    trend_score NUMERIC(5, 2),
    condition_score NUMERIC(5, 2),
    risk_penalty NUMERIC(5, 2),
    confidence NUMERIC(4, 3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    model_version TEXT
);

CREATE TABLE property_risk_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties (id),
    total_score NUMERIC(5, 2),
    risk_level TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    model_version TEXT
);

CREATE TABLE agent_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'running',
    input JSONB,
    output JSONB,
    tokens_used INTEGER,
    cost_usd NUMERIC(12, 6)
);

CREATE TABLE agent_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_run_id UUID REFERENCES agent_runs (id),
    property_id UUID REFERENCES properties (id),
    finding_type TEXT NOT NULL,
    severity TEXT,
    confidence NUMERIC(4, 3),
    description TEXT,
    classification TEXT,
    source_url TEXT,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO sources (name, enabled, crawl_interval_seconds, fetch_mode)
VALUES ('mock', TRUE, 120, 'mock');
