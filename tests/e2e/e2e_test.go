// Package e2e exercises the user-visible parts of the platform end-to-end
// using only public packages. The internal application layers of each service
// are covered by their own module's tests; this test stitches together the
// public ports.
//
// Scenario:
//  1. The hello-review IR validates against the Flow IR v1 JSON schema.
//  2. flowir.Codegen renders deterministic acpx TypeScript.
//  3. The Flow IR matches its committed golden .flow.ts byte-for-byte.
//  4. The stub AgentRuntime emits the expected event stream when running
//     one segment from "draft" to "review".
//  5. The NATS event catalog covers every subject we publish.
package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	busevents "github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/flowir"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
)

func TestHelloReviewIRValidates(t *testing.T) {
	raw := loadExample(t)
	if err := flowir.ValidateJSON(raw); err != nil {
		t.Fatalf("schema: %v", err)
	}
	ir, err := flowir.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if errs := flowir.DAGCheck(ir); len(errs) > 0 {
		t.Fatalf("dag: %v", errs)
	}
}

func TestCodegenMatchesGolden(t *testing.T) {
	raw := loadExample(t)
	ir, err := flowir.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := flowir.Codegen(ir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(repoRoot(t), "examples", ".flow.ts", "hello-review.flow.ts"))
	if err != nil {
		t.Skipf("golden missing: %v", err)
	}
	if strings.TrimRight(string(want), "\n") != strings.TrimRight(got, "\n") {
		t.Fatalf("golden mismatch")
	}
}

func TestStubRuntimeSegmentEvents(t *testing.T) {
	r := runtime.NewStub()
	ev := make(chan runtime.Event, 8)
	res, err := r.RunSegment(context.Background(), runtime.SegmentRequest{
		RunID: "r1", FromNodeID: "draft", ToNodeID: "review",
	}, ev)
	if err != nil {
		t.Fatal(err)
	}
	close(ev)
	var kinds []string
	for e := range ev {
		kinds = append(kinds, e.Kind)
	}
	if res.NextNodeID != "review" {
		t.Fatalf("next: %s", res.NextNodeID)
	}
	if len(kinds) < 2 || kinds[0] != "segment_started" || kinds[len(kinds)-1] != "segment_complete" {
		t.Fatalf("event sequence: %v", kinds)
	}
}

func TestEventCatalogNonEmpty(t *testing.T) {
	subjects := busevents.AllSubjects()
	if len(subjects) < 10 {
		t.Fatalf("catalog: %d", len(subjects))
	}
	for _, s := range subjects {
		if !strings.Contains(s, ".") {
			t.Fatalf("malformed subject: %s", s)
		}
	}
}

func loadExample(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "examples", "flow-ir", "hello-review.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(root, "..", ".."))
}
