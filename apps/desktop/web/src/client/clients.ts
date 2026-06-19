import { createClient } from "@connectrpc/connect";
import { AgentsService } from "../gen/agents/v1/agents_pb";
import { AuditService } from "../gen/audit/v1/audit_pb";
import { ExecutionService } from "../gen/execution/v1/execution_pb";
import { FlowService } from "../gen/flows/v1/flows_pb";
import { transport } from "./transport";

// All four service clients point at the gateway (single endpoint).
// The gateway proxies agents, flow, and execution to their upstream
// services; audit it serves directly from its NATS-fed aggregator.
export const agentsClient = createClient(AgentsService, transport);
export const flowClient = createClient(FlowService, transport);
export const executionClient = createClient(ExecutionService, transport);
export const auditClient = createClient(AuditService, transport);
