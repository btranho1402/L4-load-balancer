package balancer

import (
	"sync/atomic"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
)

// RoundRobin cycles through healthy backends in order.
type RoundRobin struct {
	counter atomic.Uint64
}

// NewRoundRobin returns a round-robin balancer.
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next returns the next healthy backend, or nil if none are healthy.
func (r *RoundRobin) Next(backends []*backend.Backend) *backend.Backend {
	healthy := make([]*backend.Backend, 0, len(backends))
	for _, b := range backends {
		if b != nil && b.IsHealthy() {
			healthy = append(healthy, b)
		}
	}
	n := len(healthy)
	if n == 0 {
		return nil
	}
	i := r.counter.Add(1) - 1
	return healthy[i%uint64(n)]
}
