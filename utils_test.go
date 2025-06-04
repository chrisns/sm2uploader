package main

import (
	"errors"
	"testing"
)

// errReader is an io.Reader that returns an error after the first Read.
type errReader struct {
	data []byte
	read bool
	err  error
}

func (r *errReader) Read(p []byte) (int, error) {
	if !r.read {
		copy(p, r.data)
		r.read = true
		return len(r.data), nil
	}
	return 0, r.err
}

func TestPostProcessPropagatesReadError(t *testing.T) {
	boom := errors.New("boom")
	r := &errReader{data: []byte("G0 X0\n"), err: boom}
	_, err := postProcess(r)
	if err != boom {
		t.Fatalf("expected %v, got %v", boom, err)
	}
}
