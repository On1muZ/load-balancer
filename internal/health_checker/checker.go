package healthchecker

import (
	"context"
	"lb/internal/backend"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Checker struct {
	pool     *backend.Pool
	client   *http.Client
	interval time.Duration
	log      *slog.Logger
}

func NewChecker(
	pool *backend.Pool,
	interval time.Duration,
	timeout time.Duration,
	log *slog.Logger,
) *Checker {
	return &Checker{
		pool: pool,
		client: &http.Client{
			Timeout: timeout,
		},
		interval: interval,
		log:      log,
	}
}

func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	c.checkAll(ctx)
	for {
		select {
		case <-ctx.Done():
			c.log.Info("health checker stopped")
			return
		case <-ticker.C:
			c.checkAll(ctx)
		}
	}
}

func (c *Checker) checkAll(ctx context.Context) {
	nodes := c.pool.All()
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(b *backend.Backend) {
			defer wg.Done()
			c.checkNode(ctx, b)
		}(node)
	}
	wg.Wait()
}

func (c *Checker) checkNode(ctx context.Context, b *backend.Backend) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.URL.String(), nil)
	if err != nil {
		c.markDead(b, "failed to create request")
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.markDead(b, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		c.markAlive(b)
	} else {
		c.markDead(b, resp.Status)
	}
}

func (c *Checker) markDead(b *backend.Backend, reason string) {
	wasAlive := c.pool.SetAlive(b, false)
	if wasAlive {
		b.Metrics.AddHealthEvent(false)
		c.log.Warn("backend went down", slog.String("reason", reason), slog.String("url", b.URL.String()))
	}
}

func (c *Checker) markAlive(b *backend.Backend) {
	wasAlive := c.pool.SetAlive(b, true)
	if !wasAlive {
		b.Metrics.AddHealthEvent(true)
		c.log.Info("backend recovered", slog.String("url", b.URL.String()))
	}
}
