// Package proxy forwards a client TCP connection to one upstream backend.
package proxy

import (
	"io"
	"net"
	"sync"
	"time"
)

// Forward dials backendAddr, then copies bytes both ways until either side closes.
// Each call is meant to run in its own goroutine (goroutine-per-connection).
func Forward(client net.Conn, backendAddr string, dialTimeout time.Duration) {
	defer client.Close()

	if dialTimeout <= 0 {
		dialTimeout = 5 * time.Second
	}

	d := net.Dialer{Timeout: dialTimeout}
	upstream, err := d.Dial("tcp", backendAddr)
	if err != nil {
		return
	}
	defer upstream.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(upstream, client)
		_ = closeWrite(upstream)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(client, upstream)
		_ = closeWrite(client)
	}()

	wg.Wait()
}

func closeWrite(c net.Conn) error {
	type closeWriter interface {
		CloseWrite() error
	}
	if cw, ok := c.(closeWriter); ok {
		return cw.CloseWrite()
	}
	return nil
}
