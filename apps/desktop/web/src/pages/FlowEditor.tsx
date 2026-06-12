import { useMemo } from "react";
import { useParams } from "react-router-dom";
import {
  ReactFlow,
  Background,
  Controls,
  type Node,
  type Edge,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { irToReactFlow, kindChip, type FlowIR } from "../lib/ir";

// Mirrors examples/flow-ir/hello-review.json; replaced by
// flowsClient.getFlowVersion once the generated client is wired up.
const MOCK_IR: FlowIR = {
  schemaVersion: 1,
  id: "flw_bug_triage",
  name: "Bug triage & fix",
  start: "triage",
  ends: ["done"],
  nodes: [
    { id: "triage", kind: "acp", title: "Triage bug", agentId: "agt_triage" },
    {
      id: "severity",
      kind: "decision",
      title: "Severity?",
      branches: [
        { when: "severity == 'high'", to: "fix" },
        { when: "severity != 'high'", to: "backlog" },
      ],
    },
    { id: "fix", kind: "acp", title: "Implement fix", agentId: "agt_dev" },
    { id: "backlog", kind: "action", title: "File to backlog" },
    {
      id: "approve",
      kind: "checkpoint",
      title: "Human review",
      approvers: ["role:reviewer"],
      timeoutSeconds: 86400,
      onTimeout: "escalate",
    },
    { id: "done", kind: "end", title: "Done" },
  ],
  edges: [
    { from: "triage", to: "severity" },
    { from: "severity", to: "fix", condition: "severity == 'high'" },
    { from: "severity", to: "backlog", condition: "severity != 'high'" },
    { from: "fix", to: "approve" },
    { from: "approve", to: "done" },
    { from: "backlog", to: "done" },
  ],
};

export function FlowEditor(): JSX.Element {
  const { flowId } = useParams();

  // TODO(generated): flowsClient.getFlowVersion({ flowId }) -> ir
  const ir = MOCK_IR;

  const { nodes, edges } = useMemo(() => {
    const graph = irToReactFlow(ir);
    const rfNodes: Node[] = graph.nodes.map((n) => {
      const chip = kindChip(n.data.ir.kind);
      return {
        id: n.id,
        position: n.position,
        data: { label: `${chip.label} · ${n.data.ir.title ?? n.id}` },
        style: {
          background: chip.bg,
          color: chip.fg,
          border: `1px solid ${chip.fg}`,
          borderRadius: 8,
          fontSize: 12,
          fontWeight: 600,
        },
      };
    });
    const rfEdges: Edge[] = graph.edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
      label: e.label,
    }));
    return { nodes: rfNodes, edges: rfEdges };
  }, [ir]);

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <div className="row" style={{ justifyContent: "space-between", marginBottom: 12 }}>
        <div>
          <div style={{ fontWeight: 600 }}>{ir.name}</div>
          <div className="muted" style={{ fontSize: 12 }}>
            {flowId} · schema v{ir.schemaVersion} · read-only preview (editing lands
            with the generated client)
          </div>
        </div>
        <div className="row" style={{ gap: 8 }}>
          <button
            onClick={() => {
              // TODO(generated): flowsClient.validateFlow({ ir })
              window.alert("ValidateFlow not yet wired up; awaiting generated client");
            }}
          >
            Validate
          </button>
          <button
            className="primary"
            onClick={() => {
              // TODO(generated): flowsClient.saveFlowVersion({ flowId, ir })
              window.alert("SaveFlowVersion not yet wired up; awaiting generated client");
            }}
          >
            Save version
          </button>
        </div>
      </div>
      <div style={{ flex: 1, minHeight: 420, border: "1px solid var(--border)", borderRadius: 8 }}>
        <ReactFlow nodes={nodes} edges={edges} fitView nodesDraggable={false} nodesConnectable={false}>
          <Background />
          <Controls />
        </ReactFlow>
      </div>
    </div>
  );
}
