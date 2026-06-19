// Package app holds the orchestrator's application layer.
package app

import (
	"context"
	"fmt"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/flowsrc"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/workflow"
	"github.com/google/uuid"
)

// Service is the application object.
type Service struct {
	Store      ports.Store
	Bus        events.Bus
	Starter    WorkflowStarter
	Signaler   WorkflowSignaler
	FlowSource flowsrc.FlowSource
}

// WorkflowStarter starts a Temporal workflow.
type WorkflowStarter interface {
	Start(ctx context.Context, runID string, in workflow.RunFlowInput) error
}

// WorkflowSignaler sends a Signal to a running workflow.
type WorkflowSignaler interface {
	Signal(ctx context.Context, runID, signal string, payload any) error
}

// New builds a Service.
func New(s ports.Store, b events.Bus, st WorkflowStarter, sg WorkflowSignaler, fs flowsrc.FlowSource) *Service {
	return &Service{Store: s, Bus: b, Starter: st, Signaler: sg, FlowSource: fs}
}

// StartRun persists a run row, loads the IR from flow-service, and starts the
// Temporal workflow that drives it.
func (s *Service) StartRun(ctx context.Context, flowID, flowVersionID, _input string) (ports.Run, error) {
	if flowID == "" {
		return ports.Run{}, fmt.Errorf("flowId required")
	}
	snap, err := s.FlowSource.GetFlowVersion(ctx, flowID, flowVersionID)
	if err != nil {
		return ports.Run{}, fmt.Errorf("flowsrc: %w", err)
	}
	r := ports.Run{
		ID: uuid.NewString(), FlowID: flowID, FlowVersionID: snap.VersionID,
		Status: "pending", CurrentNode: snap.IR.Start,
	}
	r, err = s.Store.CreateRun(ctx, r)
	if err != nil {
		return r, err
	}
	if s.Starter != nil {
		if err := s.Starter.Start(ctx, r.ID, workflow.RunFlowInput{
			RunID: r.ID, FlowID: flowID, FlowVersionID: snap.VersionID, IR: snap.IR,
		}); err != nil {
			_ = s.Store.UpdateRun(ctx, r.ID, "failed", "")
			return r, err
		}
		_ = s.Store.UpdateRun(ctx, r.ID, "running", snap.IR.Start)
		r.Status = "running"
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectRunStarted, r)
	}
	return r, nil
}

// ApproveCheckpoint signals the workflow and updates the checkpoint row.
func (s *Service) ApproveCheckpoint(ctx context.Context, checkpointID, approver, note string) (ports.CheckpointRequest, error) {
	return s.resolveCheckpoint(ctx, checkpointID, approver, note, "approved", workflow.SignalCheckpointApproved)
}

// RejectCheckpoint signals the workflow with rejection.
func (s *Service) RejectCheckpoint(ctx context.Context, checkpointID, approver, note string) (ports.CheckpointRequest, error) {
	return s.resolveCheckpoint(ctx, checkpointID, approver, note, "rejected", workflow.SignalCheckpointRejected)
}

func (s *Service) resolveCheckpoint(ctx context.Context, checkpointID, approver, note, status, signal string) (ports.CheckpointRequest, error) {
	c, err := s.Store.GetCheckpoint(ctx, checkpointID)
	if err != nil {
		return c, err
	}
	c.Status = status
	c.Approver = approver
	c.Note = note
	c, err = s.Store.UpsertCheckpoint(ctx, c)
	if err != nil {
		return c, err
	}
	if s.Signaler != nil {
		_ = s.Signaler.Signal(ctx, c.RunID, signal, workflow.CheckpointDecision{
			CheckpointID: checkpointID, Approver: approver, Note: note,
		})
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectRunCheckpointDone, c)
	}
	return c, nil
}
