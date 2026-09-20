// Package server runs a virtual TCP listen endpoint bound to a backend pool.
package server

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/btranho1402/l4-load-balancer/internal/pool"
	"github.com/btranho1402/l4-load-balancer/internal/proxy"
)

// Endpoint is one virtual listen address fronting a pool.
type Endpoint struct {
	Name        string
	Listen      string
	Pool        *pool.Pool
	DialTimeout time.Duration
}

// Listener accepts connections and forwards each to a pool-selected backend.
type Listener struct {
	ep Endpoint
	ln net.Listener
	mu sync.RWMutex
}

// NewListener prepares a virtual endpoint (does not bind yet).
func NewListener(ep Endpoint) *Listener {
	return &Listener{ep: ep}
}

// ListenAndServe binds Listen and serves until the listener is closed.
func (l *Listener) ListenAndServe() error {
	ln, err := net.Listen("tcp", l.ep.Listen)
	if err != nil {
		return err
	}
	l.mu.Lock()
	l.ln = ln
	l.mu.Unlock()

	log.Printf("virtual endpoint %q listening on %s (pool=%s)",
		l.ep.Name, ln.Addr().String(), l.ep.Pool.Name)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		be, err := l.ep.Pool.Next()
		if err != nil {
			log.Printf("no backend available: %v", err)
			_ = conn.Close()
			continue
		}

		go proxy.Forward(conn, be.Address, l.ep.DialTimeout)
	}
}

// Close stops accepting new connections.
func (l *Listener) Close() error {
	l.mu.RLock()
	ln := l.ln
	l.mu.RUnlock()
	if ln == nil {
		return nil
	}
	return ln.Close()
}

// Addr returns the bound address, or empty if not listening yet.
func (l *Listener) Addr() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.ln == nil {
		return ""
	}
	return l.ln.Addr().String()
}
