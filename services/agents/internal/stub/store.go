// Package stub is the in-memory Store used in dev/tests.
package stub

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/ports"
)

// Store keeps everything in memory.
type Store struct {
	mu        sync.RWMutex
	groups    map[string]ports.Group
	agents    map[string]ports.Agent
	resources map[string]ports.Resource
	links     []ports.AgentResource
}

// New returns a fresh in-memory store.
func New() *Store {
	return &Store{
		groups:    map[string]ports.Group{},
		agents:    map[string]ports.Agent{},
		resources: map[string]ports.Resource{},
	}
}

// CreateGroup persists a group.
func (s *Store) CreateGroup(_ context.Context, g ports.Group) (ports.Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g.CreatedAt = time.Now()
	s.groups[g.ID] = g
	return g, nil
}

// ListGroups returns all groups (cursor pagination is no-op in stub).
func (s *Store) ListGroups(_ context.Context, _ int, _ string) ([]ports.Group, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Group, 0, len(s.groups))
	for _, g := range s.groups {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, "", nil
}

// CreateAgent persists an agent.
func (s *Store) CreateAgent(_ context.Context, a ports.Agent) (ports.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.CreatedAt = time.Now()
	s.agents[a.ID] = a
	return a, nil
}

// GetAgent fetches an agent by id.
func (s *Store) GetAgent(_ context.Context, id string) (ports.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agents[id], nil
}

// ListAgents returns agents in a group.
func (s *Store) ListAgents(_ context.Context, group string, _ int, _ string) ([]ports.Agent, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ports.Agent
	for _, a := range s.agents {
		if group == "" || a.GroupID == group {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, "", nil
}

// CreateResource persists a resource bundle.
func (s *Store) CreateResource(_ context.Context, r ports.Resource) (ports.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.ID] = r
	return r, nil
}

// ListResources returns all resources.
func (s *Store) ListResources(_ context.Context, _ int, _ string) ([]ports.Resource, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Resource, 0, len(s.resources))
	for _, r := range s.resources {
		out = append(out, r)
	}
	return out, "", nil
}

// AttachResource links an agent to a resource bundle.
func (s *Store) AttachResource(_ context.Context, l ports.AgentResource) (ports.AgentResource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links = append(s.links, l)
	return l, nil
}

// Ping always succeeds.
func (s *Store) Ping(_ context.Context) error { return nil }
