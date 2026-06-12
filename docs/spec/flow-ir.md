# Flow IR

The Flow IR is the canonical, versioned JSON document understood by every
service. Single source of truth: `pkg/flowir/schema/flow-ir.v1.json`. Go
mirror: `pkg/flowir/types.go`. Validators: `pkg/flowir/validate.go`.

## Schema versioning

The top-level `schemaVersion` is a `const` integer. v1 is the only released
schema. Breaking changes get a new integer and a new file
(`flow-ir.v2.json`); both versions are accepted in parallel for one minor
release before v(n-1) is removed. `pkg/flowir.SchemaV1` embeds the JSON
bytes so every service validates against the same artifact.

## Top-level shape

Required fields (from the schema): `schemaVersion`, `id`, `name`, `nodes`,
`edges`, `start`. Optional: `description`, `permissions`, `ends`,
`workItems`. The Go struct:

```go
type IR struct {
    SchemaVersion int
    ID, Name, Description string
    Permissions *Permissions
    Start string
    Ends  []string
    Nodes []Node
    Edges []Edge
    WorkItems []WorkItem
}
```

`additionalProperties: false` at the top level — unknown root fields are
rejected.

## Node kinds

Eight kinds, enumerated in `flow-ir.v1.json` and mirrored as constants in
`pkg/flowir/types.go`:

| Kind         | Constant            | Purpose                                                        |
|--------------|---------------------|----------------------------------------------------------------|
| `acp`        | `NodeACP`           | Invoke a pi agent skill. Uses `agentId`, `skill`, `prompt`.    |
| `action`     | `NodeAction`        | Call a registered host function. Uses `fn`, `args`.            |
| `compute`    | `NodeCompute`       | Pure expression evaluation; no side effects.                   |
| `decision`   | `NodeDecision`      | Branching node. Uses `branches[]` (`when`, `to`).              |
| `checkpoint` | `NodeCheckpoint`    | Pause for human approval. Uses `approvers`, `timeoutSeconds`, `onTimeout`. |
| `fork`       | `NodeFork`          | Parallel split. Must be balanced by a `join`.                  |
| `join`       | `NodeJoin`          | Parallel rejoin.                                               |
| `end`        | `NodeEnd`           | Terminal node. At least one must be reachable.                 |

Per-kind fields are loosely typed at the schema level (the schema sets
`additionalProperties: true` inside `node`) so authors can carry custom
metadata; the Go struct uses an `Extra` map for everything the strongly-typed
fields don't cover.

## Edge model

```json
{ "from": "node_a", "to": "node_b", "condition": "optional expression" }
```

`additionalProperties: false`. Edges are directed. A decision node MUST NOT
use the `branches[]` *and* be the `from` of conditional edges simultaneously;
`DAGCheck` enforces that non-decision nodes have no `branches` and decisions
have at least one.

## Decision branches

A decision node carries its own outgoing targets as `branches: [{ "when":
"expr", "to": "node-id" }, ...]`. `when` is a string expression evaluated by
the runtime; the first matching branch wins. Branches are treated as outgoing
edges for reachability analysis.

## Fork / join semantics

- A `fork` starts N parallel sub-paths.
- Each path must terminate at the same `join`.
- `DAGCheck` enforces equal counts of forks and joins
  (`unbalanced fork/join: X forks, Y joins`). Pairing by reachability is
  enforced at execution time by the orchestrator.
- Work items emitted in parallel branches carry independent ids; the `join`
  waits for all branches to complete (semantics owned by the orchestrator).

## Work-item schema

```json
{ "nodeId": "review", "kind": "approval", "schema": { ... } }
```

`kind` is one of `task | approval | review`. `schema` is a free-form JSON
object describing the payload the human user fills in. Work items are
materialised by the orchestrator into `runs.work_item` rows when their node
executes.

## Permissions

```json
"permissions": { "roles": ["operator", "approver"] }
```

Empty / missing means "any authenticated caller". Roles are matched against
the auth context provided by the gateway (see [security.md](./security.md)).

## Validation rules

Two stages:

1. **JSON-schema** — `flowir.ValidateJSON(raw []byte)` runs draft-07
   validation against `flow-ir.v1.json`.
2. **`DAGCheck(ir *IR) []error`** — structural checks beyond JSON-schema:
   - duplicate node ids
   - `start` references an existing node
   - every edge endpoint exists
   - every node reachable from `start`
   - decisions have ≥1 branch; non-decisions have none
   - decision branch targets exist
   - balanced fork/join counts
   - at least one reachable end or an `ends[]` declaration

Both are run by `flows.v1.FlowService.Validate` and by `PutVersion` before
persisting.

## hello-review example

The example IR shipped at
[`examples/flow-ir/hello-review.json`](../../examples/flow-ir/hello-review.json):

```json
{
  "schemaVersion": 1,
  "id": "hello-review",
  "name": "Hello + Review",
  "description": "Draft a greeting, get human approval, post it.",
  "start": "draft",
  "ends": ["done"],
  "nodes": [
    { "id": "draft",  "kind": "acp",        "title": "Draft greeting", "agentId": "writer-1", "skill": "draft", "prompt": "Write a friendly hello." },
    { "id": "review", "kind": "checkpoint", "title": "Human review",   "prompt": "Approve the draft?", "approvers": ["alice"], "timeoutSeconds": 3600, "onTimeout": "escalate" },
    { "id": "post",   "kind": "action",     "title": "Post greeting",  "fn": "post_message", "args": { "channel": "general" } },
    { "id": "done",   "kind": "end",        "title": "Done" }
  ],
  "edges": [
    { "from": "draft",  "to": "review" },
    { "from": "review", "to": "post" },
    { "from": "post",   "to": "done" }
  ],
  "workItems": [
    { "nodeId": "review", "kind": "approval" }
  ]
}
```

Codegen output at
[`examples/.flow.ts/hello-review.flow.ts`](../../examples/.flow.ts/hello-review.flow.ts):

```ts
// Generated by flowir.Codegen — do not edit.
import {
  defineFlow,
  acpNode, actionNode, computeNode, decisionNode,
  checkpointNode, forkNode, joinNode, endNode, edge,
} from "acpx";

export default defineFlow({
  id: "hello-review",
  name: "Hello + Review",
  start: "draft",
  nodes: [
    acpNode({ id: "draft", title: "Draft greeting", agentId: "writer-1", skill: "draft", prompt: "Write a friendly hello.", args: {} }),
    checkpointNode({ id: "review", title: "Human review", prompt: "Approve the draft?", approvers: [ "alice" ], timeoutSeconds: 3600, onTimeout: "escalate" }),
    actionNode({ id: "post", title: "Post greeting", fn: "post_message", args: { "channel": "general" } }),
    endNode({ id: "done", title: "Done" }),
  ],
  edges: [
    edge("draft", "review"),
    edge("review", "post"),
    edge("post", "done"),
  ],
});
```

## Segmentation at checkpoints

The orchestrator splits an IR run at every reachable `checkpoint` node.
Segments are `[current, nextCheckpointOrEnd]`. Each segment is one Temporal
activity. See [orchestrator-service.md](./orchestrator-service.md) for the
walk algorithm.
