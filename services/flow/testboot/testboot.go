// Package testboot exposes a tiny Start helper for integration tests.
package testboot

import (
	"net"
	"testing"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/author"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/stub"
	"google.golang.org/grpc"
)

// Start boots flow-service with the stub store and stub runtime.
func Start(t *testing.T, bus events.Bus) (addr string, stop func()) {
	t.Helper()
	st := stub.New()
	svc := app.New(st, bus, author.New(runtime.NewStub()))
	gs := grpc.NewServer()
	flowsv1.RegisterFlowServiceServer(gs, grpcsrv.New(svc))
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("flow listen: %v", err)
	}
	go func() { _ = gs.Serve(lis) }()
	return lis.Addr().String(), func() { gs.GracefulStop() }
}
