// Package persist adapts the orchestrator's Store to the workflow's
// Persister interface so Activities can write checkpoint state without
// importing the store package.
package persist

import (
	"context"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/workflow"
)

// StoreAdapter implements workflow.Persister against a ports.Store and
// publishes the corresponding events on the bus.
type StoreAdapter struct {
	Store ports.Store
	Bus   events.Bus
}

// New wires an adapter.
func New(s ports.Store, b events.Bus) *StoreAdapter { return &StoreAdapter{Store: s, Bus: b} }

// RecordCheckpointAwaiting upserts an awaiting checkpoint_request row and
// emits SubjectRunCheckpointWait.
func (a *StoreAdapter) RecordCheckpointAwaiting(ctx context.Context, in workflow.RecordCheckpointInput) error {
	c := ports.CheckpointRequest{
		ID: in.CheckpointID, RunID: in.RunID, NodeID: in.NodeID,
		Prompt: in.Prompt, Status: "awaiting",
		Deadline: time.Now().Add(24 * time.Hour),
	}
	c, err := a.Store.UpsertCheckpoint(ctx, c)
	if err != nil {
		return err
	}
	_ = a.Store.UpdateRun(ctx, in.RunID, "awaiting_checkpoint", in.NodeID)
	if a.Bus != nil {
		_ = a.Bus.Publish(ctx, events.SubjectRunCheckpointWait, c)
	}
	return nil
}

// ResolveCheckpoint upserts the resolved status and emits SubjectRunCheckpointDone.
func (a *StoreAdapter) ResolveCheckpoint(ctx context.Context, in workflow.ResolveCheckpointInput) error {
	cur, err := a.Store.GetCheckpoint(ctx, in.CheckpointID)
	if err == nil {
		cur.Status = in.Status
		cur.Approver = in.Approver
		cur.Note = in.Note
		cur, err = a.Store.UpsertCheckpoint(ctx, cur)
		if err != nil {
			return err
		}
		_ = a.Store.UpdateRun(ctx, cur.RunID, "running", "")
		if a.Bus != nil {
			_ = a.Bus.Publish(ctx, events.SubjectRunCheckpointDone, cur)
		}
	}
	return nil
}
