package balancer

import (
	"fmt"
	"lb/internal/backend"
	"lb/internal/config"
	"lb/internal/proxy"
	"log/slog"
	"net/http"
)

type Balancer struct {
	algorithm Algorithm
	proxy     *proxy.Proxy
	log       *slog.Logger
}

func NewBalancer(algorithm Algorithm, proxy *proxy.Proxy, log *slog.Logger) *Balancer {
	return &Balancer{algorithm: algorithm, proxy: proxy, log: log}
}

func NewAlgorithm(algorithm config.Algorithm, pool *backend.Pool) (Algorithm, error) {
	switch algorithm {
	case config.RoundRobin:
		return NewRoundRobin(pool), nil
	case config.LeastConnections:
		return NewLeastConnections(pool), nil
	case config.WeightedLeastConnections:
		return NewWeightedLeastConnections(pool), nil
	case config.WeightedRandom:
		return NewWeightedRandom(pool), nil
	case config.P2C:
		return NewP2C(pool), nil
	}
	return nil, fmt.Errorf("unknown algorithm")
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := b.algorithm.Next()
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	target.Metrics.TotalRequests.Add(1)
	target.AliveConns.Add(1)
	defer target.AliveConns.Add(-1)
	r = proxy.WithBackend(r, target)
	b.log.Debug("proxy income request",
		slog.String("income", r.Host),
		slog.String("outcome", target.URL.Host),
	)
	b.proxy.ServeHTTP(w, r)
	b.log.Debug("done")
}
