## Context

`internal/kibana/dashboard/panel/sloerrorbudget/model.go`'s `PopulateFromAPI` receives `pm` (the
panel model being built — callers always pass it zero-valued to avoid aliasing plan pointers) and
`prior` (the prior plan on post-create/post-update read-back, or prior state on refresh; nil on
import). Current implementation:

```go
existing := pm.SloErrorBudgetConfig // always nil in production

if existing == nil {
    if prior != nil && prior.SloErrorBudgetConfig == nil {
        return nil // genuine type change away from slo_error_budget: leave config nil
    }
    pm.SloErrorBudgetConfig = &models.SloErrorBudgetConfigModel{}
    existing = pm.SloErrorBudgetConfig
}

existing.SloID = types.StringValue(apiConfig.SloId)

if typeutils.IsKnown(priorSloInstanceID) && apiConfig.SloInstanceId != nil && *apiConfig.SloInstanceId != "*" {
    existing.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
}

if (prior == nil || typeutils.IsKnown(existing.Title)) && apiConfig.Title != nil {
    existing.Title = types.StringValue(*apiConfig.Title)
}
// ...same pattern for Description, HideTitle, HideBorder
```

The allocation decision (`if existing == nil`) already keys correctly on `prior.SloErrorBudgetConfig`
for the type-change-vs-same-type-update question — this is *not* the bug PR #4404 fixed elsewhere.
The bug is in the four presentation-field conditionals immediately below: `existing.Title` is the
field on the struct just allocated on this same call (or, if `pm.SloErrorBudgetConfig` were ever
non-nil, still not the prior's field) — it is never the prior plan/state's `Title`. Since a fresh
`&models.SloErrorBudgetConfigModel{}`'s `Title` zero value is `types.StringNull()`,
`typeutils.IsKnown(existing.Title)` is always `false`, so the only way `apiConfig.Title` is ever
adopted is `prior == nil` (import). On every same-type update (`prior != nil`), all four
fields are skipped unconditionally — they stay at whatever `existing` already held, which for the
always-zero-valued `pm.SloErrorBudgetConfig` case is always null.

`SloInstanceID` a few lines above does not have this bug: its guard reads `priorSloInstanceID`, a
variable explicitly computed from `prior.SloErrorBudgetConfig.SloInstanceID` a few lines earlier
(lines 55-60), not from `existing`. The four presentation fields are the only fields in this
function using the wrong variable.

`sloerrorbudget`'s `BuildConfig` (the write path) already calls
`panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder, ...)` —
the shared helper other panel types use for this same field group. Only the read path
(`PopulateFromAPI`) is hand-rolled and buggy.

## Goals / Non-Goals

**Goals:**
- Make `title`, `description`, `hide_title`, and `hide_border` populate correctly from the API on
  every read (create, import, update, refresh).
- Apply REQ-009 null-preservation for these four fields keyed on `prior.SloErrorBudgetConfig`, not on
  `existing`/`pm.SloErrorBudgetConfig`, consistent with the contract PR #4404 documented for the
  other 10 panel types and already correctly followed by this same function's `SloInstanceID`
  handling.
- Add acceptance/unit coverage so a regression here is caught by CI, closing the blind spot the
  issue identified.

**Non-Goals:**
- Rewriting `sloerrorbudget/model.go`'s overall structure to match `sloburnrate`'s helper-extracted
  shape (`sloBurnRateConfigFromAPIImport` / `sloBurnRatePreserveNullIntentFromPrior` split, an
  `Import` vs `Update` two-path function shape). That refactor is a reasonable follow-up for
  consistency across panel types, but it is not required to fix this defect and would enlarge the
  diff beyond what issue #4423 asks for. This change adopts the shared `panelkit` field-level
  helpers (already used by `sloburnrate`) in place of the hand-rolled per-field conditionals, without
  restructuring the function's control flow.
- Any change to `SloID`, `SloInstanceID`, or `Drilldowns` handling — those are unaffected by this
  defect and already key correctly on `prior`.
- Any change to `BuildConfig` (write path) — it is already correct.

## Decisions

### 1. Fix the four field guards to key on `prior.SloErrorBudgetConfig`, using the shared `panelkit` helpers

Replace the four hand-rolled conditionals with the same two-step pattern `sloburnrate` uses for this
field group:

1. Unconditionally adopt the API value for each field (`panelkit.ApplyPresentationFromAPI`-style, or
   direct `types.StringPointerValue`/`types.BoolPointerValue` assignment) — this is correct on
   creation/import where there is no prior to preserve.
2. When `prior != nil && prior.SloErrorBudgetConfig != nil` (a genuine same-type update), reapply the
   prior's null intent via `panelkit.NullPreservePresentationFromPrior(prior.SloErrorBudgetConfig.Title,
   prior.SloErrorBudgetConfig.Description, prior.SloErrorBudgetConfig.HideTitle,
   prior.SloErrorBudgetConfig.HideBorder, &existing.Title, &existing.Description,
   &existing.HideTitle, &existing.HideBorder)`: if the prior field was null/unknown (practitioner
   never configured it), force the freshly-adopted API value back to null; otherwise leave the
   adopted API value in place.

Why:
- `panelkit.NullPreservePresentationFromPrior` already exists and is exercised by `sloburnrate`'s
  test suite; reusing it means this fix carries no new helper logic to review or maintain.
- Keys the decision on the same variable (`prior.SloErrorBudgetConfig`) already used correctly two
  lines above for `SloInstanceID` and for the function's own type-change/allocation branch —
  internally consistent with the rest of the function rather than introducing a second convention.
- Minimal diff: four conditionals become two helper calls; no function signature or control-flow
  change, no new exported symbols.

Alternatives considered:
- **Patch `typeutils.IsKnown(existing.Title)` to `typeutils.IsKnown(priorTitle)` in place**, computing
  `priorTitle`/`priorDescription`/`priorHideTitle`/`priorHideBorder` locals the same way
  `priorSloInstanceID` is computed today (lines 55-60), keeping the rest of the conditional
  structure unchanged. Rejected in favor of the `panelkit` helper call: it would work, but it
  duplicates logic `NullPreservePresentationFromPrior` already centralizes, and this function already
  imports `panelkit` for `BuildConfig`'s `BuildPresentationConfig` call, so pulling in its read-path
  counterpart is a one-line addition, not a new dependency.
- **Full `sloburnrate`-style rewrite** (extract `sloErrorBudgetConfigFromAPIImport` /
  `sloErrorBudgetPreserveNullIntentFromPrior`, split creation/import vs update into two top-level
  branches). Rejected as a Non-Goal above — larger diff than the defect requires, and changes the
  function's overall shape in a way better done as its own consistency-focused follow-up covering
  all panel types uniformly, not just this bug fix.

### 2. Scope acceptance test coverage to a create-then-refresh cycle setting all four fields

Add (or extend) an acceptance test that sets `title`, `description`, `hide_title`, and `hide_border`
inside `slo_error_budget_config`, applies, then runs a second `plan`/`apply` step (Terraform SDK's
standard no-diff-after-apply check) to assert no drift and no "Provider produced inconsistent result
after apply".

Why:
- This is the exact failure mode the issue describes and the exact gap current testdata leaves
  uncovered (existing fixtures set only `slo_id` and `drilldowns`).
- Matches the coverage shape `sloburnrate` and the other 10 panels fixed in PR #4404 already have for
  the same field group, keeping cross-panel test coverage consistent.

Alternatives considered:
- Unit-test-only coverage (no acceptance test). Rejected: the defect is specifically about
  `PopulateFromAPI`'s behavior across two calls with different `prior` values simulating a real
  refresh; a unit test can and should cover this directly (see Task 2), but an acceptance test is
  what would have caught this in CI before release, and the issue explicitly calls out the
  acceptance-testdata gap as the reason this went undetected.

## Risks / Trade-offs

- **Low risk.** The fix is scoped to four field assignments in one function; `SloID`, `SloInstanceID`,
  and `Drilldowns` handling are untouched.
- No behavior change for creation or import paths (`prior == nil`): those already adopt the API value
  unconditionally today and continue to do so.
- No API or schema changes; this is a state-population bug fix only.

## Migration Plan

None required. This is a bug fix to a read-path population function; no state migration, no schema
version bump. Practitioners currently working around the bug (e.g. by omitting these fields, or by
re-applying to fight the drift) see the fields round-trip correctly once this ships.

## Open Questions

None. The issue's suggested fix, the existing `SloInstanceID` handling in the same function, and the
established `sloburnrate` precedent all agree on the same approach (Decision 1).
