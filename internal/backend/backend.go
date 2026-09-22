package backend

import (
	"lb/internal/metrics"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL        *url.URL
	Weight     int
	Alive      atomic.Bool
	AliveConns atomic.Int64
	Metrics    *metrics.Metrics
}

func NewBackend(URL string, Weight int) (*Backend, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	b := &Backend{
		URL:     u,
		Weight:  Weight,
		Metrics: &metrics.Metrics{StartTimestamp: time.Now().Unix()},
	}
	b.Alive.Store(true)
	return b, nil
}

func (b *Backend) MakeDead() {
	b.Alive.Store(false)
}

func (b *Backend) MakeAlive() {
	b.Alive.Store(true)
}
