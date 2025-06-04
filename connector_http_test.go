package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gosuri/uilive"
)

func TestHTTPConnectorUpload(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:"+HTTPPort)
	if err != nil {
		t.Skipf("unable to listen on port %s: %v", HTTPPort, err)
	}
	defer l.Close()

	var gotFileName string
	var gotToken string
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/upload":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Errorf("parse form: %v", err)
				return
			}
			gotToken = r.FormValue("token")
			f, fh, err := r.FormFile("file")
			if err != nil {
				t.Errorf("form file: %v", err)
				return
			}
			gotFileName = fh.Filename
			io.Copy(io.Discard, f)
			w.WriteHeader(http.StatusOK)
		case "/api/v1/status":
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	server.Listener = l
	server.Start()
	defer server.Close()

	large := bytes.Repeat([]byte("a"), 1024*50)
	payload := NewPayload(bytes.NewBuffer(large), "code.gcode", int64(len(large)), false)
	NoFix = true

	hc := &HTTPConnector{printer: &Printer{IP: "127.0.0.1", Token: "secret"}}

	buf := &bytes.Buffer{}
	uilive.Out = buf
	uilive.RefreshInterval = time.Millisecond

	if err := hc.Upload(payload); err != nil {
		t.Fatalf("Upload error: %v", err)
	}

	if gotToken != "secret" {
		t.Errorf("token = %q, want %q", gotToken, "secret")
	}
	if gotFileName != "code.gcode" {
		t.Errorf("file name = %q, want %q", gotFileName, "code.gcode")
	}
	if !strings.Contains(buf.String(), "HTTP sending") {
		t.Errorf("progress callback not fired; log: %s", buf.String())
	}
}
