# docs/plan

North-star planning artifacts for the **Agentic Workflow Platform** (pi · acpx · ACP · Temporal ·
Tauri/React · Rust · Postgres).

| File | What it is |
|---|---|
| [`design-document.md`](./design-document.md) | The illustrative design document with inline Mermaid diagrams (architecture, Flow IR, execution/durability, ER model, gRPC surface, UI, roadmap). The single source of truth for intent and shape. |
| [`infographic.svg`](./infographic.svg) | One-page visual: the layered stack, the prompt → run → replay pipeline, agents-group isolation, audit layers, and roadmap. (Editable source.) |
| [`infographic.png`](./infographic.png) | Rendered raster copy of the infographic for universal preview. |
| [`PLAN.md`](./PLAN.md) | The approved iteration plan (mirrored here per the "store all plans in docs/plan" convention). |
| [`mockups/`](./mockups) | Six light-themed, card-style UI prototype illustrations for the key screens (Groups · Group detail · Create flow from prompt · Flow canvas · Run / replay · Audit). Both SVG sources and rendered PNGs. |

## Status

- **This iteration: documentation only.** No monorepo scaffolding or code yet.
- Build work begins in later phases per the roadmap in `design-document.md` §17.

## Locked architecture decisions

1. **Scope** — plan + docs only this iteration.
2. **Orchestration** — Temporal *wraps* acpx runs (coarse); acpx is the workflow engine; runs are
   segmented at checkpoints for durable human-in-the-loop.
3. **Temporal worker** — all-Rust using the preview SDK; activities shell out to `acpx`/`pi` (versions
   pinned).
4. **Dev infra** — Temporal CLI + SQLite for durability; Postgres for app domain data.
