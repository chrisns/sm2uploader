package main

import "testing"

func TestHumanReadableSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{512, "512 B"},
		{1024 * 1024, "1.0 MB"},
	}
	for _, tt := range tests {
		got := humanReadableSize(tt.size)
		if got != tt.expected {
			t.Errorf("humanReadableSize(%d) = %q, want %q", tt.size, got, tt.expected)
		}
	}
}
