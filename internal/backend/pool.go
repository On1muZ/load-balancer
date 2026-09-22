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

func (p *Pool) Update(fresh []*Backend) (added, removed int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	old := make(map[string]*Backend, len(p.backends))
	for _, b := range p.backends {
		old[b.URL.String()] = b
	}

	next := make([]*Backend, 0, len(fresh))
	for _, nb := range fresh {
		key := nb.URL.String()
		if ob, ok := old[key]; ok {
			delete(old, key)
			ob.Weight.Store(nb.Weight.Load())
			next = append(next, ob)
			continue
		}
		added++
		next = append(next, nb)
	}
	removed = len(old)
	p.backends = next
	p.RebuildAlive()
	return added, removed
}
