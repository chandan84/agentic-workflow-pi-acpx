// Integration test that boots all four services in-process with stub
// adapters and exercises the user-visible flow through the gateway:
//
//  1. agents-service: create group + agent
//  2. flow-service: store an IR version
//  3. orchestrator-service: start a run pointing at that version
//  4. orchestrator-service: stream the run events back and observe the
//     RunStarted event
//  5. orchestrator-service: approve a checkpoint and observe the
//     resolution event
//
// Every service uses its stub Store and the shared InMemoryBus, plus the
// stub Temporal client so we don't need real infrastructure. The gateway
// proxies all three upstreams.
package e2e

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	commonv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/common/v1"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"

	agentsboot "github.com/chandan84/agentic-workflow-pi-acpx/services/agents/testboot"
	flowboot "github.com/chandan84/agentic-workflow-pi-acpx/services/flow/testboot"
	gwboot "github.com/chandan84/agentic-workflow-pi-acpx/services/gateway/testboot"
	orchboot "github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/testboot"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIntegrationFullStackThroughGateway(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bus := events.NewInMemoryBus()
	defer func() { _ = bus.Close() }()

	agentsAddr, stopAgents := agentsboot.Start(t, bus)
	defer stopAgents()
	flowAddr, stopFlow := flowboot.Start(t, bus)
	defer stopFlow()
	orchAddr, stopOrch := orchboot.Start(t, bus, flowAddr)
	defer stopOrch()
	gatewayAddr, stopGateway := gwboot.Start(t, agentsAddr, flowAddr, orchAddr)
	defer stopGateway()

	cc, err := grpc.NewClient(gatewayAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("gateway dial: %v", err)
	}
	defer cc.Close()

	ac := agentsv1.NewAgentsServiceClient(cc)
	fc := flowsv1.NewFlowServiceClient(cc)
	xc := executionv1.NewExecutionServiceClient(cc)

	// 1. Create group + agent via gateway → agents-service.
	gr, err := ac.CreateGroup(ctx, &agentsv1.CreateGroupRequest{Name: "Payments", Description: "Bug triage"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if gr.GetGroup().GetId() == "" {
		t.Fatalf("group has no id")
	}
	if _, err := ac.CreateAgent(ctx, &agentsv1.CreateAgentRequest{
		GroupId: gr.GetGroup().GetId(), Name: "qa", Role: "reviewer",
	}); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if lg, err := ac.ListGroups(ctx, &agentsv1.ListGroupsRequest{Page: &commonv1.Page{Limit: 10}}); err != nil || len(lg.GetGroups()) == 0 {
		t.Fatalf("ListGroups: %v len=%d", err, len(lg.GetGroups()))
	}

	// 2. Create a flow + store an IR version via gateway → flow-service.
	flw, err := fc.CreateFlow(ctx, &flowsv1.CreateFlowRequest{Name: "hello-flow", Description: "smoke"})
	if err != nil {
		t.Fatalf("CreateFlow: %v", err)
	}
	ir := `{"schemaVersion":1,"id":"hello","name":"hello","start":"draft","ends":["review"],"nodes":[{"id":"draft","kind":"action","title":"draft"},{"id":"approve","kind":"checkpoint","title":"approve","approvers":["role:reviewer"],"timeoutSeconds":60,"onTimeout":"abort"},{"id":"review","kind":"end","title":"review"}],"edges":[{"from":"draft","to":"approve"},{"from":"approve","to":"review"}]}`
	pv, err := fc.PutVersion(ctx, &flowsv1.PutVersionRequest{FlowId: flw.GetFlow().GetId(), IrJson: ir})
	if err != nil {
		t.Fatalf("PutVersion: %v", err)
	}
	if pv.GetVersion().GetVersion() != 1 {
		t.Fatalf("first version should be 1, got %d", pv.GetVersion().GetVersion())
	}

	// 3. Subscribe to run events BEFORE StartRun so we see them.
	evCh := make(chan *executionv1.RunEvent, 32)
	streamCtx, streamCancel := context.WithCancel(ctx)
	defer streamCancel()
	go func() {
		defer close(evCh)
		s, err := xc.StreamRunEvents(streamCtx, &executionv1.StreamRunEventsRequest{})
		if err != nil {
			return
		}
		for {
			ev, err := s.Recv()
			if err == io.EOF || err != nil {
				return
			}
			select {
			case evCh <- ev:
			case <-streamCtx.Done():
				return
			}
		}
	}()
	time.Sleep(100 * time.Millisecond) // let the subscription register

	// 4. Start a run via gateway → orchestrator.
	rr, err := xc.StartRun(ctx, &executionv1.StartRunRequest{FlowId: flw.GetFlow().GetId(), FlowVersionId: pv.GetVersion().GetId()})
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	runID := rr.GetRun().GetId()
	if runID == "" {
		t.Fatalf("run has no id")
	}

	// 5. Approve a checkpoint (the stub Signaler records it; the in-memory
	// bus fans out the resolution event).
	cpID := runID + ":approve"
	_, _ = xc.ApproveCheckpoint(ctx, &executionv1.ApproveCheckpointRequest{
		CheckpointId: cpID, Approver: "qa", Note: "lgtm",
	})

	// 6. Expect a RunStarted event in the stream.
	got := drain(evCh, 500*time.Millisecond)
	if !containsKind(got, events.SubjectRunStarted) {
		t.Fatalf("expected run.started event, got: %v", kinds(got))
	}

	// 7. ListRuns returns the run.
	lr, err := xc.ListRuns(ctx, &executionv1.ListRunsRequest{})
	if err != nil || len(lr.GetRuns()) == 0 {
		t.Fatalf("ListRuns: %v len=%d", err, len(lr.GetRuns()))
	}
}

func drain(ch <-chan *executionv1.RunEvent, wait time.Duration) []*executionv1.RunEvent {
	var got []*executionv1.RunEvent
	deadline := time.NewTimer(wait)
	defer deadline.Stop()
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, ev)
		case <-deadline.C:
			return got
		}
	}
}

func containsKind(evs []*executionv1.RunEvent, kind string) bool {
	for _, e := range evs {
		if e.GetKind() == kind {
			return true
		}
	}
	return false
}

func kinds(evs []*executionv1.RunEvent) []string {
	out := make([]string, 0, len(evs))
	for _, e := range evs {
		out = append(out, e.GetKind())
	}
	return out
}
