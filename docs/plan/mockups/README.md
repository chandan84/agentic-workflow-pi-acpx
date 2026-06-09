# UI prototype mockups

Light, card-style, minimalistic, modern. 1280 × 820 per screen. Same design language as the rest of
`docs/plan` (see `../design-document.md` §14 "UI/UX north star").

| # | Screen | SVG | PNG |
|---|---|---|---|
| 01 | Groups dashboard | [`01-groups.svg`](./01-groups.svg) | [`01-groups.png`](./01-groups.png) |
| 02 | Group detail (agents · resources · flows) | [`02-group-detail.svg`](./02-group-detail.svg) | [`02-group-detail.png`](./02-group-detail.png) |
| 03 | Create a flow from a prompt | [`03-new-flow-prompt.svg`](./03-new-flow-prompt.svg) | [`03-new-flow-prompt.png`](./03-new-flow-prompt.png) |
| 04 | WYSIWYG flow canvas editor | [`04-flow-canvas.svg`](./04-flow-canvas.svg) | [`04-flow-canvas.png`](./04-flow-canvas.png) |
| 05 | Run / replay (live executed-path highlight) | [`05-run-replay.svg`](./05-run-replay.svg) | [`05-run-replay.png`](./05-run-replay.png) |
| 06 | Audit / observability | [`06-audit.svg`](./06-audit.svg) | [`06-audit.png`](./06-audit.png) |

## Design tokens

- Surfaces — page `#f5f6fa` · cards `#ffffff` · borders `#e6e8ee` · soft drop-shadow
- Text — strong `#111827` · muted `#6b7280` · faint `#9ca3af`
- Primary accent — indigo `#4f46e5` · soft `#eef2ff`
- Status — green `#10b981` (done/healthy) · amber `#f59e0b` (awaiting/checkpoint) · blue `#2563eb` (action) · violet `#7c3aed` (acp)
- Node-type chips (canvas) — acp violet · action blue · compute indigo · decision amber · checkpoint amber-bordered
- Layout — persistent left sidebar (Groups · Flows · Runs · Audit · Settings) + 64 px top breadcrumb bar + card-first content
