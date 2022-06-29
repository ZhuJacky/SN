package myconn

import (
	"context"
	"testing"
	"time"
)

const (
	network     = "tcp"
	validaddr   = "deepzz.com:443"
	invalidaddr = "haha.deepzz.com:443"
)

func TestDialValid(t *testing.T) {
	t.Log(time.Now())

	conn, err := NewWithContext(context.Background(), network, validaddr)
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
}

func TestDialInValid(t *testing.T) {
	t.Log(time.Now())

	conn, err := New("tcp", invalidaddr)
	if err == nil {
		defer conn.Close()
		t.Fatal(err)
	}
}

func BenchmarkDial(b *testing.B) {
	b.N = 1000
	for i := 0; i < b.N; i++ {
		conn, err := New(network, validaddr)
		if err != nil {
			b.Fatal(i, err)
		}
		defer conn.Close()
	}
}
