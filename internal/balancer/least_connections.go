package balancer

import (
	"fmt"
	"lb/internal/backend"
)

type LeastConnections struct {
	pool *backend.Pool
}

func NewLeastConnections(pool *backend.Pool) *LeastConnections {
	return &LeastConnections{pool: pool}
}

func (lc *LeastConnections) Next() (*backend.Backend, error) {
	nodes := lc.pool.Alive()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no alive nodes")
	}
	minAliveConns := nodes[0].AliveConns.Load()
	target := nodes[0]
	for _, node := range nodes[1:] {
		if minAliveConns > node.AliveConns.Load() {
			target = node
			minAliveConns = node.AliveConns.Load()
		}
	}
	return target, nil
}
