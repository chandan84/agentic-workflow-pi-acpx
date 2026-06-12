// Package stub is the in-memory Store used in dev/tests.
package stub

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/ports"
)

// Store keeps flows + versions in memory.
type Store struct {
	mu       sync.RWMutex
	flows    map[string]ports.Flow
	versions map[string][]ports.FlowVersion
}

// New returns a fresh store.
func New() *Store {
	return &Store{flows: map[string]ports.Flow{}, versions: map[string][]ports.FlowVersion{}}
}

// CreateFlow persists a flow.
func (s *Store) CreateFlow(_ context.Context, f ports.Flow) (ports.Flow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f.CreatedAt = time.Now()
	s.flows[f.ID] = f
	return f, nil
}

// GetFlow returns a flow and its versions sorted ascending.
func (s *Store) GetFlow(_ context.Context, id string) (ports.Flow, []ports.FlowVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f := s.flows[id]
	vs := append([]ports.FlowVersion{}, s.versions[id]...)
	sort.Slice(vs, func(i, j int) bool { return vs[i].Version < vs[j].Version })
	return f, vs, nil
}

// ListFlows returns all flows sorted by name.
func (s *Store) ListFlows(_ context.Context, _ int, _ string) ([]ports.Flow, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Flow, 0, len(s.flows))
	for _, f := range s.flows {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, "", nil
}

// PutVersion appends a new version with auto-incremented number.
func (s *Store) PutVersion(_ context.Context, v ports.FlowVersion) (ports.FlowVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v.Version = len(s.versions[v.FlowID]) + 1
	v.CreatedAt = time.Now()
	s.versions[v.FlowID] = append(s.versions[v.FlowID], v)
	return v, nil
}

// Ping always returns nil.
func (s *Store) Ping(_ context.Context) error { return nil }
