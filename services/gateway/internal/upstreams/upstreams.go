// Package upstreams holds gRPC client dials and proxy registration for the
// gateway's aggregated RPCs.
//
// The gateway is the single endpoint the desktop app talks to. It registers
// the same service interfaces it exposes (agents, flows, execution) and
// forwards each call to the appropriate upstream service through a thin
// proxy. Streaming RPCs (Generate, StreamRunEvents) pump bytes both ways
// without translation.
package upstreams

import (
	"context"
	"io"

	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial returns a gRPC ClientConn with insecure transport (dev mode).
func Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// AgentsProxy forwards AgentsService calls to the upstream agents-service.
type AgentsProxy struct {
	agentsv1.UnimplementedAgentsServiceServer
	C agentsv1.AgentsServiceClient
}

// NewAgentsProxy constructs an AgentsProxy from a dialed ClientConn.
func NewAgentsProxy(cc *grpc.ClientConn) *AgentsProxy {
	return &AgentsProxy{C: agentsv1.NewAgentsServiceClient(cc)}
}

// CreateGroup forwards.
func (p *AgentsProxy) CreateGroup(ctx context.Context, in *agentsv1.CreateGroupRequest) (*agentsv1.CreateGroupResponse, error) {
	return p.C.CreateGroup(ctx, in)
}

// ListGroups forwards.
func (p *AgentsProxy) ListGroups(ctx context.Context, in *agentsv1.ListGroupsRequest) (*agentsv1.ListGroupsResponse, error) {
	return p.C.ListGroups(ctx, in)
}

// CreateAgent forwards.
func (p *AgentsProxy) CreateAgent(ctx context.Context, in *agentsv1.CreateAgentRequest) (*agentsv1.CreateAgentResponse, error) {
	return p.C.CreateAgent(ctx, in)
}

// ListAgents forwards.
func (p *AgentsProxy) ListAgents(ctx context.Context, in *agentsv1.ListAgentsRequest) (*agentsv1.ListAgentsResponse, error) {
	return p.C.ListAgents(ctx, in)
}

// GetAgent forwards.
func (p *AgentsProxy) GetAgent(ctx context.Context, in *agentsv1.GetAgentRequest) (*agentsv1.GetAgentResponse, error) {
	return p.C.GetAgent(ctx, in)
}

// CreateResource forwards.
func (p *AgentsProxy) CreateResource(ctx context.Context, in *agentsv1.CreateResourceRequest) (*agentsv1.CreateResourceResponse, error) {
	return p.C.CreateResource(ctx, in)
}

// ListResources forwards.
func (p *AgentsProxy) ListResources(ctx context.Context, in *agentsv1.ListResourcesRequest) (*agentsv1.ListResourcesResponse, error) {
	return p.C.ListResources(ctx, in)
}

// AttachResource forwards.
func (p *AgentsProxy) AttachResource(ctx context.Context, in *agentsv1.AttachResourceRequest) (*agentsv1.AttachResourceResponse, error) {
	return p.C.AttachResource(ctx, in)
}

// FlowsProxy forwards FlowService calls.
type FlowsProxy struct {
	flowsv1.UnimplementedFlowServiceServer
	C flowsv1.FlowServiceClient
}

// NewFlowsProxy constructs a FlowsProxy from a dialed ClientConn.
func NewFlowsProxy(cc *grpc.ClientConn) *FlowsProxy {
	return &FlowsProxy{C: flowsv1.NewFlowServiceClient(cc)}
}

// CreateFlow forwards.
func (p *FlowsProxy) CreateFlow(ctx context.Context, in *flowsv1.CreateFlowRequest) (*flowsv1.CreateFlowResponse, error) {
	return p.C.CreateFlow(ctx, in)
}

// ListFlows forwards.
func (p *FlowsProxy) ListFlows(ctx context.Context, in *flowsv1.ListFlowsRequest) (*flowsv1.ListFlowsResponse, error) {
	return p.C.ListFlows(ctx, in)
}

// GetFlow forwards.
func (p *FlowsProxy) GetFlow(ctx context.Context, in *flowsv1.GetFlowRequest) (*flowsv1.GetFlowResponse, error) {
	return p.C.GetFlow(ctx, in)
}

// PutVersion forwards.
func (p *FlowsProxy) PutVersion(ctx context.Context, in *flowsv1.PutVersionRequest) (*flowsv1.PutVersionResponse, error) {
	return p.C.PutVersion(ctx, in)
}

// Validate forwards.
func (p *FlowsProxy) Validate(ctx context.Context, in *flowsv1.ValidateRequest) (*flowsv1.ValidateResponse, error) {
	return p.C.Validate(ctx, in)
}

// Codegen forwards.
func (p *FlowsProxy) Codegen(ctx context.Context, in *flowsv1.CodegenRequest) (*flowsv1.CodegenResponse, error) {
	return p.C.Codegen(ctx, in)
}

// Generate streams the upstream server stream back to the gateway client.
func (p *FlowsProxy) Generate(in *flowsv1.GenerateRequest, stream flowsv1.FlowService_GenerateServer) error {
	up, err := p.C.Generate(stream.Context(), in)
	if err != nil {
		return err
	}
	for {
		chunk, err := up.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(chunk); err != nil {
			return err
		}
	}
}

// ExecutionProxy forwards ExecutionService calls.
type ExecutionProxy struct {
	executionv1.UnimplementedExecutionServiceServer
	C executionv1.ExecutionServiceClient
}

// NewExecutionProxy constructs an ExecutionProxy from a dialed ClientConn.
func NewExecutionProxy(cc *grpc.ClientConn) *ExecutionProxy {
	return &ExecutionProxy{C: executionv1.NewExecutionServiceClient(cc)}
}

// StartRun forwards.
func (p *ExecutionProxy) StartRun(ctx context.Context, in *executionv1.StartRunRequest) (*executionv1.StartRunResponse, error) {
	return p.C.StartRun(ctx, in)
}

// GetRun forwards.
func (p *ExecutionProxy) GetRun(ctx context.Context, in *executionv1.GetRunRequest) (*executionv1.GetRunResponse, error) {
	return p.C.GetRun(ctx, in)
}

// ListRuns forwards.
func (p *ExecutionProxy) ListRuns(ctx context.Context, in *executionv1.ListRunsRequest) (*executionv1.ListRunsResponse, error) {
	return p.C.ListRuns(ctx, in)
}

// ApproveCheckpoint forwards.
func (p *ExecutionProxy) ApproveCheckpoint(ctx context.Context, in *executionv1.ApproveCheckpointRequest) (*executionv1.ApproveCheckpointResponse, error) {
	return p.C.ApproveCheckpoint(ctx, in)
}

// RejectCheckpoint forwards.
func (p *ExecutionProxy) RejectCheckpoint(ctx context.Context, in *executionv1.RejectCheckpointRequest) (*executionv1.RejectCheckpointResponse, error) {
	return p.C.RejectCheckpoint(ctx, in)
}

// StreamRunEvents pumps the upstream server stream back to the gateway client.
func (p *ExecutionProxy) StreamRunEvents(in *executionv1.StreamRunEventsRequest, stream executionv1.ExecutionService_StreamRunEventsServer) error {
	up, err := p.C.StreamRunEvents(stream.Context(), in)
	if err != nil {
		return err
	}
	for {
		ev, err := up.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(ev); err != nil {
			return err
		}
	}
}
