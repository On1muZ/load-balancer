package backend

import (
	"sync"
	"sync/atomic"
)

type Pool struct {
	mu       sync.RWMutex
	backends []*Backend
	alive    atomic.Pointer[[]*Backend]
}

func NewPool(backends []*Backend) *Pool {
	p := &Pool{
		backends: backends,
	}
	p.RebuildAlive()
	return p
}

func (p *Pool) All() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Backend, len(p.backends))
	copy(out, p.backends)
	return out
}

func (p *Pool) Alive() []*Backend {
	ptr := p.alive.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}

func (p *Pool) SetAlive(b *Backend, alive bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	wasAlive := b.Alive.Swap(alive)
	if wasAlive == alive {
		return alive
	}
	p.RebuildAlive()
	return wasAlive
}

func (p *Pool) RebuildAlive() {
	alive := make([]*Backend, 0, len(p.backends))
	for _, b := range p.backends {
		if b.Alive.Load() {
			alive = append(alive, b)
		}
	}
	p.alive.Store(&alive)
}
