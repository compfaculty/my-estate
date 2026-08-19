package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexg/my-estate/internal/agents/monitoring"
	"github.com/alexg/my-estate/internal/config"
	"github.com/alexg/my-estate/internal/ingest"
	"github.com/alexg/my-estate/internal/repository/postgres"
)

func main() {
	once := flag.Bool("once", false, "run all source agents once")
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
	ing := ingest.New(store, slog.Default())
	coord := monitoring.NewCoordinator(cfg, store, ing, slog.Default())

	run := func() {
		if err := coord.Run(ctx); err != nil {
			slog.Error("coordinator", "err", err)
		}
	}
	if *once {
		run()
		return
	}
	run()
	t := time.NewTicker(cfg.CollectorInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			run()
		}
	}
}
