package main

import "testing"

func TestLocalStorageAddDuplicateID(t *testing.T) {
	path := t.TempDir() + "/ls.yaml"
	ls := NewLocalStorage(path)

	p1 := &Printer{ID: "Printer1", IP: "192.168.1.1"}
	p2 := &Printer{ID: "Printer1", IP: "192.168.1.2"}

	ls.Add(p1)
	ls.Add(p2)

	if len(ls.Printers) != 1 {
		t.Fatalf("expected 1 printer, got %d", len(ls.Printers))
	}

	if ls.Printers[0].IP != p2.IP {
		t.Fatalf("expected IP %s, got %s", p2.IP, ls.Printers[0].IP)
	}
}
