DELETE FROM sources WHERE name IN (
    'dimria', 'nbu', 'derzhstat', 'prozorro', 'rieltor', 'olx',
    'lun', 'flatfy', 'address_ua', 'realty100', 'm2bomber',
    'opendatabot', 'youcontrol', 'facebook', 'telegram'
);

DROP TABLE IF EXISTS risk_evidence;
DROP TABLE IF EXISTS legal_entities;
DROP TABLE IF EXISTS auction_lots;
DROP TABLE IF EXISTS market_statistics;
DROP TABLE IF EXISTS fx_rates;

ALTER TABLE listing_history
    DROP COLUMN IF EXISTS exchange_rate_date,
    DROP COLUMN IF EXISTS exchange_rate,
    DROP COLUMN IF EXISTS price_usd,
    DROP COLUMN IF EXISTS price_uah,
    DROP COLUMN IF EXISTS currency_original,
    DROP COLUMN IF EXISTS price_original;

ALTER TABLE listings
    DROP COLUMN IF EXISTS price_per_m2_uah,
    DROP COLUMN IF EXISTS price_usd,
    DROP COLUMN IF EXISTS price_uah,
    DROP COLUMN IF EXISTS fx_source,
    DROP COLUMN IF EXISTS exchange_rate_date,
    DROP COLUMN IF EXISTS exchange_rate,
    DROP COLUMN IF EXISTS currency_original,
    DROP COLUMN IF EXISTS price_original,
    ALTER COLUMN currency SET DEFAULT 'ILS';

ALTER TABLE properties
    DROP COLUMN IF EXISTS cadastral_number,
    DROP COLUMN IF EXISTS hromada,
    DROP COLUMN IF EXISTS raion,
    DROP COLUMN IF EXISTS oblast,
    ALTER COLUMN country SET DEFAULT 'IL';
