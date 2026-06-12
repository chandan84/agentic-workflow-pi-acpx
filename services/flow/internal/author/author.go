// Package author drives the flow-author skill: prompt → Flow IR.
//
// The author runs the configured AgentRuntime to invoke the flow-author skill
// with the user's prompt and a system message that instructs it to produce
// strictly-valid Flow IR v1 JSON. The streamed output is forwarded to the
// caller; on completion the JSON is validated against the IR schema + DAG
// validators. If validation fails the validator error is fed back to the agent
// and the run is retried up to MaxAttempts times.
package author

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
)

// Author runs the flow-author skill.
type Author struct {
	Runtime     runtime.AgentRuntime
	MaxAttempts int
}

// New builds an Author.
func New(r runtime.AgentRuntime) *Author { return &Author{Runtime: r, MaxAttempts: 3} }

// Result is the produced + validated IR.
type Result struct {
	IR     *flowir.IR
	RawIR  []byte
	FlowTS string
	Logs   []string
}

// Stream is a sink for partial output during generation.
type Stream interface {
	OnLog(line string)
	OnPartialIR(raw string)
}

type noopStream struct{}

func (noopStream) OnLog(string)       {}
func (noopStream) OnPartialIR(string) {}

// Generate runs the flow-author skill and validates the produced IR.
func (a *Author) Generate(ctx context.Context, agentID, workspace, prompt string, s Stream) (*Result, error) {
	if s == nil {
		s = noopStream{}
	}
	attempts := a.MaxAttempts
	if attempts <= 0 {
		attempts = 1
	}
	var lastValidationErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		effectivePrompt := prompt
		if lastValidationErr != nil {
			effectivePrompt = fmt.Sprintf("%s\n\nPrevious attempt failed validation:\n%s\nReturn corrected IR JSON only.", prompt, lastValidationErr.Error())
		}
		ev := make(chan runtime.Event, 16)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for e := range ev {
				if e.Kind == "log" {
					if msg, ok := e.Payload["message"].(string); ok {
						s.OnLog(msg)
					}
				} else if e.Kind == "partial" {
					if raw, ok := e.Payload["ir"].(string); ok {
						s.OnPartialIR(raw)
					}
				}
			}
		}()
		res, err := a.Runtime.RunSkill(ctx, runtime.SkillRequest{
			AgentID: agentID, WorkspaceDir: workspace,
			Skill: "flow-author", Prompt: effectivePrompt,
		}, ev)
		close(ev)
		<-done
		if err != nil {
			return nil, fmt.Errorf("attempt %d: runtime: %w", attempt, err)
		}
		raw, err := extractIR(res.Output, res.Logs)
		if err != nil {
			lastValidationErr = err
			continue
		}
		ir, err := flowir.Parse(raw)
		if err != nil {
			lastValidationErr = err
			continue
		}
		if errs := flowir.DAGCheck(ir); len(errs) > 0 {
			lastValidationErr = joinErrors(errs)
			continue
		}
		ts, err := flowir.Codegen(ir)
		if err != nil {
			return nil, err
		}
		return &Result{IR: ir, RawIR: raw, FlowTS: ts, Logs: res.Logs}, nil
	}
	if lastValidationErr == nil {
		lastValidationErr = errors.New("flow-author produced no IR")
	}
	return nil, fmt.Errorf("flow-author validation failed after %d attempts: %w", attempts, lastValidationErr)
}

func extractIR(output map[string]any, logs []string) ([]byte, error) {
	if output != nil {
		if v, ok := output["ir"]; ok {
			switch t := v.(type) {
			case string:
				return []byte(t), nil
			case map[string]any:
				return json.Marshal(t)
			}
		}
	}
	for i := len(logs) - 1; i >= 0; i-- {
		var v any
		if err := json.Unmarshal([]byte(logs[i]), &v); err == nil {
			if m, ok := v.(map[string]any); ok {
				if _, ok := m["schemaVersion"]; ok {
					return json.Marshal(m)
				}
			}
		}
	}
	return nil, errors.New("no IR in output")
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	msg := ""
	for _, e := range errs {
		msg += "- " + e.Error() + "\n"
	}
	return errors.New(msg)
}
