// Package backend represents one upstream TCP server in a pool.
package backend

import "sync/atomic"

// Backend is a single upstream identified by address.
// Healthy starts true; the Stage 3 health monitor flips it without
// removing the backend from the pool slice.
type Backend struct {
	Name    string
	Address string // host:port
	healthy atomic.Bool
}

// New creates a Backend that starts as healthy. Name defaults to address when empty.
func New(name, address string) *Backend {
	if name == "" {
		name = address
	}
	b := &Backend{Name: name, Address: address}
	b.healthy.Store(true)
	return b
}

// IsHealthy reports whether this backend is in rotation.
func (b *Backend) IsHealthy() bool { return b.healthy.Load() }

// SetHealthy marks the backend in or out of rotation.
func (b *Backend) SetHealthy(ok bool) { b.healthy.Store(ok) }
