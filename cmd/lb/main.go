package main

import (
	"context"
	"fmt"
	"lb/internal/backend"
	"lb/internal/balancer"
	"lb/internal/config"
	healthchecker "lb/internal/health_checker"
	"lb/internal/logger"
	"lb/internal/proxy"
	"lb/internal/transport"
	"lb/internal/watcher"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString("config error: " + err.Error() + "\n")
		os.Exit(1)
	}

	log := logger.New(cfg)
	slog.SetDefault(log)
	log.Info("config and logger successfully inited")

	backends, err := backend.LoadNodesFromFile(cfg.NodesFilePath)
	if err != nil {
		log.Error("failed to load nodes", "error", err)
		os.Exit(1)
	}
	pool := backend.NewPool(backends)
	log.Info("nodes loaded", "count", len(pool.All()))

	algo, err := balancer.NewAlgorithm(cfg.Algorithm, pool)
	if err != nil {
		log.Error("failed to init algorithm", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	checker := healthchecker.NewChecker(pool, cfg.HealthChecker.Interval, cfg.HealthChecker.Timeout, log)
	go checker.Start(ctx)
	log.Info("health checker started", slog.Duration("interval", cfg.HealthChecker.Interval), slog.Duration("timeout", cfg.HealthChecker.Timeout))

	nodesWatcher := watcher.NewNodesWatcher(cfg.NodesFilePath, pool, log)
	go func() {
		if err := nodesWatcher.Start(ctx); err != nil {
			log.Error("nodes watcher failed", "error", err)
		}
	}()
	log.Info("nodes watcher started", slog.String("path", cfg.NodesFilePath))

	transport := transport.NewTransport()
	proxy := proxy.NewProxy(transport)
	lb := balancer.NewBalancer(algo, proxy, log)

	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port),
		Handler:           lb,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
	}

	go func() {
		log.Info("starting load balancer", "addr", server.Addr, "algorithm", cfg.Algorithm)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server fatal error", "error", err)
			panic("server fatal error")
		}
	}()

	<-ctx.Done()
	log.Info("shutting down gracefully...")

	for _, b := range pool.All() {
		uptime := b.Metrics.Uptime()
		log.Info("backend metrics", slog.String("url", b.URL.String()), slog.Int64("total_requests", b.Metrics.TotalRequests.Load()), slog.Float64("uptime", uptime.AliveTime.Seconds()), slog.Float64("uptime_persentage", uptime.AlivePercentage))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced shutdown", "error", err)
	}
}
