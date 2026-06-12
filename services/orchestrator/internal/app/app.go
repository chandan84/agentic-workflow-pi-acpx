// Package app holds the orchestrator's application layer.
package app

import (
	"context"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/workflow"
	"github.com/google/uuid"
)

// Service is the application object.
type Service struct {
	Store    ports.Store
	Bus      events.Bus
	Starter  WorkflowStarter
	Signaler WorkflowSignaler
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
func New(s ports.Store, b events.Bus, st WorkflowStarter, sg WorkflowSignaler) *Service {
	return &Service{Store: s, Bus: b, Starter: st, Signaler: sg}
}

// StartRun persists a run row and asks Temporal to start the workflow.
func (s *Service) StartRun(ctx context.Context, flowID, flowVersionID, input string) (ports.Run, error) {
	r := ports.Run{
		ID: uuid.NewString(), FlowID: flowID, FlowVersionID: flowVersionID,
		Status: "pending", CurrentNode: "",
	}
	r, err := s.Store.CreateRun(ctx, r)
	if err != nil {
		return r, err
	}
	if s.Starter != nil {
		// Real wiring: load the IR from flow-service and pass it through. The
		// skeleton starts with an empty IR — orchestrator real wiring fills this in.
		if err := s.Starter.Start(ctx, r.ID, workflow.RunFlowInput{
			RunID: r.ID, FlowID: flowID, FlowVersionID: flowVersionID,
		}); err != nil {
			return r, err
		}
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectRunStarted, r)
	}
	return r, nil
}

// ApproveCheckpoint signals the workflow and updates the checkpoint row.
func (s *Service) ApproveCheckpoint(ctx context.Context, checkpointID, approver, note string) (ports.CheckpointRequest, error) {
	c, err := s.Store.GetCheckpoint(ctx, checkpointID)
	if err != nil {
		return c, err
	}
	c.Status = "approved"
	c.Approver = approver
	c.Note = note
	c, err = s.Store.UpsertCheckpoint(ctx, c)
	if err != nil {
		return c, err
	}
	if s.Signaler != nil {
		_ = s.Signaler.Signal(ctx, c.RunID, workflow.SignalCheckpointApproved, workflow.CheckpointDecision{
			CheckpointID: checkpointID, Approver: approver, Note: note,
		})
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectRunCheckpointDone, c)
	}
	return c, nil
}

// RejectCheckpoint signals the workflow with rejection.
func (s *Service) RejectCheckpoint(ctx context.Context, checkpointID, approver, note string) (ports.CheckpointRequest, error) {
	c, err := s.Store.GetCheckpoint(ctx, checkpointID)
	if err != nil {
		return c, err
	}
	c.Status = "rejected"
	c.Approver = approver
	c.Note = note
	c, err = s.Store.UpsertCheckpoint(ctx, c)
	if err != nil {
		return c, err
	}
	if s.Signaler != nil {
		_ = s.Signaler.Signal(ctx, c.RunID, workflow.SignalCheckpointRejected, workflow.CheckpointDecision{
			CheckpointID: checkpointID, Approver: approver, Note: note,
		})
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectRunCheckpointDone, c)
	}
	return c, nil
}
