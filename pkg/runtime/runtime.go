// Package runtime is the AgentRuntime port plus the acpx/pi child-process
// supervisor.
//
// The supervisor materializes a per-agent .pi/ workspace, mounts resource
// bundles, launches `acpx` (or `pi` for direct skill invocations) with the
// right env, captures stdout/stderr as structured events, and forwards
// heartbeats. Two implementations ship:
//
//   - Exec: spawns real child processes via os/exec
//   - Stub: returns canned events for tests and dev mode
//
// Both satisfy the AgentRuntime interface used by flow-service and
// orchestrator-service.
package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Event is the structured event stream emitted by an acpx/pi child process.
type Event struct {
	Kind      string         `json:"kind"`
	NodeID    string         `json:"nodeId,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp time.Time      `json:"ts"`
}

// AgentRuntime is the port: spawn one segment or one skill invocation,
// stream events back, and report completion.
type AgentRuntime interface {
	RunSegment(ctx context.Context, req SegmentRequest, events chan<- Event) (SegmentResult, error)
	RunSkill(ctx context.Context, req SkillRequest, events chan<- Event) (SkillResult, error)
	Shutdown(ctx context.Context) error
}

// SegmentRequest asks acpx to run a slice of one flow's IR between two checkpoints.
type SegmentRequest struct {
	RunID        string
	FlowTSPath   string
	FromNodeID   string
	ToNodeID     string
	WorkspaceDir string
	Env          map[string]string
}

// SegmentResult is the outcome of running one segment.
type SegmentResult struct {
	NextNodeID string
	Output     map[string]any
}

// SkillRequest asks the pi agent to run one named skill.
type SkillRequest struct {
	AgentID      string
	WorkspaceDir string
	Skill        string
	Prompt       string
	Args         map[string]any
	Env          map[string]string
}

// SkillResult is the outcome of invoking a skill.
type SkillResult struct {
	Output map[string]any
	Logs   []string
}

// ExecRuntime spawns real child processes.
type ExecRuntime struct {
	AcpxBin       string
	PiBin         string
	WorkspaceRoot string
}

// NewExec returns an exec-backed runtime.
func NewExec(acpx, pi, workspaceRoot string) *ExecRuntime {
	if acpx == "" {
		acpx = "acpx"
	}
	if pi == "" {
		pi = "pi"
	}
	if workspaceRoot == "" {
		workspaceRoot = filepath.Join(os.TempDir(), "awpa-workspaces")
	}
	return &ExecRuntime{AcpxBin: acpx, PiBin: pi, WorkspaceRoot: workspaceRoot}
}

// EnsureWorkspace materializes a .pi/ workspace for agentID under the root.
func (e *ExecRuntime) EnsureWorkspace(agentID string) (string, error) {
	dir := filepath.Join(e.WorkspaceRoot, agentID, ".pi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// RunSegment shells out to acpx and streams parsed JSON events.
func (e *ExecRuntime) RunSegment(ctx context.Context, req SegmentRequest, events chan<- Event) (SegmentResult, error) {
	if req.FlowTSPath == "" {
		return SegmentResult{}, errors.New("flow.ts path required")
	}
	args := []string{"flow", "run", req.FlowTSPath, "--run-id", req.RunID,
		"--from", req.FromNodeID, "--to", req.ToNodeID, "--json-events"}
	cmd := exec.CommandContext(ctx, e.AcpxBin, args...)
	cmd.Env = mergeEnv(req.Env)
	cmd.Dir = req.WorkspaceDir
	return e.runStreaming(cmd, events)
}

// RunSkill shells out to pi to invoke one named skill.
func (e *ExecRuntime) RunSkill(ctx context.Context, req SkillRequest, events chan<- Event) (SkillResult, error) {
	args := []string{"skill", "run", req.Skill, "--prompt", req.Prompt, "--json-events"}
	cmd := exec.CommandContext(ctx, e.PiBin, args...)
	cmd.Env = mergeEnv(req.Env)
	cmd.Dir = req.WorkspaceDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return SkillResult{}, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return SkillResult{}, err
	}
	var logs []string
	var output map[string]any
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		logs = append(logs, line)
		var ev Event
		if err := json.Unmarshal([]byte(line), &ev); err == nil && ev.Kind != "" {
			events <- ev
			if ev.Kind == "result" {
				output = ev.Payload
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		return SkillResult{Logs: logs}, err
	}
	return SkillResult{Output: output, Logs: logs}, nil
}

// Shutdown is a no-op for ExecRuntime (children are bound to ctx).
func (e *ExecRuntime) Shutdown(_ context.Context) error { return nil }

func (e *ExecRuntime) runStreaming(cmd *exec.Cmd, events chan<- Event) (SegmentResult, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return SegmentResult{}, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return SegmentResult{}, err
	}
	var res SegmentResult
	if err := streamJSON(stdout, events, &res); err != nil {
		return res, err
	}
	if err := cmd.Wait(); err != nil {
		return res, err
	}
	return res, nil
}

func streamJSON(r io.Reader, events chan<- Event, res *SegmentResult) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		if ev.Timestamp.IsZero() {
			ev.Timestamp = time.Now()
		}
		events <- ev
		if ev.Kind == "segment_complete" {
			if v, ok := ev.Payload["nextNodeId"].(string); ok {
				res.NextNodeID = v
			}
			if out, ok := ev.Payload["output"].(map[string]any); ok {
				res.Output = out
			}
		}
	}
	return scanner.Err()
}

func mergeEnv(extra map[string]string) []string {
	env := os.Environ()
	for k, v := range extra {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// StubRuntime returns canned events for tests and dev.
type StubRuntime struct {
	mu sync.Mutex
}

// NewStub returns a stub runtime.
func NewStub() *StubRuntime { return &StubRuntime{} }

// RunSegment emits a started + finished event and returns the next node.
func (s *StubRuntime) RunSegment(_ context.Context, req SegmentRequest, events chan<- Event) (SegmentResult, error) {
	events <- Event{Kind: "segment_started", NodeID: req.FromNodeID, Timestamp: time.Now()}
	events <- Event{Kind: "segment_complete", NodeID: req.ToNodeID, Payload: map[string]any{
		"nextNodeId": req.ToNodeID, "output": map[string]any{"stub": true},
	}, Timestamp: time.Now()}
	return SegmentResult{NextNodeID: req.ToNodeID, Output: map[string]any{"stub": true}}, nil
}

// RunSkill emits a result event with the prompt echoed back.
func (s *StubRuntime) RunSkill(_ context.Context, req SkillRequest, events chan<- Event) (SkillResult, error) {
	out := map[string]any{"echo": req.Prompt, "skill": req.Skill}
	events <- Event{Kind: "result", Payload: out, Timestamp: time.Now()}
	return SkillResult{Output: out, Logs: []string{"stub log line"}}, nil
}

// Shutdown is a no-op.
func (s *StubRuntime) Shutdown(_ context.Context) error { return nil }
