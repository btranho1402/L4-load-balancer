// Command lb is a Stage 3 L4 load balancer: virtual endpoint, backend pool,
// round-robin, and active TCP health monitors.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
	"github.com/btranho1402/l4-load-balancer/internal/balancer"
	"github.com/btranho1402/l4-load-balancer/internal/health"
	"github.com/btranho1402/l4-load-balancer/internal/pool"
	"github.com/btranho1402/l4-load-balancer/internal/server"
)

// stringList collects repeated -backend flags and comma-separated values.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(v string) error {
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*s = append(*s, part)
		}
	}
	return nil
}

func main() {
	listen := flag.String("listen", "127.0.0.1:8080", "virtual endpoint listen address")
	timeout := flag.Duration("dial-timeout", 5*time.Second, "timeout dialing a backend")
	hInterval := flag.Duration("health-interval", time.Second, "active health probe interval")
	hTimeout := flag.Duration("health-timeout", 500*time.Millisecond, "health probe dial timeout")
	hDown := flag.Int("health-unhealthy-after", 2, "consecutive probe failures before removing from rotation")
	hUp := flag.Int("health-healthy-after", 2, "consecutive probe successes before restoring")
	var backends stringList
	flag.Var(&backends, "backend", "backend host:port (repeatable or comma-separated)")
	flag.Parse()

	if len(backends) == 0 {
		fmt.Fprintln(os.Stderr, "at least one -backend is required")
		flag.Usage()
		os.Exit(2)
	}

	bes := make([]*backend.Backend, 0, len(backends))
	for i, addr := range backends {
		bes = append(bes, backend.New(fmt.Sprintf("backend-%d", i+1), addr))
	}

	p := pool.New("default", balancer.NewRoundRobin(), bes...)
	ln := server.NewListener(server.Endpoint{
		Name:        "app",
		Listen:      *listen,
		Pool:        p,
		DialTimeout: *timeout,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mon := health.NewMonitor(health.Config{
		Interval:       *hInterval,
		Timeout:        *hTimeout,
		HealthyAfter:   *hUp,
		UnhealthyAfter: *hDown,
	}, bes...)
	go mon.Run(ctx)

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		log.Printf("shutting down")
		cancel()
		_ = ln.Close()
	}()

	log.Printf("stage3 lb starting: listen=%s backends=%v algorithm=round_robin health=%s",
		*listen, []string(backends), *hInterval)
	if err := ln.ListenAndServe(); err != nil {
		log.Printf("listener stopped: %v", err)
	}
}
