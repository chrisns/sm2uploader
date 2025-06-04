package main

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestSACPDecodeValid(t *testing.T) {
	orig := SACP_pack{
		ReceiverID: 1,
		SenderID:   2,
		Attribute:  0,
		Sequence:   0x1234,
		CommandSet: 0x56,
		CommandID:  0x78,
		Data:       []byte{0x9a, 0xbc},
	}

	encoded := orig.Encode()

	var got SACP_pack
	if err := got.Decode(encoded); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if got.ReceiverID != orig.ReceiverID ||
		got.SenderID != orig.SenderID ||
		got.Attribute != orig.Attribute ||
		got.Sequence != orig.Sequence ||
		got.CommandSet != orig.CommandSet ||
		got.CommandID != orig.CommandID ||
		string(got.Data) != string(orig.Data) {
		t.Fatalf("Decoded packet mismatch: %+v vs %+v", got, orig)
	}
}

func TestSACPDecodeInvalidHeader(t *testing.T) {
	p := SACP_pack{ReceiverID: 1, SenderID: 2, Attribute: 0, Sequence: 1}
	encoded := p.Encode()

	encoded[0] ^= 0xff // corrupt first header byte

	var got SACP_pack
	err := got.Decode(encoded)
	if err != errInvalidSACP {
		t.Fatalf("expected errInvalidSACP, got %v", err)
	}
}

func TestSACPSendCommandTimeout(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	go io.Copy(io.Discard, c2)

	err := SACP_send_command(c1, 1, 2, bytes.Buffer{}, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
