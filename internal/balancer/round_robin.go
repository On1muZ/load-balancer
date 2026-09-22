package balancer

import (
	"fmt"
	"lb/internal/backend"
	"sync/atomic"
)

type RoundRobin struct {
	pool    *backend.Pool
	counter atomic.Uint64
}

func NewRoundRobin(pool *backend.Pool) *RoundRobin {
	return &RoundRobin{
		pool: pool,
	}
}

func (rr *RoundRobin) Next() (*backend.Backend, error) {
	nodes := rr.pool.Alive()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	index := rr.counter.Add(1) - 1
	return nodes[index%uint64(len(nodes))], nil
}
