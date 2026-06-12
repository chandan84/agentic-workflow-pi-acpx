// Package stub is the in-memory orchestrator Store.
package stub

import (
	"context"
	"sync"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
)

// Store is in-memory.
type Store struct {
	mu          sync.RWMutex
	runs        map[string]ports.Run
	checkpoints map[string]ports.CheckpointRequest
	workItems   []ports.WorkItem
}

// New returns a fresh stub.
func New() *Store {
	return &Store{
		runs:        map[string]ports.Run{},
		checkpoints: map[string]ports.CheckpointRequest{},
	}
}

// CreateRun persists a run.
func (s *Store) CreateRun(_ context.Context, r ports.Run) (ports.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	r.StartedAt = now
	r.UpdatedAt = now
	s.runs[r.ID] = r
	return r, nil
}

// UpdateRun mutates status/current node.
func (s *Store) UpdateRun(_ context.Context, id, status, node string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.runs[id]
	if status != "" {
		r.Status = status
	}
	if node != "" {
		r.CurrentNode = node
	}
	r.UpdatedAt = time.Now()
	s.runs[id] = r
	return nil
}

// GetRun fetches a run.
func (s *Store) GetRun(_ context.Context, id string) (ports.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runs[id], nil
}

// ListRuns returns all runs.
func (s *Store) ListRuns(_ context.Context, flowID string, _ int, _ string) ([]ports.Run, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ports.Run
	for _, r := range s.runs {
		if flowID == "" || r.FlowID == flowID {
			out = append(out, r)
		}
	}
	return out, "", nil
}

// UpsertCheckpoint stores a checkpoint request.
func (s *Store) UpsertCheckpoint(_ context.Context, c ports.CheckpointRequest) (ports.CheckpointRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkpoints[c.ID] = c
	return c, nil
}

// GetCheckpoint reads a checkpoint request.
func (s *Store) GetCheckpoint(_ context.Context, id string) (ports.CheckpointRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.checkpoints[id], nil
}

// InsertWorkItem records a work item.
func (s *Store) InsertWorkItem(_ context.Context, w ports.WorkItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workItems = append(s.workItems, w)
	return nil
}

// Ping always returns nil.
func (s *Store) Ping(_ context.Context) error { return nil }
