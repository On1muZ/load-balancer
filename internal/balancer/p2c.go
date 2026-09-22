package balancer

import (
	"fmt"
	"lb/internal/backend"
	"math/rand/v2"
)

type P2C struct {
	pool *backend.Pool
}

func NewP2C(pool *backend.Pool) *P2C {
	return &P2C{pool: pool}
}

func (p2c *P2C) Next() (*backend.Backend, error) {
	nodes := p2c.pool.Alive()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no alive nodes")
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	i, j := rand.IntN(len(nodes)), rand.IntN(len(nodes))
	for i == j {
		j = rand.IntN(len(nodes))
	}
	var target *backend.Backend
	if nodes[i].AliveConns.Load()*int64(nodes[j].Weight) > nodes[j].AliveConns.Load()*int64(nodes[i].Weight) {
		target = nodes[j]
	} else {
		target = nodes[i]
	}
	return target, nil
}
