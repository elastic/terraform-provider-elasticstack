## Why

`internal/kibana/dashboard/panel/sloerrorbudget/model.go`'s `PopulateFromAPI` never populates
`title`, `description`, `hide_title`, or `hide_border` in `slo_error_budget_config` on a same-type
update or refresh, even when the practitioner explicitly configured them. The function allocates a
fresh zero-valued `existing := &models.SloErrorBudgetConfigModel{}` when `pm.SloErrorBudgetConfig`
is nil (which it always is — callers pass a zero-valued `PanelModel`), and then gates each field's
API-value adoption on `typeutils.IsKnown(existing.Title)` (etc.) instead of on the prior plan/state
value. Since `existing` was just allocated, `existing.Title` is always null/unknown, so that
condition is always false whenever `prior != nil` (a same-type update) — the four fields are
permanently stuck at whatever they were seeded to (null, on refresh) regardless of what the
practitioner configured or what the API returns.

This produces `Error: Provider produced inconsistent result after apply` on the apply immediately
following any create where `title`, `description`, `hide_title`, or `hide_border` was set inside
`slo_error_budget_config`: the value round-trips correctly on create, then is forced back to null on
the very next refresh/update, contradicting the plan's known value.

This is the same underlying defect class fixed for 10 other dashboard panel types in PR #4404 (REQ-009
null-preservation must key on `prior.<Type>Config`, not on the panel model being built), but in a
differently-shaped instance that PR's sweep did not cover. That sweep's audit boundary was packages
with the literal `pm.<Type>Config == nil` "type-change recovery" branch mis-firing on every
same-type update. `sloerrorbudget`'s type-change/allocation decision (the `if existing == nil { if
prior != nil && prior.SloErrorBudgetConfig == nil { return nil } ... }` block) already keys
correctly on `prior`. The bug here is one level down: the per-field null-preservation checks within
that already-correct branch read `existing`'s own (freshly zero-valued) field instead of
`prior.SloErrorBudgetConfig`'s field, which is the inverse-shaped version of the same class of bug —
identified as a related-but-out-of-scope finding during PR #4404's review and filed as issue #4423.

No `sloerrorbudget` acceptance testdata currently sets `title`, `description`, `hide_title`, or
`hide_border` inside `slo_error_budget_config` (existing fixtures only set `slo_id` and
`drilldowns`), so this defect is undetected by CI today.

## What Changes

- Fix `PopulateFromAPI` in `internal/kibana/dashboard/panel/sloerrorbudget/model.go` so the
  presentation fields (`title`, `description`, `hide_title`, `hide_border`) are populated from the
  API on every read and, on a same-type update, have their null intent re-derived from
  `prior.SloErrorBudgetConfig`'s corresponding field instead of from the freshly-allocated
  `existing` struct — matching the `prior.<Type>Config`-keyed null-preservation contract already
  documented for other typed panel config blocks under REQ-009, and using the same
  `panelkit.ApplyPresentationFromAPI` / `panelkit.NullPreservePresentationFromPrior` helpers already
  adopted by `sloburnrate` and other panel types for this exact field group.
- Extend the `kibana-dashboard` capability's REQ-031 (SLO error budget panel behavior) with an
  explicit null-preservation requirement for `title`, `description`, `hide_title`, and
  `hide_border`, plus a scenario covering the create-then-refresh round trip that is currently
  broken.
- Add acceptance and unit test coverage that sets `title`, `description`, `hide_title`, and
  `hide_border` inside `slo_error_budget_config` and asserts they survive a refresh/update without
  producing "Provider produced inconsistent result after apply" — closing the exact CI blind spot
  called out in issue #4423.

## Capabilities

### Modified Capabilities
- `kibana-dashboard`: `slo_error_budget_config`'s `title`, `description`, `hide_title`, and
  `hide_border` fields must be populated from the API on read and must apply REQ-009
  null-preservation keyed on `prior.SloErrorBudgetConfig`, not on the panel model under
  construction.

## Impact

- `internal/kibana/dashboard/panel/sloerrorbudget/model.go`: `PopulateFromAPI` fix.
- `internal/kibana/dashboard/panel/sloerrorbudget/model_test.go`: new/updated unit tests for the four
  presentation fields' null-preservation and create/import population behavior.
- `internal/kibana/dashboard/panel/sloerrorbudget/acc_test.go` and its `testdata/` fixtures: new
  acceptance test coverage exercising `title`/`description`/`hide_title`/`hide_border` through a
  create-then-refresh cycle.
- `openspec/specs/kibana-dashboard/spec.md`: REQ-031 delta (this change's proposal only; the
  canonical spec is updated when this change is archived).
