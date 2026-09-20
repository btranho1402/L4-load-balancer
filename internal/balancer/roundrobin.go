package balancer

import (
	"sync/atomic"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
)

// RoundRobin cycles through backends in order.
type RoundRobin struct {
	counter atomic.Uint64
}

// NewRoundRobin returns a round-robin balancer.
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next returns the next backend, or nil if the list is empty.
func (r *RoundRobin) Next(backends []*backend.Backend) *backend.Backend {
	n := len(backends)
	if n == 0 {
		return nil
	}
	i := r.counter.Add(1) - 1
	return backends[i%uint64(n)]
}
