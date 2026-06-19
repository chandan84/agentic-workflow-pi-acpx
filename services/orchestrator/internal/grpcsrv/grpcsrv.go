// Package grpcsrv adapts the orchestrator application to gRPC.
package grpcsrv

import (
	"context"
	"encoding/json"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	commonv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/common/v1"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements execution.v1.ExecutionServiceServer.
type Server struct {
	executionv1.UnimplementedExecutionServiceServer
	Svc *app.Service
	Bus events.Bus
}

// NewServer wires a Server with a bus subscription for live run events.
func NewServer(s *app.Service, b events.Bus) *Server { return &Server{Svc: s, Bus: b} }

// New wires a Server without bus (compat shim for callers that don't stream).
func New(s *app.Service) *Server { return &Server{Svc: s} }

// StartRun starts a new run.
func (s *Server) StartRun(ctx context.Context, in *executionv1.StartRunRequest) (*executionv1.StartRunResponse, error) {
	r, err := s.Svc.StartRun(ctx, in.GetFlowId(), in.GetFlowVersionId(), in.GetInputJson())
	if err != nil {
		return nil, err
	}
	return &executionv1.StartRunResponse{Run: toProtoRun(r)}, nil
}

// GetRun returns a run.
func (s *Server) GetRun(ctx context.Context, in *executionv1.GetRunRequest) (*executionv1.GetRunResponse, error) {
	r, err := s.Svc.Store.GetRun(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &executionv1.GetRunResponse{Run: toProtoRun(r)}, nil
}

// ListRuns returns runs for a flow.
func (s *Server) ListRuns(ctx context.Context, in *executionv1.ListRunsRequest) (*executionv1.ListRunsResponse, error) {
	rows, next, err := s.Svc.Store.ListRuns(ctx, in.GetFlowId(), pageLimit(in.GetPage()), pageCursor(in.GetPage()))
	if err != nil {
		return nil, err
	}
	out := &executionv1.ListRunsResponse{PageInfo: &commonv1.PageInfo{NextCursor: next}}
	for _, r := range rows {
		out.Runs = append(out.Runs, toProtoRun(r))
	}
	return out, nil
}

// ApproveCheckpoint signals approval.
func (s *Server) ApproveCheckpoint(ctx context.Context, in *executionv1.ApproveCheckpointRequest) (*executionv1.ApproveCheckpointResponse, error) {
	c, err := s.Svc.ApproveCheckpoint(ctx, in.GetCheckpointId(), in.GetApprover(), in.GetNote())
	if err != nil {
		return nil, err
	}
	return &executionv1.ApproveCheckpointResponse{Checkpoint: toProtoCheckpoint(c)}, nil
}

// RejectCheckpoint signals rejection.
func (s *Server) RejectCheckpoint(ctx context.Context, in *executionv1.RejectCheckpointRequest) (*executionv1.RejectCheckpointResponse, error) {
	c, err := s.Svc.RejectCheckpoint(ctx, in.GetCheckpointId(), in.GetApprover(), in.GetNote())
	if err != nil {
		return nil, err
	}
	return &executionv1.RejectCheckpointResponse{Checkpoint: toProtoCheckpoint(c)}, nil
}

// StreamRunEvents subscribes to the bus and forwards run-scoped events to the
// caller. The subscription is cancelled when the client closes the stream.
func (s *Server) StreamRunEvents(in *executionv1.StreamRunEventsRequest, stream executionv1.ExecutionService_StreamRunEventsServer) error {
	if s.Bus == nil {
		return nil
	}
	wanted := in.GetRunId()
	subjects := []string{
		events.SubjectRunStarted, events.SubjectRunUpdated, events.SubjectRunCompleted, events.SubjectRunFailed,
		events.SubjectRunCheckpointWait, events.SubjectRunCheckpointDone,
		events.SubjectRunSegmentStart, events.SubjectRunSegmentEnd, events.SubjectRunWorkItemEmitted,
	}
	cancels := make([]func() error, 0, len(subjects))
	for _, subj := range subjects {
		subject := subj
		cancel, err := s.Bus.Subscribe(stream.Context(), subject, func(_ string, data []byte) {
			if !runIDMatches(wanted, data) {
				return
			}
			_ = stream.Send(&executionv1.RunEvent{
				RunId: extractRunID(data), Kind: subject, PayloadJson: string(data),
				At: timestamppb.Now(),
			})
		})
		if err != nil {
			continue
		}
		cancels = append(cancels, cancel)
	}
	defer func() {
		for _, c := range cancels {
			_ = c()
		}
	}()
	<-stream.Context().Done()
	return nil
}

func runIDMatches(wanted string, data []byte) bool {
	if wanted == "" {
		return true
	}
	return extractRunID(data) == wanted
}

func extractRunID(data []byte) string {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if v, ok := m["runId"].(string); ok && v != "" {
		return v
	}
	if v, ok := m["RunID"].(string); ok && v != "" {
		return v
	}
	if v, ok := m["id"].(string); ok && v != "" {
		return v
	}
	return ""
}

func toProtoRun(r ports.Run) *executionv1.Run {
	return &executionv1.Run{
		Id: r.ID, FlowId: r.FlowID, FlowVersionId: r.FlowVersionID,
		Status: r.Status, CurrentNode: r.CurrentNode,
		StartedAt: timestamppb.New(r.StartedAt), UpdatedAt: timestamppb.New(r.UpdatedAt),
	}
}

func toProtoCheckpoint(c ports.CheckpointRequest) *executionv1.CheckpointRequest {
	return &executionv1.CheckpointRequest{
		Id: c.ID, RunId: c.RunID, NodeId: c.NodeID, Prompt: c.Prompt, Status: c.Status,
		Deadline: timestamppb.New(c.Deadline),
	}
}

func pageLimit(p *commonv1.Page) int {
	if p == nil || p.Limit <= 0 {
		return 100
	}
	return int(p.Limit)
}
func pageCursor(p *commonv1.Page) string {
	if p == nil {
		return ""
	}
	return p.Cursor
}
