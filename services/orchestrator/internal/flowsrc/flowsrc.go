// Package flowsrc is the FlowSource port + gRPC-client adapter.
//
// The orchestrator needs the Flow IR before starting a Temporal workflow
// (segmenter walks the IR). It fetches the IR from the flow-service via the
// generated gRPC client. A Stub implementation lets the orchestrator boot
// and serve in stub mode without the upstream service.
package flowsrc

import (
	"context"
	"errors"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"google.golang.org/grpc"
)

// FlowSource is the orchestrator's outbound port for fetching a specific
// version's IR + generated .flow.ts path.
type FlowSource interface {
	GetFlowVersion(ctx context.Context, flowID, versionID string) (FlowSnapshot, error)
}

// FlowSnapshot is the IR + acpx .flow.ts source for a specific version.
type FlowSnapshot struct {
	FlowID    string
	VersionID string
	IR        flowir.IR
	IRJSON    string
	FlowTS    string
}

// GRPCClient calls the upstream flow-service.
type GRPCClient struct {
	C flowsv1.FlowServiceClient
}

// NewGRPCClient wraps a gRPC ClientConn.
func NewGRPCClient(cc *grpc.ClientConn) *GRPCClient {
	return &GRPCClient{C: flowsv1.NewFlowServiceClient(cc)}
}

// GetFlowVersion fetches a flow's versions and picks the requested one.
// If versionID is empty, the latest version is returned.
func (g *GRPCClient) GetFlowVersion(ctx context.Context, flowID, versionID string) (FlowSnapshot, error) {
	res, err := g.C.GetFlow(ctx, &flowsv1.GetFlowRequest{Id: flowID})
	if err != nil {
		return FlowSnapshot{}, err
	}
	if len(res.GetVersions()) == 0 {
		return FlowSnapshot{}, errors.New("flow has no versions")
	}
	v := res.GetVersions()[len(res.GetVersions())-1]
	if versionID != "" {
		for _, candidate := range res.GetVersions() {
			if candidate.GetId() == versionID {
				v = candidate
				break
			}
		}
	}
	ir, err := flowir.Parse([]byte(v.GetIrJson()))
	if err != nil {
		return FlowSnapshot{}, err
	}
	return FlowSnapshot{
		FlowID: flowID, VersionID: v.GetId(),
		IR: *ir, IRJSON: v.GetIrJson(), FlowTS: v.GetFlowTs(),
	}, nil
}

// Stub returns a canned single-segment IR for dev/test boot.
type Stub struct {
	Snapshot FlowSnapshot
}

// NewStub returns a stub holding the minimal IR (start → end).
func NewStub() *Stub {
	ir := flowir.IR{
		SchemaVersion: 1, ID: "stub", Name: "stub-flow",
		Start: "start", Ends: []string{"end"},
		Nodes: []flowir.Node{
			{ID: "start", Kind: flowir.NodeAction, Title: "start"},
			{ID: "end", Kind: flowir.NodeEnd, Title: "end"},
		},
		Edges: []flowir.Edge{{From: "start", To: "end"}},
	}
	return &Stub{Snapshot: FlowSnapshot{IR: ir, IRJSON: `{"schemaVersion":1,"id":"stub","name":"stub-flow","start":"start","ends":["end"],"nodes":[{"id":"start","kind":"action"},{"id":"end","kind":"end"}],"edges":[{"from":"start","to":"end"}]}`}}
}

// GetFlowVersion returns the canned snapshot.
func (s *Stub) GetFlowVersion(_ context.Context, _, _ string) (FlowSnapshot, error) {
	return s.Snapshot, nil
}
