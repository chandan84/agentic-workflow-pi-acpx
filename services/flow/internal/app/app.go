// Package app wires the flow service application layer.
package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/author"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/ports"
	"github.com/google/uuid"
)

// Service is the flow service.
type Service struct {
	Store  ports.Store
	Bus    events.Bus
	Author *author.Author
}

// New builds a Service.
func New(s ports.Store, b events.Bus, a *author.Author) *Service {
	return &Service{Store: s, Bus: b, Author: a}
}

// CreateFlow creates a flow.
func (s *Service) CreateFlow(ctx context.Context, name, desc string) (ports.Flow, error) {
	f, err := s.Store.CreateFlow(ctx, ports.Flow{ID: uuid.NewString(), Name: name, Description: desc})
	if err != nil {
		return f, err
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectFlowCreated, f)
	}
	return f, nil
}

// Validate validates raw IR JSON.
func (s *Service) Validate(_ context.Context, raw string) []string {
	if err := flowir.ValidateJSON([]byte(raw)); err != nil {
		return []string{err.Error()}
	}
	ir, err := flowir.Parse([]byte(raw))
	if err != nil {
		return []string{err.Error()}
	}
	errs := flowir.DAGCheck(ir)
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return out
}

// Codegen renders the IR to acpx TypeScript.
func (s *Service) Codegen(_ context.Context, raw string) (string, error) {
	ir, err := flowir.Parse([]byte(raw))
	if err != nil {
		return "", err
	}
	return flowir.Codegen(ir)
}

// PutVersion stores a new IR version, validating before persisting and
// generating the .flow.ts alongside.
func (s *Service) PutVersion(ctx context.Context, flowID, raw string) (ports.FlowVersion, error) {
	ir, err := flowir.Parse([]byte(raw))
	if err != nil {
		return ports.FlowVersion{}, err
	}
	ts, err := flowir.Codegen(ir)
	if err != nil {
		return ports.FlowVersion{}, err
	}
	v, err := s.Store.PutVersion(ctx, ports.FlowVersion{
		ID: uuid.NewString(), FlowID: flowID, IRJSON: raw, FlowTS: ts,
	})
	if err == nil && s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectFlowVersionCreated, v)
	}
	return v, err
}

// Generate runs the flow-author skill and persists the produced IR.
type GenerateProgress func(line string)

// Generate streams log lines via progress callback and returns the persisted version.
func (s *Service) Generate(ctx context.Context, flowID, agentID, prompt string, progress GenerateProgress) (ports.FlowVersion, error) {
	stream := &streamAdapter{progress: progress, bus: s.Bus, ctx: ctx}
	if agentID == "" {
		agentID = "flow-author"
	}
	res, err := s.Author.Generate(ctx, agentID, "", prompt, stream)
	if err != nil {
		if s.Bus != nil {
			_ = s.Bus.Publish(ctx, events.SubjectFlowGenerationError, map[string]string{
				"flowId": flowID, "error": err.Error(),
			})
		}
		return ports.FlowVersion{}, err
	}
	// Use the produced IR's id as flow id when one isn't supplied.
	if flowID == "" {
		flowID = res.IR.ID
	}
	raw, _ := json.Marshal(json.RawMessage(res.RawIR))
	v, err := s.Store.PutVersion(ctx, ports.FlowVersion{
		ID: uuid.NewString(), FlowID: flowID, IRJSON: strings.TrimSpace(string(raw)), FlowTS: res.FlowTS,
	})
	if err != nil {
		return v, err
	}
	if s.Bus != nil {
		_ = s.Bus.Publish(ctx, events.SubjectFlowGenerationDone, v)
	}
	return v, nil
}

type streamAdapter struct {
	progress GenerateProgress
	bus      events.Bus
	ctx      context.Context
}

func (s *streamAdapter) OnLog(line string) {
	if s.progress != nil {
		s.progress(line)
	}
	if s.bus != nil {
		_ = s.bus.Publish(s.ctx, events.SubjectFlowGenerationLog, map[string]string{"line": line})
	}
}

func (s *streamAdapter) OnPartialIR(_ string) {}
