# flow-author

The shared pi agent configuration that drives prompt → Flow IR generation in
flow-service. Prompts and the acpx-authoring guideline live alongside the code
as data files so they can be edited without recompiling Go.

Files:

- `.pi/config.yaml` — pi agent config (model, tools, scopes).
- `skills/flow-author.md` — the skill prompt, including the system message and
  the response contract (strict Flow IR v1 JSON).
- `guidelines/acpx-authoring.md` — the authoring guideline the agent must
  follow when constructing the IR.
- `schemas/flow-ir.v1.json` — a copy of `pkg/flowir/schema/flow-ir.v1.json`
  embedded for the agent's local reference. Keep in sync via CI check.
