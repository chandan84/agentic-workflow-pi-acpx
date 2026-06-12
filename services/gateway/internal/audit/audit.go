// Package audit implements the audit/replay aggregator.
//
// The gateway exposes a single audit timeline that merges three sources:
//
//	temporal: Temporal workflow history (server-streamed from orchestrator)
//	acpx:     trace bundle events captured by the AgentRuntime supervisor
//	app:      domain events written to audit.audit_event
//
// This package owns the unified Event shape and the streaming aggregator.
// Implementations connect to NATS + the audit Postgres schema; the in-memory
// implementation here lets the gateway boot and serve in stub mode.
package audit

import (
	"context"
	"sort"
	"sync"
	"time"
)

// Event is the unified audit row.
type Event struct {
	ID, RunID, NodeID, Layer, Kind, PayloadJSON string
	At                                          time.Time
}

// Source emits historical and live audit events.
type Source interface {
	List(ctx context.Context, runID string, limit int) ([]Event, error)
	Subscribe(ctx context.Context, runID string, handler func(Event)) (cancel func(), err error)
}

// Memory is the in-memory Source used in stub mode.
type Memory struct {
	mu     sync.RWMutex
	events []Event
	subs   map[string][]func(Event)
}

// New returns a fresh in-memory source.
func New() *Memory { return &Memory{subs: map[string][]func(Event){}} }

// Add appends an event and notifies subscribers.
func (m *Memory) Add(ev Event) {
	m.mu.Lock()
	m.events = append(m.events, ev)
	handlers := append([]func(Event){}, m.subs[ev.RunID]...)
	m.mu.Unlock()
	for _, h := range handlers {
		h(ev)
	}
}

// List returns events for a run sorted by time.
func (m *Memory) List(_ context.Context, runID string, limit int) ([]Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Event
	for _, e := range m.events {
		if e.RunID == runID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// Subscribe attaches a handler for live events on a run.
func (m *Memory) Subscribe(_ context.Context, runID string, handler func(Event)) (func(), error) {
	m.mu.Lock()
	m.subs[runID] = append(m.subs[runID], handler)
	idx := len(m.subs[runID]) - 1
	m.mu.Unlock()
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if list, ok := m.subs[runID]; ok && idx < len(list) {
			list[idx] = nil
		}
	}, nil
}
