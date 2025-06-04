package main

import (
	"reflect"
	"testing"
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

func TestSACPEncodeDecodeRoundTrip(t *testing.T) {
	want := SACP_pack{
		ReceiverID: 3,
		SenderID:   4,
		Attribute:  1,
		Sequence:   0xBEEF,
		CommandSet: 0x9A,
		CommandID:  0xBC,
		Data:       []byte{0xde, 0xad, 0xbe, 0xef},
	}

	encoded := want.Encode()

	var got SACP_pack
	if err := got.Decode(encoded); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, want)
	}
}
