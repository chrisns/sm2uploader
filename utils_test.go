package main

import "testing"

func TestNormalizedFilename(t *testing.T) {
	tests := map[string]string{
		"foo.gcode":           "foo.gcode",
		"../foo.gcode":        "foo.gcode",
		"./foo.gcode":         "foo.gcode",
		"~/foo.gcode":         "foo.gcode",
		"dir1/foo.gcode":      "foo.gcode",
		"dir1/dir2/foo.gcode": "foo.gcode",
		"dir1\\foo.gcode":     "foo.gcode",
		"C:\\tmp\\foo.gcode":  "foo.gcode",
	}
	for input, want := range tests {
		got := normalizedFilename(input)
		if got != want {
			t.Errorf("normalizedFilename(%q)=%q, want %q", input, got, want)
		}
	}
}
