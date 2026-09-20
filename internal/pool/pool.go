// Package pool holds backends for one virtual endpoint and delegates selection.
package pool

import (
	"fmt"
	"sync"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
	"github.com/btranho1402/l4-load-balancer/internal/balancer"
)

// Pool is a named set of backends selected by a Balancer.
type Pool struct {
	Name     string
	backends []*backend.Backend
	balancer balancer.Balancer
	mu       sync.RWMutex
}

// New creates a pool with the given selection strategy.
func New(name string, b balancer.Balancer, backends ...*backend.Backend) *Pool {
	return &Pool{
		Name:     name,
		backends: append([]*backend.Backend(nil), backends...),
		balancer: b,
	}
}

// Next returns the next backend, or an error if none are available.
func (p *Pool) Next() (*backend.Backend, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	be := p.balancer.Next(p.backends)
	if be == nil {
		return nil, fmt.Errorf("pool %q: no healthy backends", p.Name)
	}
	return be, nil
}

// Backends returns a snapshot of the pool members.
func (p *Pool) Backends() []*backend.Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*backend.Backend, len(p.backends))
	copy(out, p.backends)
	return out
}
