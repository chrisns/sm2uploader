package main

import "testing"

func TestShouldBeFix(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"model.gcode", true},
		{"toolpath.nc", false},
		{"firmware.bin", false},
		{"README.txt", false},
	}
	for _, tt := range tests {
		got := shouldBeFix(tt.filename)
		if got != tt.want {
			t.Errorf("shouldBeFix(%q) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}
