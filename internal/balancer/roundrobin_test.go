package balancer_test

import (
	"testing"

	"github.com/btranho1402/l4-load-balancer/internal/backend"
	"github.com/btranho1402/l4-load-balancer/internal/balancer"
)

func TestRoundRobinCycles(t *testing.T) {
	a := backend.New("a", "127.0.0.1:1")
	b := backend.New("b", "127.0.0.1:2")
	c := backend.New("c", "127.0.0.1:3")
	list := []*backend.Backend{a, b, c}

	rr := balancer.NewRoundRobin()
	got := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		be := rr.Next(list)
		if be == nil {
			t.Fatal("expected backend")
		}
		got = append(got, be.Name)
	}

	want := []string{"a", "b", "c", "a", "b", "c"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestRoundRobinEmpty(t *testing.T) {
	if balancer.NewRoundRobin().Next(nil) != nil {
		t.Fatal("expected nil")
	}
}
