package main

import (
	"net"
	"testing"
)

func TestBroadcastForIPNetNonClassful(t *testing.T) {
	_, n, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		t.Fatalf("failed to parse cidr: %v", err)
	}
	got := broadcastForIPNet(n)
	if got == nil {
		t.Fatalf("got nil broadcast address")
	}
	if got.String() != "10.0.0.255" {
		t.Fatalf("expected 10.0.0.255, got %s", got)
	}
}
