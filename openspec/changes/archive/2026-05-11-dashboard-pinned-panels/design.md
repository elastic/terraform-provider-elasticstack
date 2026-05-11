## Context

`kbn-dashboard-data.pinned_panels` is a list of control-group entries with the same shape as the four panel-level control schemas (`kbn-controls-schemas-controls-group-schema-{options-list,range-slider,time-slider,esql}-control`). The Terraform resource already types these control schemas as `*_control_config` blocks under `panels[]`. Pinned panels are the same controls in a different position — perfect reuse opportunity.

## Goals / Non-Goals

**Goals:**
- Expose dashboard-level pinned controls without re-defining the four control config schemas.
- Validation rules for "exactly one config block, matching the declared `type`" should be identical between in-grid panel controls and pinned controls.
- Future expansion of any control's typed fields (see `expand-control-fields` change) automatically applies here too.

**Non-Goals:**
- New typed control kinds (none in spec).
- Grid attributes for pinned controls — the API does not place pinned controls on the grid.
- Migration helpers between pinned and in-grid control representations.

## Decisions

- **Reuse, don't redefine**: import the same `*_control_config` nested attribute schemas used by `panels[]`. Implementation likely extracts them into a shared schema builder if not already done.
- **Top-level placement**: `pinned_panels` is a sibling of `panels`, not nested inside `panels`. Matches API shape.
- **Discriminator validators**: reuse the existing conditional validators that enforce "exactly one `*_control_config` block, matching `type`" on `panels[]`.
- **No grid attribute**: pinned controls do not have `grid` in the API; the schema omits it.
- **Normalization**: rely on existing per-control normalization (already covered by REQ-026 / REQ-027 / REQ-028 / REQ-029).
- **Coupling with `expand-control-fields`**: schemas are shared, so whichever change lands second picks up the other's improvements automatically. No special sequencing logic required in code.

## Risks / Trade-offs

- [Risk] Practitioners may put a pinned control inside `panels` (or vice versa) and be confused → Mitigation: clear schema-level descriptions and an acceptance test demonstrating both placements.
- [Risk] Empty-vs-unset semantics — Kibana defaults `pinned_panels` to `[]` → Mitigation: same handling as REQ-009; unset stays unset on read.
