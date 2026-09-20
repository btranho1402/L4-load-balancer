// Package balancer selects which backend should receive a new connection.
package balancer

import "github.com/btranho1402/l4-load-balancer/internal/backend"

// Balancer picks the next backend. Safe for concurrent use.
type Balancer interface {
	Next(backends []*backend.Backend) *backend.Backend
}
