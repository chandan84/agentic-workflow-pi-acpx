// Package testboot exposes a tiny Start helper for integration tests.
//
// The integration test in /tests/e2e cannot reach
// services/gateway/internal/upstreams directly because of Go's internal-
// package rule; this package re-exposes the proxy-registration code.
package testboot

import (
	"net"
	"testing"

	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/gateway/internal/upstreams"
	"google.golang.org/grpc"
)

// Start boots gateway-service registering proxies that forward to the
// given upstream addresses. Any upstream may be empty — its handler is
// just not registered.
func Start(t *testing.T, agentsAddr, flowAddr, orchAddr string) (addr string, stop func()) {
	t.Helper()
	gs := grpc.NewServer()
	var closers []func()
	if agentsAddr != "" {
		cc, err := upstreams.Dial(agentsAddr)
		if err != nil {
			t.Fatalf("agents dial: %v", err)
		}
		agentsv1.RegisterAgentsServiceServer(gs, upstreams.NewAgentsProxy(cc))
		closers = append(closers, func() { _ = cc.Close() })
	}
	if flowAddr != "" {
		cc, err := upstreams.Dial(flowAddr)
		if err != nil {
			t.Fatalf("flow dial: %v", err)
		}
		flowsv1.RegisterFlowServiceServer(gs, upstreams.NewFlowsProxy(cc))
		closers = append(closers, func() { _ = cc.Close() })
	}
	if orchAddr != "" {
		cc, err := upstreams.Dial(orchAddr)
		if err != nil {
			t.Fatalf("orchestrator dial: %v", err)
		}
		executionv1.RegisterExecutionServiceServer(gs, upstreams.NewExecutionProxy(cc))
		closers = append(closers, func() { _ = cc.Close() })
	}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("gateway listen: %v", err)
	}
	go func() { _ = gs.Serve(lis) }()
	return lis.Addr().String(), func() {
		gs.GracefulStop()
		for _, c := range closers {
			c()
		}
	}
}
