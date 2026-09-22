package balancer

import (
	"fmt"
	"lb/internal/backend"
)

type WeightedLeastConnections struct {
	pool *backend.Pool
}

func NewWeightedLeastConnections(pool *backend.Pool) *WeightedLeastConnections {
	return &WeightedLeastConnections{pool: pool}
}

func (wll *WeightedLeastConnections) Next() (*backend.Backend, error) {
	nodes := wll.pool.Alive()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no alive nodes")
	}
	minLoad := float64(nodes[0].AliveConns.Load()) / float64((nodes[0].Weight))
	target := nodes[0]
	for _, node := range nodes[1:] {
		currentLoad := float64(node.AliveConns.Load()) / float64((node.Weight))
		if minLoad > currentLoad {
			target = node
			minLoad = currentLoad
		}
	}
	return target, nil
}

