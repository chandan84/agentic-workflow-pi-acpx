// Package grpcsrv adapts the flow application to the generated gRPC server.
package grpcsrv

import (
	"context"

	commonv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/common/v1"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/ports"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements flows.v1.FlowServiceServer.
type Server struct {
	flowsv1.UnimplementedFlowServiceServer
	Svc *app.Service
}

// New wires a Server.
func New(s *app.Service) *Server { return &Server{Svc: s} }

// CreateFlow creates a flow.
func (s *Server) CreateFlow(ctx context.Context, in *flowsv1.CreateFlowRequest) (*flowsv1.CreateFlowResponse, error) {
	f, err := s.Svc.CreateFlow(ctx, in.GetName(), in.GetDescription())
	if err != nil {
		return nil, err
	}
	return &flowsv1.CreateFlowResponse{Flow: toProtoFlow(f)}, nil
}

// ListFlows lists flows.
func (s *Server) ListFlows(ctx context.Context, in *flowsv1.ListFlowsRequest) (*flowsv1.ListFlowsResponse, error) {
	rows, next, err := s.Svc.Store.ListFlows(ctx, pageLimit(in.GetPage()), pageCursor(in.GetPage()))
	if err != nil {
		return nil, err
	}
	out := &flowsv1.ListFlowsResponse{PageInfo: &commonv1.PageInfo{NextCursor: next}}
	for _, f := range rows {
		out.Flows = append(out.Flows, toProtoFlow(f))
	}
	return out, nil
}

// GetFlow returns a flow with its versions.
func (s *Server) GetFlow(ctx context.Context, in *flowsv1.GetFlowRequest) (*flowsv1.GetFlowResponse, error) {
	f, vs, err := s.Svc.Store.GetFlow(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	out := &flowsv1.GetFlowResponse{Flow: toProtoFlow(f)}
	for _, v := range vs {
		out.Versions = append(out.Versions, toProtoVersion(v))
	}
	return out, nil
}

// PutVersion stores a new IR version.
func (s *Server) PutVersion(ctx context.Context, in *flowsv1.PutVersionRequest) (*flowsv1.PutVersionResponse, error) {
	v, err := s.Svc.PutVersion(ctx, in.GetFlowId(), in.GetIrJson())
	if err != nil {
		return nil, err
	}
	return &flowsv1.PutVersionResponse{Version: toProtoVersion(v)}, nil
}

// Validate validates IR JSON.
func (s *Server) Validate(ctx context.Context, in *flowsv1.ValidateRequest) (*flowsv1.ValidateResponse, error) {
	return &flowsv1.ValidateResponse{Errors: s.Svc.Validate(ctx, in.GetIrJson())}, nil
}

// Codegen returns the acpx TypeScript for an IR.
func (s *Server) Codegen(ctx context.Context, in *flowsv1.CodegenRequest) (*flowsv1.CodegenResponse, error) {
	ts, err := s.Svc.Codegen(ctx, in.GetIrJson())
	if err != nil {
		return nil, err
	}
	return &flowsv1.CodegenResponse{FlowTs: ts}, nil
}

// Generate streams the flow-author run.
func (s *Server) Generate(in *flowsv1.GenerateRequest, stream flowsv1.FlowService_GenerateServer) error {
	v, err := s.Svc.Generate(stream.Context(), in.GetFlowId(), in.GetAuthorAgentId(), in.GetPrompt(),
		func(line string) { _ = stream.Send(&flowsv1.GenerateChunk{Body: &flowsv1.GenerateChunk_LogLine{LogLine: line}}) })
	if err != nil {
		return stream.Send(&flowsv1.GenerateChunk{Body: &flowsv1.GenerateChunk_Error{Error: &commonv1.Error{Message: err.Error()}}})
	}
	return stream.Send(&flowsv1.GenerateChunk{Body: &flowsv1.GenerateChunk_Completed{Completed: toProtoVersion(v)}})
}

func toProtoFlow(f ports.Flow) *flowsv1.Flow {
	return &flowsv1.Flow{Id: f.ID, Name: f.Name, Description: f.Description, CreatedAt: timestamppb.New(f.CreatedAt)}
}

func toProtoVersion(v ports.FlowVersion) *flowsv1.FlowVersion {
	return &flowsv1.FlowVersion{
		Id: v.ID, FlowId: v.FlowID, Version: int32(v.Version),
		IrJson: v.IRJSON, FlowTs: v.FlowTS, CreatedAt: timestamppb.New(v.CreatedAt),
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
