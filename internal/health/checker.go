// Package health runs active TCP probes and updates backend health flags.
package health

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
)

// Config controls active health monitoring.
type Config struct {
	Interval       time.Duration
	Timeout        time.Duration
	HealthyAfter   int // consecutive successes to mark healthy
	UnhealthyAfter int // consecutive failures to mark unhealthy
}

// DefaultConfig returns sensible lab defaults.
func DefaultConfig() Config {
	return Config{
		Interval:       1 * time.Second,
		Timeout:        500 * time.Millisecond,
		HealthyAfter:   2,
		UnhealthyAfter: 2,
	}
}

// Monitor periodically dials each backend. Failed backends leave rotation
// (IsHealthy=false); recovered backends return automatically — no pool edits.
type Monitor struct {
	cfg      Config
	backends []*backend.Backend

	mu      sync.Mutex
	streaks map[string]int // +success / -failure streak
}

// NewMonitor watches the given backends.
func NewMonitor(cfg Config, backends ...*backend.Backend) *Monitor {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig().Interval
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	if cfg.HealthyAfter <= 0 {
		cfg.HealthyAfter = DefaultConfig().HealthyAfter
	}
	if cfg.UnhealthyAfter <= 0 {
		cfg.UnhealthyAfter = DefaultConfig().UnhealthyAfter
	}
	return &Monitor{
		cfg:      cfg,
		backends: backends,
		streaks:  make(map[string]int),
	}
}

// Run blocks until ctx is cancelled, probing on cfg.Interval.
func (m *Monitor) Run(ctx context.Context) {
	t := time.NewTicker(m.cfg.Interval)
	defer t.Stop()

	m.probeAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.probeAll(ctx)
		}
	}
}

func (m *Monitor) probeAll(ctx context.Context) {
	var wg sync.WaitGroup
	for _, b := range m.backends {
		wg.Add(1)
		go func(b *backend.Backend) {
			defer wg.Done()
			m.probeOne(ctx, b)
		}(b)
	}
	wg.Wait()
}

func (m *Monitor) probeOne(ctx context.Context, b *backend.Backend) {
	d := net.Dialer{Timeout: m.cfg.Timeout}
	conn, err := d.DialContext(ctx, "tcp", b.Address)
	if err != nil {
		m.record(b, false)
		return
	}
	_ = conn.Close()
	m.record(b, true)
}

func (m *Monitor) record(b *backend.Backend, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := b.Address
	s := m.streaks[key]

	if ok {
		if s < 0 {
			s = 0
		}
		s++
		m.streaks[key] = s
		if !b.IsHealthy() && s >= m.cfg.HealthyAfter {
			b.SetHealthy(true)
			log.Printf("backend restored: %s (%s)", b.Name, b.Address)
		}
		return
	}

	if s > 0 {
		s = 0
	}
	s--
	m.streaks[key] = s
	if b.IsHealthy() && -s >= m.cfg.UnhealthyAfter {
		b.SetHealthy(false)
		log.Printf("backend removed from rotation: %s (%s)", b.Name, b.Address)
	}
}
