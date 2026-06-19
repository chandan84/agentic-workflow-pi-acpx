// Package testboot exposes a tiny Start helper for integration tests.
package testboot

import (
	"net"
	"testing"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/flowsrc"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/stub"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/temporalio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Start boots orchestrator-service with the stub Store + stub Temporal
// client. It dials flow-service at flowAddr to satisfy the FlowSource port.
func Start(t *testing.T, bus events.Bus, flowAddr string) (addr string, stop func()) {
	t.Helper()
	st := stub.New()
	cc, err := grpc.NewClient(flowAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("orch → flow dial: %v", err)
	}
	fs := flowsrc.NewGRPCClient(cc)
	temp := temporalio.NewStub()
	svc := app.New(st, bus, temp, temp, fs)
	gs := grpc.NewServer()
	executionv1.RegisterExecutionServiceServer(gs, grpcsrv.NewServer(svc, bus))
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("orchestrator listen: %v", err)
	}
	go func() { _ = gs.Serve(lis) }()
	return lis.Addr().String(), func() { gs.GracefulStop(); _ = cc.Close() }
}
