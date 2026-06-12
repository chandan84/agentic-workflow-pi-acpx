// Package workflow holds the Temporal workflow and activities that drive an
// acpx flow run.
//
// The workflow segments the IR at `checkpoint` nodes:
//
//	while current != endNode:
//	    segment := findSegmentFrom(current)        // [current, nextCheckpoint or end]
//	    res := ExecuteSegmentActivity(segment)     // shells out to `acpx flow run --from --to`
//	    if next is checkpoint:
//	        RecordCheckpointAwaitingActivity(...)  // creates the checkpoint_request row + event
//	        await Signal("CheckpointApproved" | "CheckpointRejected") with Timer(timeout)
//	        if timeout: handle onTimeout (escalate|abort|auto-approve)
//	    current = res.NextNodeID
//
// Activities use heartbeats for long-running acpx calls. The acpx runtime is
// invoked through the AgentRuntime port; in tests, a Stub returns canned
// segment results.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SignalCheckpointApproved is the Temporal Signal name used to resume a paused run.
const SignalCheckpointApproved = "CheckpointApproved"

// SignalCheckpointRejected aborts a paused run.
const SignalCheckpointRejected = "CheckpointRejected"

// RunFlowInput is the workflow input.
type RunFlowInput struct {
	RunID         string
	FlowID        string
	FlowVersionID string
	IR            flowir.IR
	FlowTSPath    string
	WorkspaceDir  string
}

// RunFlowResult is the workflow result.
type RunFlowResult struct {
	FinalNodeID string
	Outputs     map[string]any
}

// CheckpointDecision is sent via Signal.
type CheckpointDecision struct {
	CheckpointID string
	Approver     string
	Note         string
}

// RunFlow is the workflow.
func RunFlow(ctx workflow.Context, in RunFlowInput) (RunFlowResult, error) {
	log := workflow.GetLogger(ctx)
	log.Info("RunFlow started", "runId", in.RunID, "flowId", in.FlowID)

	nodes := indexNodes(in.IR.Nodes)
	current := in.IR.Start

	for {
		toNodeID, isCheckpoint, isEnd, err := nextSegment(current, in.IR, nodes)
		if err != nil {
			return RunFlowResult{}, err
		}

		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Minute,
			HeartbeatTimeout:    1 * time.Minute,
			RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
		}
		segCtx := workflow.WithActivityOptions(ctx, ao)
		var segRes SegmentResult
		err = workflow.ExecuteActivity(segCtx, ExecuteSegmentActivityName, ExecuteSegmentInput{
			RunID: in.RunID, FromNodeID: current, ToNodeID: toNodeID,
			FlowTSPath: in.FlowTSPath, WorkspaceDir: in.WorkspaceDir,
		}).Get(ctx, &segRes)
		if err != nil {
			return RunFlowResult{}, fmt.Errorf("segment %s→%s: %w", current, toNodeID, err)
		}
		current = segRes.NextNodeID
		if current == "" {
			current = toNodeID
		}

		if isCheckpoint {
			node := nodes[toNodeID]
			cpID := fmt.Sprintf("%s:%s", in.RunID, toNodeID)
			recCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
			if err := workflow.ExecuteActivity(recCtx, RecordCheckpointAwaitingActivityName, RecordCheckpointInput{
				CheckpointID: cpID, RunID: in.RunID, NodeID: toNodeID, Prompt: node.Prompt,
			}).Get(ctx, nil); err != nil {
				return RunFlowResult{}, err
			}

			timeoutSeconds := node.TimeoutSeconds
			if timeoutSeconds <= 0 {
				timeoutSeconds = 86400 // 1 day default
			}
			decision, timedOut := awaitDecision(ctx, time.Duration(timeoutSeconds)*time.Second)
			if timedOut {
				switch node.OnTimeout {
				case "auto-approve":
					decision = CheckpointDecision{CheckpointID: cpID, Approver: "auto", Note: "timeout"}
				case "abort":
					return RunFlowResult{}, fmt.Errorf("checkpoint %s timed out and onTimeout=abort", toNodeID)
				case "escalate":
					return RunFlowResult{}, fmt.Errorf("checkpoint %s timed out — escalation required", toNodeID)
				default:
					return RunFlowResult{}, fmt.Errorf("checkpoint %s timed out", toNodeID)
				}
			}
			resCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
			_ = workflow.ExecuteActivity(resCtx, ResolveCheckpointActivityName, ResolveCheckpointInput{
				CheckpointID: cpID, Approver: decision.Approver, Note: decision.Note, Status: "approved",
			}).Get(ctx, nil)
		}

		if isEnd || current == "" {
			return RunFlowResult{FinalNodeID: toNodeID, Outputs: segRes.Output}, nil
		}
	}
}

func awaitDecision(ctx workflow.Context, timeout time.Duration) (CheckpointDecision, bool) {
	ch := workflow.GetSignalChannel(ctx, SignalCheckpointApproved)
	rej := workflow.GetSignalChannel(ctx, SignalCheckpointRejected)
	timer := workflow.NewTimer(ctx, timeout)
	sel := workflow.NewSelector(ctx)
	var decision CheckpointDecision
	timedOut := false
	sel.AddReceive(ch, func(c workflow.ReceiveChannel, _ bool) { c.Receive(ctx, &decision) })
	sel.AddReceive(rej, func(c workflow.ReceiveChannel, _ bool) {
		var d CheckpointDecision
		c.Receive(ctx, &d)
		_ = workflow.GetLogger(ctx)
	})
	sel.AddFuture(timer, func(workflow.Future) { timedOut = true })
	sel.Select(ctx)
	return decision, timedOut
}

func nextSegment(from string, ir flowir.IR, nodes map[string]flowir.Node) (toID string, isCheckpoint bool, isEnd bool, err error) {
	cur := from
	visited := map[string]bool{}
	for {
		if visited[cur] {
			return "", false, false, errors.New("cycle in flow without checkpoint or end")
		}
		visited[cur] = true
		nexts := outgoing(cur, ir)
		if len(nexts) == 0 {
			if n, ok := nodes[cur]; ok && n.Kind == flowir.NodeEnd {
				return cur, false, true, nil
			}
			return cur, false, true, nil
		}
		next := nexts[0]
		n, ok := nodes[next]
		if !ok {
			return "", false, false, fmt.Errorf("edge to unknown node %s", next)
		}
		if n.Kind == flowir.NodeCheckpoint {
			return next, true, false, nil
		}
		if n.Kind == flowir.NodeEnd {
			return next, false, true, nil
		}
		cur = next
	}
}

func outgoing(from string, ir flowir.IR) []string {
	var out []string
	for _, e := range ir.Edges {
		if e.From == from {
			out = append(out, e.To)
		}
	}
	for _, n := range ir.Nodes {
		if n.ID == from && n.Kind == flowir.NodeDecision {
			for _, b := range n.Branches {
				out = append(out, b.To)
			}
		}
	}
	return out
}

func indexNodes(nodes []flowir.Node) map[string]flowir.Node {
	m := make(map[string]flowir.Node, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n
	}
	return m
}

// Activity names — used both by the worker registration and the workflow body
// (workflows reference activities by name when their inputs are dynamic).
const (
	ExecuteSegmentActivityName           = "ExecuteSegment"
	RecordCheckpointAwaitingActivityName = "RecordCheckpointAwaiting"
	ResolveCheckpointActivityName        = "ResolveCheckpoint"
)

// SegmentResult is returned by ExecuteSegment.
type SegmentResult struct {
	NextNodeID string
	Output     map[string]any
}

// ExecuteSegmentInput is the input to the ExecuteSegment activity.
type ExecuteSegmentInput struct {
	RunID, FromNodeID, ToNodeID, FlowTSPath, WorkspaceDir string
}

// RecordCheckpointInput is the input to the RecordCheckpointAwaiting activity.
type RecordCheckpointInput struct {
	CheckpointID, RunID, NodeID, Prompt string
}

// ResolveCheckpointInput is the input to the ResolveCheckpoint activity.
type ResolveCheckpointInput struct {
	CheckpointID, Approver, Note, Status string
}

// Activities groups the registered activity implementations.
type Activities struct {
	Runtime runtime.AgentRuntime
	Notify  func(string, any)
	Persist Persister
}

// Persister abstracts the orchestrator store so activities don't bind to it.
type Persister interface {
	RecordCheckpointAwaiting(ctx context.Context, in RecordCheckpointInput) error
	ResolveCheckpoint(ctx context.Context, in ResolveCheckpointInput) error
}

// ExecuteSegment runs one segment of an acpx flow.
func (a *Activities) ExecuteSegment(ctx context.Context, in ExecuteSegmentInput) (SegmentResult, error) {
	ev := make(chan runtime.Event, 16)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for e := range ev {
			activity.RecordHeartbeat(ctx, e.Kind, e.NodeID)
			if a.Notify != nil {
				a.Notify(e.Kind, map[string]any{"runId": in.RunID, "nodeId": e.NodeID, "payload": e.Payload})
			}
		}
	}()
	res, err := a.Runtime.RunSegment(ctx, runtime.SegmentRequest{
		RunID: in.RunID, FlowTSPath: in.FlowTSPath, FromNodeID: in.FromNodeID, ToNodeID: in.ToNodeID,
		WorkspaceDir: in.WorkspaceDir,
	}, ev)
	close(ev)
	<-done
	if err != nil {
		return SegmentResult{}, err
	}
	return SegmentResult{NextNodeID: res.NextNodeID, Output: res.Output}, nil
}

// RecordCheckpointAwaiting writes a checkpoint_request row and emits the event.
func (a *Activities) RecordCheckpointAwaiting(ctx context.Context, in RecordCheckpointInput) error {
	if a.Persist != nil {
		return a.Persist.RecordCheckpointAwaiting(ctx, in)
	}
	return nil
}

// ResolveCheckpoint marks a checkpoint approved/rejected.
func (a *Activities) ResolveCheckpoint(ctx context.Context, in ResolveCheckpointInput) error {
	if a.Persist != nil {
		return a.Persist.ResolveCheckpoint(ctx, in)
	}
	return nil
}
