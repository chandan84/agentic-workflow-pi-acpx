// Package temporalio wires the orchestrator app to a real Temporal client.
//
// Two implementations satisfy the app's WorkflowStarter/Signaler interfaces:
//   - Real: backed by go.temporal.io/sdk/client
//   - Stub: no-op, used when temporal:stub mode is set
package temporalio

import (
	"context"
	"sync"

	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/workflow"
	"go.temporal.io/sdk/client"
)

// Real is the production starter/signaler.
type Real struct {
	C         client.Client
	TaskQueue string
}

// New connects a Temporal client.
func New(hostPort, namespace, taskQueue string) (*Real, error) {
	c, err := client.Dial(client.Options{HostPort: hostPort, Namespace: namespace})
	if err != nil {
		return nil, err
	}
	return &Real{C: c, TaskQueue: taskQueue}, nil
}

// Start kicks off the RunFlow workflow.
func (r *Real) Start(ctx context.Context, runID string, in workflow.RunFlowInput) error {
	_, err := r.C.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: "run-" + runID, TaskQueue: r.TaskQueue},
		workflow.RunFlow, in)
	return err
}

// Signal forwards a Signal to a running workflow.
func (r *Real) Signal(ctx context.Context, runID, name string, payload any) error {
	return r.C.SignalWorkflow(ctx, "run-"+runID, "", name, payload)
}

// Close releases the client.
func (r *Real) Close() { r.C.Close() }

// Stub records calls in memory for tests.
type Stub struct {
	mu      sync.Mutex
	Started []string
	Signals []SignalRecord
}

// SignalRecord captures a stub signal.
type SignalRecord struct {
	RunID, Name string
	Payload     any
}

// NewStub returns a stub starter/signaler.
func NewStub() *Stub { return &Stub{} }

// Start records the run.
func (s *Stub) Start(_ context.Context, runID string, _ workflow.RunFlowInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Started = append(s.Started, runID)
	return nil
}

// Signal records the signal.
func (s *Stub) Signal(_ context.Context, runID, name string, payload any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Signals = append(s.Signals, SignalRecord{RunID: runID, Name: name, Payload: payload})
	return nil
}
