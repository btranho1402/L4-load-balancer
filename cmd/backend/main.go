// Command backend is a tiny TCP echo server used to exercise the Stage 1 proxy.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9001", "listen address")
	name := flag.String("name", "backend-1", "identity returned to clients")
	mode := flag.String("mode", "line", "line (echo) or http")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s listening on %s (%s)", *name, ln.Addr(), *mode)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(conn, *name, *mode)
	}
}

func handle(c net.Conn, name, mode string) {
	defer c.Close()
	if mode == "http" {
		serveHTTP(c, name)
		return
	}

	r := bufio.NewReader(c)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		if _, err := io.WriteString(c, fmt.Sprintf("%s: %s", name, line)); err != nil {
			return
		}
	}
}

func serveHTTP(c net.Conn, name string) {
	r := bufio.NewReader(c)
	_, _ = r.ReadString('\n')
	for {
		line, err := r.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" {
			break
		}
	}

	body := name + "\n"
	resp := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s",
		len(body), body,
	)
	_, _ = io.WriteString(c, resp)
}
