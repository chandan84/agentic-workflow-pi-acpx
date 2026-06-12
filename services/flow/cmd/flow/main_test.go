package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/health"
)

func TestHealthzSmoke(t *testing.T) {
	h := health.New()
	srv := httptest.NewServer(h.Mux())
	defer srv.Close()
	r, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 200 {
		t.Fatalf("status: %d", r.StatusCode)
	}
}
