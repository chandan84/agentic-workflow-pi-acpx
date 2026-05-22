# Plan — Agentic Workflow Platform (pi · acpx · ACP · Temporal)

## Context

`agentic-workflow-pi-acpx` is a greenfield repo (only `README.md` + `LICENSE`). The goal is a
**cross-platform desktop app** that lets a user define **groups of isolated AI coding agents** (each a
`pi` agent), turn a **natural-language prompt into a multi-step workflow** (authored to **acpx** flow
conventions, driven over the **Agent Client Protocol**), **edit that flow on a WYSIWYG canvas**, and
**run it durably** (wrapped by **Temporal**) with **human checkpoints, full audit, replay, and observability**.

This iteration's deliverable is **documentation only**: an illustrative **design document with inline
diagrams** and an **infographic**, both stored in `docs/plan`, to serve as the north star for all
later implementation. No monorepo scaffolding or code is produced yet (per the chosen scope).

The design below is grounded in deep research of all references (ACP, acpx flows/CLI/replay-viewer,
pi coding-agent, Temporal, Tauri v2 + React Flow + tonic). The four architecture decisions were
confirmed with the user and are treated as locked.

## Locked decisions (from the user)

| Decision | Choice | Consequence baked into the design |
|---|---|---|
| Iteration scope | **Plan + docs only** | Produce design doc + infographic in `docs/plan`; no code. |
| Orchestration boundary | **Temporal wraps acpx runs (coarse)** | acpx is the workflow engine; Temporal is the durable run-level wrapper. Human gates handled by acpx `checkpoint` nodes; runs **segmented at checkpoints** so pauses don't block a single activity. |
| Temporal worker | **Rust preview SDK** | All-Rust worker; activities **shell out** to `acpx`/`pi` binaries. Pin `temporalio-sdk` + acpx/pi versions (preview API churn). |
| Dev infra | **Temporal CLI + SQLite** for durability; **Postgres** for app data | App spawns `temporal server start-dev --db-filename`; Postgres holds domain data only. |

## Grounded facts that shape the design

- **pi** = Node/TS coding-agent runtime. An agent is isolated by `cwd` + a project-local `.pi/`
  (`settings.json`, `extensions/`, `skills/`, `prompts/`) layered over global `~/.pi/agent/`.
  User-uploaded zips map to `pi install` / extraction into those dirs. pi exposes an RPC/JSON mode and
  **does not natively speak ACP**.
- **acpx** = Node/TS headless ACP client **and** flow engine. Ships a built-in **`pi` adapter** (the
  ACP bridge to pi). Flows are TypeScript `.flow.ts` via `defineFlow` with node types
  `acp | action | compute | decision | checkpoint`, `edges` + `decisionEdge`, a shared accumulating
  `outputs` payload, and a declared `permissions` block. CLI: `acpx <agent> exec|prompt`,
  `acpx flow run <file> --input-json ...`, `--approve-*`, `--format json`. Runs emit replayable
  bundles at `~/.acpx/flows/runs/<id>/` (`manifest.json` + append-only `trace.ndjson` + artifacts);
  there is a WebSocket replay viewer using ELK layout.
- **ACP** = JSON-RPC 2.0 over stdio; agent-as-subprocess; `initialize` (capability negotiation),
  `session/new` (`cwd`, `mcpServers`), `session/prompt` → streamed `session/update`,
  `session/request_permission`, `session/cancel`. Rust SDK exists (`agent-client-protocol`).
- **Temporal** = durable execution (Workflows + Activities + Workers, Signals, Queries, durable
  Timers, event history). **Rust SDK is Public Preview, not GA (May 2026)** → pin versions.
- **Frontend** = Tauri v2 (commands/events/channels + sidecar binaries), React Flow
  (`@xyflow/react`) + ELK for the canvas and path highlighting, `tonic` for Rust gRPC, Connect
  protocol only if the webview ever needs gRPC directly, Zustand for execution state.

## Target architecture (the north star the docs will detail)

**Layered responsibility model**

- **pi** — agent runtime; one isolated `.pi` workspace per agent.
- **acpx** — workflow engine + ACP client; runs `.flow.ts`, drives pi over ACP, owns intra-flow step
  routing, the shared `outputs`/work-item, human `checkpoint` gates, and trace bundles.
- **Temporal (CLI dev server, SQLite)** — durable wrapper: a Temporal **Workflow == a flow run**; it
  invokes acpx execution as **Activities**. To make long human pauses durable, the run is **segmented
  at acpx checkpoints** — each segment is a bounded, heartbeating activity (`acpx flow run` of that
  segment); between segments the Temporal workflow **awaits a Signal** (with a durable Timer for
  timeout) before resuming the next acpx segment. Gives durability, retry-on-transient-failure,
  scheduling, cancellation, and a durable run record.
- **Rust backend** — the core: tonic **gRPC services** (all remote access), Temporal **client +
  Rust-preview-SDK worker** (activities shell out to `acpx`/`pi`), **Postgres** persistence (sqlx),
  and the **Tauri core** (commands/events) bridging the UI.
- **Tauri + React frontend** — minimalistic, card-based UI; React Flow WYSIWYG canvas with ELK
  auto-layout and live executed-path highlighting fed by Tauri events (Rust backend tails the acpx
  trace + Temporal status). Zustand state.

**Flow representation — a JSON Flow IR is the single source of truth**

- The **"common pi agent"** (a default, non-modifiable pi agent per group, configured with a bundled
  `flow-author` skill encoding acpx authoring guidelines) turns the user prompt → **Flow IR** (nodes,
  edges/decisionEdge, input/output schemas, conditions, forks/joins, checkpoints, work-item schema,
  permissions, start/end).
- IR ⇄ acpx `.flow.ts` (codegen for execution + acpx replay-viewer compatibility) and IR ⇄ React
  Flow graph (canvas editing). **Validatable/verifiable** = JSON-schema validation + DAG checks
  (single start, reachable end, balanced fork/join, no orphans) + an `acpx flow` dry-run/validate.
- At run start, the UI collects the IR-declared input values; the run carries a **work-item JSONB**
  updated at each step (persisted in Postgres + reflected in acpx `outputs`).

**Audit / replay / observability (three complementary layers)**

1. **Temporal event history** — run-level durability + replay record.
2. **acpx `trace.ndjson` bundles** — intra-flow step trace + agent turns (prompts, tool calls,
   outputs); powers canvas path replay and/or the native acpx replay viewer.
3. **Postgres `audit_event`** (append-only) — business events: group/agent/flow create-edit, checkpoint
   approvals, resource uploads, agent invocations.

**Persistence (Postgres, app data only)** — `agents_group`, `agent` (pi workspace path + config),
`resource_bundle` (zip metadata + blob path) + `agent_resource` association, `flow` (current IR) +
`flow_version` (history), `run` (links Temporal workflow id + acpx run id), `work_item` (JSONB per
run), `checkpoint_request` (pending human gates), `audit_event`. Temporal state lives in its own
SQLite; acpx trace bundles live on disk, referenced from `run`.

**gRPC surface (tonic)** — `AgentGroupService` (CRUD groups/agents, upload+associate resources),
`WorkflowService` (generate-from-prompt, CRUD/validate flows, IR↔graph), `ExecutionService` (start
run, signal checkpoint, cancel, **server-streaming** live updates, fetch replay), `ObservabilityService`
(audit/trace queries). In-app the frontend uses Tauri IPC; gRPC is for remote access.

**Monorepo (isolated projects) — described in the doc, built in later iterations**

```
/proto                  gRPC contracts (shared)
/crates/                Rust workspace
  backend-core/         domain + Postgres (sqlx) + Temporal client
  grpc-gateway/         tonic services
  temporal-worker/      Rust preview-SDK worker; activities shell out to acpx/pi
  flow-ir/              Flow IR types + validation + acpx .flow.ts codegen
  tauri-app/ (src-tauri) Tauri core
/apps/desktop/          React + React Flow canvas (Zustand, Tauri IPC)
/agents/flow-author/    "common pi agent" config + flow-author skill (acpx guidelines)
/db/migrations/         Postgres migrations + seed
/scripts/               dev-up, cleanup, db-reset, temporal-dev (CLI+SQLite)
/infra/                 local dev orchestration (spawn temporal dev + postgres)
/docs/plan/             design doc + infographic  ← THIS iteration
```

## This iteration — deliverables (docs only, in `docs/plan`)

1. **`docs/plan/design-document.md`** — the illustrative north-star design doc with inline **Mermaid**
   diagrams. Sections:
   - Vision & objectives; glossary (ACP, acpx, pi, Temporal, Flow IR, Agents Group, work-item).
   - Personas & key user journeys.
   - System architecture (component diagram) + layered responsibility model and the "who owns what"
     table, explicitly stating the **acpx-engine / Temporal-wrapper** boundary.
   - Agents Group & Agent model (pi-dir isolation; resource zips → association) — diagram.
   - Flow lifecycle: prompt → common pi agent → Flow IR → canvas edit/validate → `.flow.ts` →
     Temporal-wrapped, checkpoint-segmented run → replay — sequence + state diagrams.
   - Flow IR spec (node types, edges/decisionEdge, conditions, forks/joins, checkpoints, work-item
     schema) — annotated JSON example.
   - Execution & durability model (Temporal wraps acpx; checkpoint segmentation; signals/timers;
     retries; **honest caveats**: coarse per-step durability, activity heartbeat/timeout limits,
     preview Rust SDK churn) — sequence diagram.
   - Human-in-the-loop checkpoints — sequence diagram.
   - Data model — **Postgres ER (Mermaid erDiagram)**.
   - gRPC service contracts (services + key RPCs incl. server-streaming).
   - Audit, replay & observability (the three layers).
   - UI/UX north star: minimalistic card-based screens (Groups, Agents, Resources, Flow canvas,
     Run/replay, Audit) — wireframe sketches.
   - Monorepo layout & isolated projects.
   - Local dev, cleanup & DB-reset, ER generation.
   - Phased delivery roadmap; risks & open questions.
2. **`docs/plan/infographic.svg`** — self-contained one-page visual (renders on GitHub): the layered
   stack, the prompt→flow→run→replay pipeline, and the card-based UI concept. Optionally also a
   rendered `infographic.png`.
3. **`docs/plan/README.md`** — short index of the above; also mirror this plan into `docs/plan` per
   the user's "store all your plans into docs/plan" instruction.

Format note: design doc = Markdown + Mermaid; infographic = hand-authored SVG. (Redirectable.)

## Verification (docs-only)

- Validate every Mermaid diagram parses (e.g. `npx -y @mermaid-js/mermaid-cli` on extracted blocks,
  or a markdown preview) — no syntax errors.
- Open `infographic.svg` (and PNG if rendered) to confirm it renders correctly.
- Check the doc's TOC/internal links and that all referenced URLs are the real reference sources.
- Commit and push to `claude/agent-workflow-orchestration-Qdu48` (`git push -u origin`).

## Explicitly out of scope this iteration

No Rust/TS/React code, no `.proto`, no migrations, no scripts, no scaffolding — only the `docs/plan`
design document + infographic (the north star). Building begins in a later iteration following the
roadmap in the design doc.
