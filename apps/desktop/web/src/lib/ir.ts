/**
 * Flow IR <-> React Flow conversion helpers.
 *
 * Types mirror pkg/flowir/schema/flow-ir.v1.json. Round-trippable: any IR that
 * passes the schema converts to ReactFlow nodes/edges and back to an
 * IR with the same nodes, edges, kinds, ids, decision branches, and
 * checkpoint approval metadata.
 */

export type NodeKind =
  | "acp"
  | "action"
  | "compute"
  | "decision"
  | "checkpoint"
  | "fork"
  | "join"
  | "end";

export type CheckpointTimeout = "escalate" | "abort" | "auto-approve";

export interface DecisionBranch {
  when: string;
  to: string;
}

export interface IRNode {
  id: string;
  kind: NodeKind;
  title?: string;
  agentId?: string;
  skill?: string;
  prompt?: string;
  fn?: string;
  args?: Record<string, unknown>;
  branches?: DecisionBranch[];
  approvers?: string[];
  timeoutSeconds?: number;
  onTimeout?: CheckpointTimeout;
}

export interface IREdge {
  from: string;
  to: string;
  condition?: string;
}

export type WorkItemKind = "task" | "approval" | "review";

export interface IRWorkItem {
  nodeId: string;
  kind: WorkItemKind;
  schema?: Record<string, unknown>;
}

export interface IRPermissions {
  roles?: string[];
}

export interface FlowIR {
  schemaVersion: 1;
  id: string;
  name: string;
  description?: string;
  permissions?: IRPermissions;
  start: string;
  ends?: string[];
  nodes: IRNode[];
  edges: IREdge[];
  workItems?: IRWorkItem[];
}

// ---- ReactFlow shapes (kept minimal so we don't take a hard dep on
// the @xyflow/react types here — useful for unit tests that don't load
// the DOM build).

export interface RFPosition {
  x: number;
  y: number;
}

export interface RFNodeData {
  ir: IRNode;
}

export interface RFNode {
  id: string;
  type: NodeKind;
  position: RFPosition;
  data: RFNodeData;
}

export interface RFEdgeData {
  ir: IREdge;
}

export interface RFEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
  data: RFEdgeData;
}

export interface ReactFlowGraph {
  nodes: RFNode[];
  edges: RFEdge[];
}

const COL_W = 220;
const ROW_H = 120;

/**
 * Lay nodes out left-to-right by BFS depth from the `start` node.
 * Deterministic — important for snapshot tests.
 */
function layout(ir: FlowIR): Map<string, RFPosition> {
  const depth = new Map<string, number>();
  const adj = new Map<string, string[]>();
  for (const e of ir.edges) {
    const list = adj.get(e.from) ?? [];
    list.push(e.to);
    adj.set(e.from, list);
  }
  const queue: string[] = [ir.start];
  depth.set(ir.start, 0);
  while (queue.length > 0) {
    const cur = queue.shift() as string;
    const d = depth.get(cur) ?? 0;
    for (const next of adj.get(cur) ?? []) {
      if (!depth.has(next)) {
        depth.set(next, d + 1);
        queue.push(next);
      }
    }
  }
  // Group nodes by depth, then by their order in ir.nodes for stability.
  const byDepth = new Map<number, string[]>();
  for (const node of ir.nodes) {
    const d = depth.get(node.id) ?? 0;
    const row = byDepth.get(d) ?? [];
    row.push(node.id);
    byDepth.set(d, row);
  }
  const positions = new Map<string, RFPosition>();
  for (const [d, ids] of byDepth) {
    ids.forEach((id, i) => {
      positions.set(id, { x: d * COL_W, y: i * ROW_H });
    });
  }
  return positions;
}

export function irToReactFlow(ir: FlowIR): ReactFlowGraph {
  const positions = layout(ir);
  const nodes: RFNode[] = ir.nodes.map((node) => ({
    id: node.id,
    type: node.kind,
    position: positions.get(node.id) ?? { x: 0, y: 0 },
    data: { ir: node },
  }));
  const edges: RFEdge[] = ir.edges.map((edge) => {
    const rf: RFEdge = {
      id: `${edge.from}->${edge.to}`,
      source: edge.from,
      target: edge.to,
      data: { ir: edge },
    };
    if (edge.condition !== undefined) {
      rf.label = edge.condition;
    }
    return rf;
  });
  return { nodes, edges };
}

/**
 * Convert a ReactFlow graph back into IR. Required `start` and `name` cannot
 * be recovered from RF state alone, so they are passed via `meta`.
 */
export interface IRMeta {
  schemaVersion: 1;
  id: string;
  name: string;
  description?: string;
  permissions?: IRPermissions;
  start: string;
  ends?: string[];
  workItems?: IRWorkItem[];
}

export function reactFlowToIR(graph: ReactFlowGraph, meta: IRMeta): FlowIR {
  const nodes: IRNode[] = graph.nodes.map((n) => n.data.ir);
  const edges: IREdge[] = graph.edges.map((e) => e.data.ir);
  const ir: FlowIR = {
    schemaVersion: meta.schemaVersion,
    id: meta.id,
    name: meta.name,
    start: meta.start,
    nodes,
    edges,
  };
  if (meta.description !== undefined) ir.description = meta.description;
  if (meta.permissions !== undefined) ir.permissions = meta.permissions;
  if (meta.ends !== undefined) ir.ends = meta.ends;
  if (meta.workItems !== undefined) ir.workItems = meta.workItems;
  return ir;
}

/** Extract round-trip metadata from an existing IR. */
export function extractMeta(ir: FlowIR): IRMeta {
  const meta: IRMeta = {
    schemaVersion: ir.schemaVersion,
    id: ir.id,
    name: ir.name,
    start: ir.start,
  };
  if (ir.description !== undefined) meta.description = ir.description;
  if (ir.permissions !== undefined) meta.permissions = ir.permissions;
  if (ir.ends !== undefined) meta.ends = ir.ends;
  if (ir.workItems !== undefined) meta.workItems = ir.workItems;
  return meta;
}

/**
 * Lightweight structural validation: ids unique, start references a node,
 * every edge references existing nodes, decision branches reference existing
 * nodes. The authoritative validator lives in the gateway-service.
 */
export function validateIR(ir: FlowIR): string[] {
  const errors: string[] = [];
  const ids = new Set<string>();
  for (const n of ir.nodes) {
    if (ids.has(n.id)) errors.push(`duplicate node id: ${n.id}`);
    ids.add(n.id);
  }
  if (!ids.has(ir.start)) errors.push(`start references unknown node: ${ir.start}`);
  for (const end of ir.ends ?? []) {
    if (!ids.has(end)) errors.push(`end references unknown node: ${end}`);
  }
  for (const e of ir.edges) {
    if (!ids.has(e.from)) errors.push(`edge.from unknown: ${e.from}`);
    if (!ids.has(e.to)) errors.push(`edge.to unknown: ${e.to}`);
  }
  for (const n of ir.nodes) {
    for (const b of n.branches ?? []) {
      if (!ids.has(b.to)) errors.push(`branch.to unknown on ${n.id}: ${b.to}`);
    }
  }
  return errors;
}

/** Visual style chip per node kind — matches docs/plan/mockups palette. */
export function kindChip(kind: NodeKind): { bg: string; fg: string; label: string } {
  switch (kind) {
    case "acp":
      return { bg: "#f3e8ff", fg: "#7c3aed", label: "acp" };
    case "action":
      return { bg: "#dbeafe", fg: "#2563eb", label: "action" };
    case "compute":
      return { bg: "#e0e7ff", fg: "#4f46e5", label: "compute" };
    case "decision":
      return { bg: "#fef3c7", fg: "#b45309", label: "decision" };
    case "checkpoint":
      return { bg: "#fef3c7", fg: "#b45309", label: "checkpoint" };
    case "fork":
      return { bg: "#e5e7eb", fg: "#111827", label: "fork" };
    case "join":
      return { bg: "#e5e7eb", fg: "#111827", label: "join" };
    case "end":
      return { bg: "#d1fae5", fg: "#047857", label: "end" };
  }
}
