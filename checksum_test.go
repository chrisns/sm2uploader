package main

import "testing"

func TestHeadChksum(t *testing.T) {
	tests := []struct {
		data []byte
		want byte
	}{
		{[]byte{0xaa, 0x55, 0x00, 0x00, 0x01, 0x01}, 0xcf},
		{[]byte{0xaa, 0x55, 0x0c, 0x00, 0x01, 0x02}, 0x2e},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, 0x2f},
	}
	var s SACP_pack
	for _, tt := range tests {
		if got := s.headChksum(tt.data); got != tt.want {
			t.Errorf("headChksum(%v) = 0x%02x, want 0x%02x", tt.data, got, tt.want)
		}
	}
}

func TestU16Chksum(t *testing.T) {
	tests := []struct {
		data   []byte
		length uint16
		want   uint16
	}{
		{[]byte{}, 0, 0xffff},
		{[]byte{1, 2, 3, 4}, 4, 0xfbf9},
		{[]byte{1, 2, 3}, 3, 0xfefa},
		{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 10, 0xebe6},
	}
	var s SACP_pack
	for _, tt := range tests {
		if got := s.U16Chksum(tt.data, tt.length); got != tt.want {
			t.Errorf("U16Chksum(%v, %d) = 0x%04x, want 0x%04x", tt.data, tt.length, got, tt.want)
		}
	}
}
