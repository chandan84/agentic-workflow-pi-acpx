package author

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
)

type stubRT struct {
	out map[string]any
}

func (s stubRT) RunSegment(ctx context.Context, r runtime.SegmentRequest, ev chan<- runtime.Event) (runtime.SegmentResult, error) {
	return runtime.SegmentResult{}, nil
}
func (s stubRT) RunSkill(ctx context.Context, r runtime.SkillRequest, ev chan<- runtime.Event) (runtime.SkillResult, error) {
	return runtime.SkillResult{Output: s.out}, nil
}
func (s stubRT) Shutdown(context.Context) error { return nil }

func TestAuthorGenerateValidIR(t *testing.T) {
	ir := map[string]any{
		"schemaVersion": 1, "id": "f", "name": "n", "start": "a", "ends": []string{"b"},
		"nodes": []map[string]any{
			{"id": "a", "kind": "action", "fn": "noop"},
			{"id": "b", "kind": "end"},
		},
		"edges": []map[string]any{{"from": "a", "to": "b"}},
	}
	raw, _ := json.Marshal(ir)
	a := New(stubRT{out: map[string]any{"ir": string(raw)}})
	r, err := a.Generate(context.Background(), "a1", "/tmp", "make a flow", nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.IR.ID != "f" {
		t.Fatalf("ir id: %s", r.IR.ID)
	}
	if r.FlowTS == "" {
		t.Fatal("expected flow.ts output")
	}
}

func TestAuthorRetriesOnInvalid(t *testing.T) {
	a := New(stubRT{out: map[string]any{"ir": "{not json"}})
	a.MaxAttempts = 2
	if _, err := a.Generate(context.Background(), "a1", "/tmp", "make a flow", nil); err == nil {
		t.Fatal("expected failure")
	}
}
