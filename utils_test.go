package main

import (
	"errors"
	"testing"
)

type errReader struct{ err error }

func (e errReader) Read(p []byte) (int, error) { return 0, e.err }

func TestPostProcessPropagatesScannerError(t *testing.T) {
	testErr := errors.New("boom")
	_, err := postProcess(errReader{err: testErr})
	if err != testErr {
		t.Fatalf("expected %v, got %v", testErr, err)
	}
}
