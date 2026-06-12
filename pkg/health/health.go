// Package health exposes /healthz and /readyz HTTP endpoints.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

// Probe returns nil when the dependency is reachable.
type Probe func(context.Context) error

// Server tracks named readiness probes.
type Server struct {
	mu     sync.RWMutex
	probes map[string]Probe
}

// New returns an empty server.
func New() *Server { return &Server{probes: map[string]Probe{}} }

// Register adds or replaces a probe.
func (s *Server) Register(name string, p Probe) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.probes[name] = p
}

// Mux returns an http.Handler with /healthz and /readyz mounted.
func (s *Server) Mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/readyz", s.readyz)
	return mux
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	probes := make(map[string]Probe, len(s.probes))
	for k, v := range s.probes {
		probes[k] = v
	}
	s.mu.RUnlock()

	results := map[string]string{}
	ok := true
	for name, p := range probes {
		if err := p(r.Context()); err != nil {
			results[name] = err.Error()
			ok = false
		} else {
			results[name] = "ok"
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(results)
}
