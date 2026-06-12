package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

type persistStub struct {
	recorded  int
	resolved  int
}

func (p *persistStub) RecordCheckpointAwaiting(_ context.Context, _ RecordCheckpointInput) error {
	p.recorded++
	return nil
}
func (p *persistStub) ResolveCheckpoint(_ context.Context, _ ResolveCheckpointInput) error {
	p.resolved++
	return nil
}

func newSampleIR() flowir.IR {
	return flowir.IR{
		SchemaVersion: 1, ID: "demo", Name: "demo", Start: "draft",
		Ends: []string{"done"},
		Nodes: []flowir.Node{
			{ID: "draft", Kind: flowir.NodeAction, Fn: "noop"},
			{ID: "review", Kind: flowir.NodeCheckpoint, Prompt: "ok?", Approvers: []string{"alice"}, TimeoutSeconds: 60, OnTimeout: "escalate"},
			{ID: "post", Kind: flowir.NodeAction, Fn: "noop"},
			{ID: "done", Kind: flowir.NodeEnd},
		},
		Edges: []flowir.Edge{
			{From: "draft", To: "review"},
			{From: "review", To: "post"},
			{From: "post", To: "done"},
		},
	}
}

func TestRunFlow_HappyPath(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()

	activities := &Activities{Runtime: runtime.NewStub(), Persist: &persistStub{}}
	env.RegisterActivityWithOptions(activities.ExecuteSegment, activityOptsByName(ExecuteSegmentActivityName))
	env.RegisterActivityWithOptions(activities.RecordCheckpointAwaiting, activityOptsByName(RecordCheckpointAwaitingActivityName))
	env.RegisterActivityWithOptions(activities.ResolveCheckpoint, activityOptsByName(ResolveCheckpointActivityName))

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalCheckpointApproved, CheckpointDecision{
			CheckpointID: "demo:review", Approver: "alice", Note: "ok",
		})
	}, 100*time.Millisecond)

	env.ExecuteWorkflow(RunFlow, RunFlowInput{
		RunID: "r1", FlowID: "demo", IR: newSampleIR(),
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var res RunFlowResult
	require.NoError(t, env.GetWorkflowResult(&res))
	require.Equal(t, "done", res.FinalNodeID)
}

// activityOptsByName names activities so workflow.ExecuteActivity by-name lookups work.
func activityOptsByName(name string) activityOpts { return activityOpts{Name: name} }

type activityOpts = struct {
	Name                          string
	DisableAlreadyRegisteredCheck bool
	SkipInvalidStructFunctions    bool
}
