package flowir

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGoldenHelloReview regenerates examples/.flow.ts/hello-review.flow.ts
// from examples/flow-ir/hello-review.json. Set REGEN_GOLDEN=1 to overwrite.
func TestGoldenHelloReview(t *testing.T) {
	root, _ := os.Getwd()
	// pkg/flowir -> repo root
	repo := filepath.Clean(filepath.Join(root, "..", ".."))
	in := filepath.Join(repo, "examples", "flow-ir", "hello-review.json")
	out := filepath.Join(repo, "examples", ".flow.ts", "hello-review.flow.ts")

	raw, err := os.ReadFile(in)
	if err != nil {
		t.Skipf("input missing: %v", err)
	}
	ir, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Codegen(ir)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("REGEN_GOLDEN") == "1" {
		_ = os.MkdirAll(filepath.Dir(out), 0o755)
		if err := os.WriteFile(out, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(out)
	if err != nil {
		t.Skipf("golden missing (run with REGEN_GOLDEN=1): %v", err)
	}
	if string(want) != got {
		t.Fatalf("golden mismatch — regen with REGEN_GOLDEN=1\n--- want\n%s\n--- got\n%s", want, got)
	}
}
