package main

import (
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestLocalStorageLoad(t *testing.T) {
	tmp, err := os.CreateTemp("", "ls*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp error: %v", err)
	}
	defer os.Remove(tmp.Name())

	orig := &LocalStorage{
		Printers: []*Printer{
			{IP: "192.168.1.10", ID: "P1", Model: "M1", Token: "abc", Sacp: true},
			{IP: "192.168.1.11", ID: "P2", Model: "M2", Token: "", Sacp: false},
		},
	}
	b, err := yaml.Marshal(orig)
	if err != nil {
		t.Fatalf("yaml.Marshal error: %v", err)
	}
	if _, err = tmp.Write(b); err != nil {
		t.Fatalf("write yaml error: %v", err)
	}
	tmp.Close()

	loaded := NewLocalStorage(tmp.Name())
	if len(loaded.Printers) != len(orig.Printers) {
		t.Fatalf("expected %d printers, got %d", len(orig.Printers), len(loaded.Printers))
	}
	for i, p := range loaded.Printers {
		o := orig.Printers[i]
		if p.IP != o.IP || p.ID != o.ID || p.Model != o.Model || p.Token != o.Token || p.Sacp != o.Sacp {
			t.Fatalf("printer %d mismatch: %+v vs %+v", i, p, o)
		}
	}
}
