package health_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
	"github.com/btranho1402/l4-load-balancer/internal/health"
)

func TestMonitorRemovesAndRestores(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	be := backend.New("t", ln.Addr().String())
	cfg := health.Config{
		Interval:       40 * time.Millisecond,
		Timeout:        80 * time.Millisecond,
		HealthyAfter:   1,
		UnhealthyAfter: 1,
	}
	mon := health.NewMonitor(cfg, be)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mon.Run(ctx)

	waitHealthy(t, be, true, 2*time.Second)

	_ = ln.Close()
	waitHealthy(t, be, false, 2*time.Second)

	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln2.Close()
	go func() {
		for {
			c, err := ln2.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	// New monitor on a fresh address simulates restore after recovery.
	be2 := backend.New("t2", ln2.Addr().String())
	be2.SetHealthy(false)
	mon2 := health.NewMonitor(cfg, be2)
	go mon2.Run(ctx)
	waitHealthy(t, be2, true, 2*time.Second)
}

func waitHealthy(t *testing.T, b *backend.Backend, want bool, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if b.IsHealthy() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for healthy=%v (got %v)", want, b.IsHealthy())
}
