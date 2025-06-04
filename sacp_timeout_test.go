package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

// Test that SACP_send_command aborts after the timeout when the connection does not reply.
func TestSACPSendCommandTimeout(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	// Consume all writes and never reply
	go io.Copy(io.Discard, c2)

	start := time.Now()
	err := SACP_send_command(c1, 0x10, 0x02, bytes.Buffer{}, 200*time.Millisecond)
	elapsed := time.Since(start)

	if !errors.Is(err, errTimeoutExceeded) {
		t.Fatalf("expected timeout error, got %v", err)
	}

	if elapsed < 200*time.Millisecond {
		t.Fatalf("expected duration >=200ms, got %v", elapsed)
	}
}
