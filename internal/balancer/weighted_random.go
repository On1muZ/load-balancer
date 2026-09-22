package balancer

import (
	"fmt"
	"lb/internal/backend"
	"math/rand/v2"
)

type WeightedRandom struct {
	pool *backend.Pool
}

func NewWeightedRandom(pool *backend.Pool) *WeightedRandom {
	return &WeightedRandom{pool: pool}
}

func (wr *WeightedRandom) Next() (*backend.Backend, error) {
	nodes := wr.pool.Alive()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no alive nodes")
	}
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += int(node.Weight.Load())
	}
	if totalWeight <= 0 {
		return nodes[rand.IntN(len(nodes))], nil
	}
	target := rand.IntN(totalWeight)
	for _, node := range nodes {
		if target < int(node.Weight.Load()) {
			return node, nil
		}
		target -= int(node.Weight.Load())
	}
	return nodes[0], nil
}
