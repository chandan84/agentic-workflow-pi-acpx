import { create } from "zustand";

export interface AgentGroup {
  id: string;
  name: string;
  description: string;
  createdAt: string;
}

export interface Agent {
  id: string;
  groupId: string;
  name: string;
  role: string;
  piWorkspace: string;
  env: Record<string, string>;
}

export interface ResourceBundle {
  id: string;
  name: string;
  kind: "mcp" | "tool" | "fs" | "secret";
  spec: string;
}

export interface FlowSummary {
  id: string;
  name: string;
  description: string;
  createdAt: string;
}

export interface RunSummary {
  id: string;
  flowId: string;
  flowVersionId: string;
  status:
    | "pending"
    | "running"
    | "awaiting_checkpoint"
    | "completed"
    | "failed"
    | "cancelled";
  currentNode: string;
  startedAt: string;
  updatedAt: string;
}

export interface CheckpointRequest {
  id: string;
  runId: string;
  nodeId: string;
  prompt: string;
  status: "awaiting" | "approved" | "rejected" | "timeout";
  deadline: string;
}

export interface FeatureFlags {
  streamingFlowAuthor: boolean;
  liveRunReplay: boolean;
  auditTail: boolean;
}

export interface AppState {
  // Static config surfaced from import.meta.env so views can show it.
  gatewayUrl: string;
  flags: FeatureFlags;
  setGatewayUrl(url: string): void;
  setFlag(name: keyof FeatureFlags, value: boolean): void;

  // Cached server-driven collections (populated from ConnectRPC calls once
  // generated clients exist). Pages currently seed these from mock data.
  groups: AgentGroup[];
  agents: Agent[];
  resources: ResourceBundle[];
  flows: FlowSummary[];
  runs: RunSummary[];
  checkpoints: CheckpointRequest[];

  setGroups(groups: AgentGroup[]): void;
  setAgents(agents: Agent[]): void;
  setResources(resources: ResourceBundle[]): void;
  setFlows(flows: FlowSummary[]): void;
  setRuns(runs: RunSummary[]): void;
  setCheckpoints(checkpoints: CheckpointRequest[]): void;
}

const defaultFlags: FeatureFlags = {
  streamingFlowAuthor: true,
  liveRunReplay: true,
  auditTail: true,
};

export const useAppStore = create<AppState>((set) => ({
  gatewayUrl:
    (import.meta.env.VITE_GATEWAY_URL as string | undefined) ??
    "http://localhost:7000",
  flags: defaultFlags,
  setGatewayUrl: (url: string) => set({ gatewayUrl: url }),
  setFlag: (name, value) =>
    set((state) => ({ flags: { ...state.flags, [name]: value } })),

  groups: [],
  agents: [],
  resources: [],
  flows: [],
  runs: [],
  checkpoints: [],

  setGroups: (groups) => set({ groups }),
  setAgents: (agents) => set({ agents }),
  setResources: (resources) => set({ resources }),
  setFlows: (flows) => set({ flows }),
  setRuns: (runs) => set({ runs }),
  setCheckpoints: (checkpoints) => set({ checkpoints }),
}));
