package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	s := New()
	srv := httptest.NewServer(s.Mux())
	defer srv.Close()
	r, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 200 {
		t.Fatalf("status: %d", r.StatusCode)
	}
}

func TestReadyzFailing(t *testing.T) {
	s := New()
	s.Register("db", func(context.Context) error { return errors.New("down") })
	srv := httptest.NewServer(s.Mux())
	defer srv.Close()
	r, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 503 {
		t.Fatalf("status: %d", r.StatusCode)
	}
}

func TestReadyzOK(t *testing.T) {
	s := New()
	s.Register("db", func(context.Context) error { return nil })
	srv := httptest.NewServer(s.Mux())
	defer srv.Close()
	r, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 200 {
		t.Fatalf("status: %d", r.StatusCode)
	}
}
