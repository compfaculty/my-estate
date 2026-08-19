package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexg/my-estate/internal/collectors/mock"
	"github.com/alexg/my-estate/internal/config"
	"github.com/alexg/my-estate/internal/ingest"
	"github.com/alexg/my-estate/internal/repository/postgres"
)

func main() {
	once := flag.Bool("once", false, "run a single ingest cycle and exit")
	wave := flag.Int("wave", 0, "mock data wave (0 = auto from prior collector runs)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := postgres.New(pool)
	svc := ingest.New(store, slog.Default())
	col := mock.New()

	runOnce := func() error {
		w := *wave
		if w <= 0 {
			src, err := store.Sources().GetByName(ctx, col.Name())
			if err != nil {
				return err
			}
			if src != nil && src.LastCrawlAt != nil {
				w = 2
			} else {
				w = 1
			}
		}
		col.SetWave(w)
		_, err := svc.Run(ctx, col)
		return err
	}

	if *once {
		if err := runOnce(); err != nil {
			slog.Error("ingest", "err", err)
			os.Exit(1)
		}
		return
	}

	ticker := time.NewTicker(cfg.CollectorInterval)
	defer ticker.Stop()
	if err := runOnce(); err != nil {
		slog.Error("ingest", "err", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := runOnce(); err != nil {
				slog.Error("ingest", "err", err)
			}
		}
	}
}
