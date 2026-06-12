package flowir

import (
	"encoding/json"
	"testing"
)

const sampleIR = `{
  "schemaVersion": 1,
  "id": "f1",
  "name": "hello",
  "start": "n1",
  "ends": ["n3"],
  "nodes": [
    { "id": "n1", "kind": "acp", "agentId": "a1", "skill": "draft", "prompt": "hi" },
    { "id": "n2", "kind": "checkpoint", "prompt": "ok?", "approvers": ["alice"], "timeoutSeconds": 60, "onTimeout": "escalate" },
    { "id": "n3", "kind": "end" }
  ],
  "edges": [
    { "from": "n1", "to": "n2" },
    { "from": "n2", "to": "n3" }
  ]
}`

func TestValidateAndParse(t *testing.T) {
	if err := ValidateJSON([]byte(sampleIR)); err != nil {
		t.Fatalf("validate: %v", err)
	}
	ir, err := Parse([]byte(sampleIR))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ir.ID != "f1" || len(ir.Nodes) != 3 {
		t.Fatalf("ir mismatch: %+v", ir)
	}
	if errs := DAGCheck(ir); len(errs) != 0 {
		t.Fatalf("dag: %v", errs)
	}
}

func TestDAGCheckDetectsOrphan(t *testing.T) {
	ir := &IR{
		SchemaVersion: 1, ID: "x", Name: "x", Start: "a",
		Nodes: []Node{{ID: "a", Kind: "action", Fn: "noop"}, {ID: "b", Kind: "end"}, {ID: "c", Kind: "end"}},
		Edges: []Edge{{From: "a", To: "b"}},
	}
	errs := DAGCheck(ir)
	if len(errs) == 0 {
		t.Fatal("expected unreachable error")
	}
}

func TestCodegenStable(t *testing.T) {
	ir, err := Parse([]byte(sampleIR))
	if err != nil {
		t.Fatal(err)
	}
	a, err := Codegen(ir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Codegen(ir)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("non-deterministic codegen output")
	}
	// Re-parse, re-emit, must round-trip stably.
	raw, _ := json.Marshal(ir)
	ir2, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := Codegen(ir2)
	if a != c {
		t.Fatalf("codegen not deterministic across re-parse")
	}
}
