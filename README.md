# Estate Terminal — Ukraine (Kyiv)

Local real-estate warehouse: collectors → PostgreSQL (PostGIS) → REST API → React UI.

Market default: **Kyiv, sale apartments**. Prices stored as original + UAH/USD using **NBU rates on the observation date** (history is never recomputed).

## Run

Fire-and-forget (Postgres, migrate, mock seed, API, UI):

```bash
cp .env.example .env   # first time
make start             # returns immediately; open http://localhost:5173
make logs              # follow .run/*.log
make stop              # api/web/worker; Postgres stays up
make down              # stop processes + Postgres
```

Step by step instead:

```bash
cp .env.example .env
make db-up
make migrate
make collect          # mock Kyiv listings (offline)
make collect          # wave 2: price cuts + one removal
make api              # :8080
make web              # :5173, proxies /api
```

Live source agents (Go worker; OLX uses local Chrome via DevTools, no npm):

```bash
make worker           # DIM.RIA / NBU / Derzhstat / Prozorro / Rieltor / stubs / OLX
```

OLX needs a system Chrome/Chromium (or Edge on macOS). Optional: `CHROME_PATH=/path/to/chrome`. Allowlist `olx.ua` only; captcha/WAF stops the source (no solving).

Open http://localhost:5173

## Env

- `DIMRIA_API_KEY` — official DIM.RIA key (free tier is rate-limited)
- `ENABLE_OLX=true`
- `CHROME_PATH` — optional explicit Chrome binary
- `MARKET_CITIES=kyiv`
- NBU FX needs no key; cite NBU for open-data reuse

## Layout

- `cmd/api`, `cmd/collector`, `cmd/worker`, `cmd/migrate`
- `internal/agents/monitoring` — parallel source agents (`errgroup`; ADK graph-shaped)
- `internal/collectors` — dimria, nbu, derzhstat, prozorro, rieltor, olx, mock, stubs
- `internal/collectors/olx` — Chrome DevTools extract (chromedp)
- `deploy/docker/docker-compose.yml` — PostGIS 16 on host port **55433**
