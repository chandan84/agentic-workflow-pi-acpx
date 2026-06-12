import { describe, it, expect } from "vitest";
import {
  irToReactFlow,
  reactFlowToIR,
  extractMeta,
  validateIR,
  type FlowIR,
} from "./ir";

// Exact shape of examples/flow-ir/hello-review.json — duplicated here so the
// test stays self-contained and doesn't depend on a JSON import path.
const helloReview: FlowIR = {
  schemaVersion: 1,
  id: "hello-review",
  name: "Hello + Review",
  description: "Draft a greeting, get human approval, post it.",
  start: "draft",
  ends: ["done"],
  nodes: [
    {
      id: "draft",
      kind: "acp",
      title: "Draft greeting",
      agentId: "writer-1",
      skill: "draft",
      prompt: "Write a friendly hello.",
    },
    {
      id: "review",
      kind: "checkpoint",
      title: "Human review",
      prompt: "Approve the draft?",
      approvers: ["alice"],
      timeoutSeconds: 3600,
      onTimeout: "escalate",
    },
    {
      id: "post",
      kind: "action",
      title: "Post greeting",
      fn: "post_message",
      args: { channel: "general" },
    },
    { id: "done", kind: "end", title: "Done" },
  ],
  edges: [
    { from: "draft", to: "review" },
    { from: "review", to: "post" },
    { from: "post", to: "done" },
  ],
  workItems: [{ nodeId: "review", kind: "approval" }],
};

describe("ir <-> reactflow", () => {
  it("validates the sample without errors", () => {
    expect(validateIR(helloReview)).toEqual([]);
  });

  it("maps each IR node kind onto a ReactFlow node type 1:1", () => {
    const { nodes } = irToReactFlow(helloReview);
    expect(nodes.map((n) => [n.id, n.type])).toEqual([
      ["draft", "acp"],
      ["review", "checkpoint"],
      ["post", "action"],
      ["done", "end"],
    ]);
  });

  it("preserves all edges and produces stable ids", () => {
    const { edges } = irToReactFlow(helloReview);
    expect(edges.map((e) => e.id)).toEqual([
      "draft->review",
      "review->post",
      "post->done",
    ]);
  });

  it("round-trips: ir -> reactflow -> ir is deep-equal", () => {
    const graph = irToReactFlow(helloReview);
    const meta = extractMeta(helloReview);
    const back = reactFlowToIR(graph, meta);
    expect(back).toEqual(helloReview);
  });

  it("preserves decision branches across the round-trip", () => {
    const withDecision: FlowIR = {
      schemaVersion: 1,
      id: "triage",
      name: "Triage",
      start: "classify",
      ends: ["close", "fix"],
      nodes: [
        {
          id: "classify",
          kind: "decision",
          title: "Classify",
          branches: [
            { when: "bug", to: "fix" },
            { when: "noise", to: "close" },
          ],
        },
        { id: "fix", kind: "end", title: "Fix" },
        { id: "close", kind: "end", title: "Close" },
      ],
      edges: [
        { from: "classify", to: "fix", condition: "bug" },
        { from: "classify", to: "close", condition: "noise" },
      ],
    };
    expect(validateIR(withDecision)).toEqual([]);
    const graph = irToReactFlow(withDecision);
    const back = reactFlowToIR(graph, extractMeta(withDecision));
    expect(back).toEqual(withDecision);
    // Edge conditions surface as labels for the canvas.
    expect(graph.edges.map((e) => e.label)).toEqual(["bug", "noise"]);
    // Branches survive on the node.
    const classify = graph.nodes.find((n) => n.id === "classify");
    expect(classify?.data.ir.branches).toEqual([
      { when: "bug", to: "fix" },
      { when: "noise", to: "close" },
    ]);
  });

  it("flags unknown edge targets", () => {
    const broken: FlowIR = {
      schemaVersion: 1,
      id: "broken",
      name: "Broken",
      start: "a",
      nodes: [{ id: "a", kind: "action" }],
      edges: [{ from: "a", to: "ghost" }],
    };
    const errs = validateIR(broken);
    expect(errs).toContain("edge.to unknown: ghost");
  });
});
