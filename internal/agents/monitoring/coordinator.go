package monitoring

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/alexg/my-estate/internal/collectors"
	"github.com/alexg/my-estate/internal/collectors/derzhstat"
	"github.com/alexg/my-estate/internal/collectors/dimria"
	"github.com/alexg/my-estate/internal/collectors/nbu"
	"github.com/alexg/my-estate/internal/collectors/olx"
	"github.com/alexg/my-estate/internal/collectors/prozorro"
	"github.com/alexg/my-estate/internal/collectors/rieltor"
	"github.com/alexg/my-estate/internal/collectors/stub"
	"github.com/alexg/my-estate/internal/config"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/ingest"
	"golang.org/x/sync/errgroup"
)

// SourceAgent is one ADK-style monitoring node: typed tools only, no SQL/shell.
type SourceAgent interface {
	Name() string
	Run(ctx context.Context) error
}

type ingestAgent struct {
	col collectors.SourceCollector
	ing *ingest.Service
	log *slog.Logger
}

func (a ingestAgent) Name() string { return a.col.Name() }

func (a ingestAgent) Run(ctx context.Context) error {
	_, err := a.ing.Run(ctx, a.col)
	if err != nil {
		a.log.Error("source agent", "source", a.Name(), "err", err)
		if strings.Contains(strings.ToLower(err.Error()), "captcha") {
			return fmt.Errorf("degraded: %w", err)
		}
	}
	return nil
}

type fxAgent struct {
	nbu   *nbu.Client
	repos domain.Repositories
}

func (fxAgent) Name() string { return nbu.Name }

func (a fxAgent) Run(ctx context.Context) error {
	rates, err := a.nbu.FetchDate(ctx, a.nbu.Now())
	if err != nil {
		return err
	}
	for _, r := range rates {
		if err := a.repos.FX().Upsert(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

type statsAgent struct {
	c     *derzhstat.Collector
	repos domain.Repositories
}

func (statsAgent) Name() string { return derzhstat.Name }

func (a statsAgent) Run(ctx context.Context) error {
	stat, err := a.c.Fetch(ctx)
	if err != nil {
		return err
	}
	return a.repos.MarketStats().Upsert(ctx, stat)
}

type auctionAgent struct {
	c     *prozorro.Collector
	repos domain.Repositories
}

func (auctionAgent) Name() string { return prozorro.Name }

func (a auctionAgent) Run(ctx context.Context) error {
	lots, err := a.c.FetchLots(ctx)
	if err != nil {
		return err
	}
	for _, lot := range lots {
		if err := a.repos.Auctions().Upsert(ctx, lot); err != nil {
			return err
		}
	}
	return nil
}

type stubAgent struct{ c *stub.Collector }

func (a stubAgent) Name() string { return a.c.Name() }
func (a stubAgent) Run(context.Context) error {
	return nil
}

// Coordinator runs source agents in parallel (ADK workflow Join equivalent via errgroup).
// Typed tools live in Tools; collectors never see SQL or a shell.
type Coordinator struct {
	Agents []SourceAgent
	Log    *slog.Logger
}

func NewCoordinator(cfg config.Config, repos domain.Repositories, ing *ingest.Service, log *slog.Logger) *Coordinator {
	if log == nil {
		log = slog.Default()
	}
	agents := []SourceAgent{
		ingestAgent{col: rieltor.New(nil), ing: ing, log: log},
		fxAgent{nbu: nbu.New(nil), repos: repos},
		statsAgent{c: derzhstat.New(nil), repos: repos},
		auctionAgent{c: prozorro.New(nil), repos: repos},
		stubAgent{c: stub.New("lun")},
		stubAgent{c: stub.New("flatfy")},
		stubAgent{c: stub.New("address_ua")},
		stubAgent{c: stub.New("realty100")},
		stubAgent{c: stub.New("m2bomber")},
		stubAgent{c: stub.New("opendatabot")},
		stubAgent{c: stub.New("youcontrol")},
		stubAgent{c: stub.New("facebook")},
		stubAgent{c: stub.New("telegram")},
	}
	if cfg.DimriaAPIKey != "" {
		agents = append(agents, ingestAgent{col: dimria.New(cfg.DimriaAPIKey, nil), ing: ing, log: log})
	}
	if cfg.EnableOLX {
		agents = append(agents, ingestAgent{col: olx.New(), ing: ing, log: log})
	}
	return &Coordinator{Agents: agents, Log: log}
}

func (c *Coordinator) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(6)
	for _, a := range c.Agents {
		agent := a
		g.Go(func() error {
			c.Log.Info("adk source node", "source", agent.Name())
			if err := agent.Run(ctx); err != nil {
				c.Log.Error("adk source node failed", "source", agent.Name(), "err", err)
			}
			return nil
		})
	}
	return g.Wait()
}
