// Package backend represents one upstream TCP server in a pool.
package backend

// Backend is a single upstream identified by address.
type Backend struct {
	Name    string
	Address string // host:port
}

// New creates a Backend. Name defaults to address when empty.
func New(name, address string) *Backend {
	if name == "" {
		name = address
	}
	return &Backend{Name: name, Address: address}
}
