# ADR 0001: acpx as the flow execution engine

Status: Accepted · Date: 2026-06-11

## Context

Two candidate engines were evaluated for executing multi-agent flows over pi
agents: zigflow and acpx. acpx natively speaks ACP to pi agents, expresses
flows as `.flow.ts` programs, and produces replayable trace bundles. zigflow
would have required an adapter layer for ACP and duplicated orchestration we
already get from Temporal.

## Decision

acpx is the only flow engine. zigflow is dropped entirely. Temporal wraps acpx
coarse-grained: it never orchestrates individual nodes, only whole flow
segments (see ADR 0004), each executed by shelling out to `acpx flow run`.

## Consequences

- One engine to operate, one trace format to audit.
- The Flow IR codegen target is acpx `.flow.ts` only (ADR 0003).
- acpx becomes a hard runtime dependency of the orchestrator-service workers;
  its version is pinned in `services/orchestrator/config.example.yaml`.
- If acpx gains native checkpoint/resume, the segmentation scheme in ADR 0004
  can be simplified without touching the IR.
