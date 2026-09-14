// Command lb is a Stage 1 L4 TCP proxy: accept on one address, forward every
// connection to a single backend. Later stages add pools, algorithms, and health.
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btranho1402/l4-load-balancer/internal/proxy"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8080", "address to accept client connections")
	backend := flag.String("backend", "127.0.0.1:9001", "single upstream TCP backend")
	timeout := flag.Duration("dial-timeout", 5*time.Second, "timeout dialing the backend")
	flag.Parse()

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen %s: %v", *listen, err)
	}
	log.Printf("stage1 proxy listening on %s -> %s", *listen, *backend)

	// Graceful stop on Ctrl-C / SIGTERM.
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		log.Printf("shutting down")
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			return
		}
		go proxy.Forward(conn, *backend, *timeout)
	}
}
