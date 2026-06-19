// Package testboot exposes a tiny Start helper that boots the agents
// service backed by its in-memory stub store. Integration tests in the
// /tests/e2e module use it because Go's internal-package rule blocks
// importing the real internal packages from outside this module.
package testboot

import (
	"net"
	"testing"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/stub"
	"google.golang.org/grpc"
)

// Start boots agents-service with the stub Store on a random local port and
// returns its listening address along with a stop func.
func Start(t *testing.T, bus events.Bus) (addr string, stop func()) {
	t.Helper()
	st := stub.New()
	svc := app.New(st, bus)
	gs := grpc.NewServer()
	agentsv1.RegisterAgentsServiceServer(gs, grpcsrv.New(svc))
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("agents listen: %v", err)
	}
	go func() { _ = gs.Serve(lis) }()
	return lis.Addr().String(), func() { gs.GracefulStop() }
}
