package runtime

import (
	"context"
	"testing"
)

func TestStubRunSegment(t *testing.T) {
	s := NewStub()
	ev := make(chan Event, 8)
	res, err := s.RunSegment(context.Background(), SegmentRequest{
		RunID: "r1", FromNodeID: "a", ToNodeID: "b",
	}, ev)
	if err != nil {
		t.Fatal(err)
	}
	if res.NextNodeID != "b" {
		t.Fatalf("next: %q", res.NextNodeID)
	}
}

func TestStubRunSkill(t *testing.T) {
	s := NewStub()
	ev := make(chan Event, 4)
	res, err := s.RunSkill(context.Background(), SkillRequest{
		AgentID: "a1", Skill: "draft", Prompt: "hello",
	}, ev)
	if err != nil {
		t.Fatal(err)
	}
	if res.Output["echo"] != "hello" {
		t.Fatalf("echo: %v", res.Output)
	}
}
