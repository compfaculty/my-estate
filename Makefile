.PHONY: db-up db-down migrate collect worker api web test tidy start stop restart logs down

DATABASE_URL ?= postgres://estate:estate@localhost:55433/estate?sslmode=disable
export DATABASE_URL
export ENABLE_DEV_ENDPOINTS ?= true
export CORS_ORIGINS ?= http://localhost:5173

db-up:
	docker compose -f deploy/docker/docker-compose.yml up -d

db-down:
	docker compose -f deploy/docker/docker-compose.yml down

migrate:
	go run ./cmd/migrate

collect:
	go run ./cmd/collector --once

worker:
	go run ./cmd/worker --once

api:
	go run ./cmd/api

web:
	cd web && npm run dev

test:
	go test ./...

tidy:
	go mod tidy

RUN_DIR := .run

# Fire-and-forget local stack: Postgres, migrate, mock seed, API (:8080), UI (:5173).
# Optional live worker loop: make start WORKER=1
start: db-up
	@mkdir -p $(RUN_DIR)
	@echo "waiting for postgres…"
	@i=0; \
	until docker exec estate-postgres pg_isready -U estate -d estate >/dev/null 2>&1; do \
		i=$$((i+1)); \
		if [ $$i -ge 40 ]; then echo "postgres did not become ready"; exit 1; fi; \
		sleep 0.5; \
	done
	@$(MAKE) migrate
	@if [ ! -d web/node_modules ]; then cd web && npm install; fi
	@$(MAKE) collect
	@$(MAKE) --no-print-directory _bg-api
	@$(MAKE) --no-print-directory _bg-web
	@if [ "$(WORKER)" = "1" ]; then $(MAKE) --no-print-directory _bg-worker; fi
	@echo
	@echo "running in background"
	@echo "  UI   http://localhost:5173"
	@echo "  API  http://localhost:8080/api/health"
	@echo "  logs make logs    stop make stop"
	@if [ "$(WORKER)" != "1" ]; then echo "  live collectors: make start WORKER=1"; fi

_bg-api:
	@if [ -f $(RUN_DIR)/api.pid ] && kill -0 $$(cat $(RUN_DIR)/api.pid) 2>/dev/null; then \
		echo "api already running pid $$(cat $(RUN_DIR)/api.pid)"; \
	else \
		go build -o $(RUN_DIR)/estate-api ./cmd/api; \
		nohup $(RUN_DIR)/estate-api >$(RUN_DIR)/api.log 2>&1 & echo $$! >$(RUN_DIR)/api.pid; \
		echo "api pid $$(cat $(RUN_DIR)/api.pid)  log $(RUN_DIR)/api.log"; \
	fi

_bg-web:
	@if [ -f $(RUN_DIR)/web.pid ] && kill -0 $$(cat $(RUN_DIR)/web.pid) 2>/dev/null; then \
		echo "web already running pid $$(cat $(RUN_DIR)/web.pid)"; \
	else \
		cd web && nohup npm run dev >../$(RUN_DIR)/web.log 2>&1 & echo $$! >../$(RUN_DIR)/web.pid; \
		echo "web pid $$(cat $(RUN_DIR)/web.pid)  log $(RUN_DIR)/web.log"; \
	fi

_bg-worker:
	@if [ -f $(RUN_DIR)/worker.pid ] && kill -0 $$(cat $(RUN_DIR)/worker.pid) 2>/dev/null; then \
		echo "worker already running pid $$(cat $(RUN_DIR)/worker.pid)"; \
	else \
		go build -o $(RUN_DIR)/estate-worker ./cmd/worker; \
		nohup $(RUN_DIR)/estate-worker >$(RUN_DIR)/worker.log 2>&1 & echo $$! >$(RUN_DIR)/worker.pid; \
		echo "worker pid $$(cat $(RUN_DIR)/worker.pid)  log $(RUN_DIR)/worker.log"; \
	fi

stop:
	@for name in api web worker; do \
		if [ -f $(RUN_DIR)/$$name.pid ]; then \
			pid=$$(cat $(RUN_DIR)/$$name.pid); \
			echo "stopping $$name pid $$pid"; \
			for child in $$(pgrep -P $$pid 2>/dev/null); do kill $$child 2>/dev/null || true; done; \
			kill $$pid 2>/dev/null || true; \
			sleep 0.2; \
			kill -9 $$pid 2>/dev/null || true; \
			rm -f $(RUN_DIR)/$$name.pid; \
		fi; \
	done
	@echo "stopped api/web/worker (postgres still up; make db-down to drop it)"

restart: stop start

logs:
	@mkdir -p $(RUN_DIR)
	@touch $(RUN_DIR)/api.log $(RUN_DIR)/web.log $(RUN_DIR)/worker.log
	@tail -n 80 -f $(RUN_DIR)/api.log $(RUN_DIR)/web.log $(RUN_DIR)/worker.log

down: stop db-down
