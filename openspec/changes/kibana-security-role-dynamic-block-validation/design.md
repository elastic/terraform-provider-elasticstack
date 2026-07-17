## Context

`elasticstack_kibana_security_role` registers a `configValidator` that runs
`ValidateResource` during Terraform's `ValidateResourceConfig` RPC call. At that
point, attributes inside dynamic blocks may still be **unknown** — Terraform
resolves `for_each` expressions only during the plan phase, which comes after
config validation.

The current loop in `ValidateResource` (`validators.go:61-72`) iterates over all
elements of the `kibana` set and calls `kibanaPrivilegeCounts(obj)`. That helper
(`expand.go:49-59`) returns `featureLen = 0` when `feature.IsUnknown()` and
`baseLen = 0` when `base.IsUnknown()`. Because the caller treats `(0, 0)` as
"neither privilege set", `validateKibanaPrivileges(0, 0)` fires the error even
though the values will be known (and valid) at apply time.

The outer guard on line 58 (`if kibana.IsNull() || kibana.IsUnknown()`) only
skips the loop when the **set itself** is null/unknown — it does not protect
against individual elements or their attributes being unknown.

This issue was introduced with the Plugin Framework migration (PR #3071, v0.15.0).
The SDKv2 implementation placed validation inside Create/Update handlers where
all values are already resolved.

### Pattern precedent

`internal/utils/validators/exactly_one_of_nested.go` already implements the
correct pattern: it tracks `hasUnknown` across attributes and returns early from
validation when any controlling value is unknown (lines 75-88), deferring the
constraint to apply time.

## Goals

- Allow practitioners to use `dynamic` blocks on `kibana.feature` without a
  spurious plan-phase error.
- Preserve the plan-phase error for statically-incorrect configs (both `base` and
  `feature` omitted without a dynamic block).
- Do not touch API-facing validation or the apply-time path in `expandKibana`.

## Non-Goals

- Changing the apply-time `validateKibanaPrivileges` call inside `expandKibana`.
- Modifying validation for the `base` block when used without `feature`.
- Changing the `elasticsearch` null-presence check in `ValidateResource`
  (lines 40-51).
- Fixing `dynamic` block support on any other resource.

## Decisions

| Topic | Decision |
|-------|-----------|
| Detection approach | Check `obj.IsUnknown()`, `featureAttr.IsUnknown()`, and `baseAttr.IsUnknown()` before calling `kibanaPrivilegeCounts`. Skip (continue) when any is unknown. |
| Attribute access | Use `obj.Attributes()["feature"]` and `obj.Attributes()["base"]` directly — consistent with how `kibanaPrivilegeCounts` accesses them. |
| Apply-time enforcement | No change — `expandKibana` already calls `validateKibanaPrivileges` with fully-resolved values. |
| Acceptance test coverage | Add a test case with `dynamic "feature" { for_each = local.features … }` that exercises plan+apply. This must pass after the fix and fail before it. |
| Unit test coverage | Add a unit test that passes an `obj` with an unknown `feature` attribute to confirm validation is skipped. |

## Risks / Trade-offs

- **Deferred error**: When a practitioner uses a `dynamic` block but the
  `for_each` expression evaluates to an empty list at apply time (no `feature`
  blocks, no `base`), the "either one must be set" error is deferred from plan to
  apply — one extra Terraform cycle before the user sees it. This is the same
  trade-off accepted by `ExactlyOneOfNestedAttrsValidator` and is the correct
  behaviour for unknown-deferral.
- **Edge case — fully-unknown element**: If `for_each` references a
  completely unknown collection, the element itself (`obj`) may be unknown.
  The `obj.IsUnknown()` guard handles this case.

## Open Questions

- Should `expandKibana`'s call to `validateKibanaPrivileges` (`expand.go:244`)
  also gain an unknown-skip guard, or is that path guaranteed to receive
  fully-resolved values at apply time? (Most likely no change needed — the expand
  path runs during Create/Update where unknowns are already resolved.)
- Can the outer `kibana` set element itself (`obj`) be an unknown `types.Object`
  — for example when `for_each` references an entirely unknown collection? The
  `obj.IsUnknown()` guard in the fix handles this case; a unit test would confirm
  coverage.
- Are there other callers of `validateKibanaPrivileges` beyond `validators.go` and
  `expand.go` that need the same unknown-awareness treatment? A grep of the repo
  found only those two call sites.

## Implementation sketch

```go
// validators.go — inside ValidateResource, replacing the current loop body
for _, elem := range kibana.Elements() {
    obj, ok := elem.(types.Object)
    if !ok {
        resp.Diagnostics.AddError("Invalid kibana block", "unexpected element type")
        return
    }
    // Defer validation when the object itself or either controlling attribute is unknown.
    if obj.IsUnknown() {
        continue
    }
    featureAttr, featureOk := obj.Attributes()["feature"]
    baseAttr, baseOk := obj.Attributes()["base"]
    if (featureOk && featureAttr.IsUnknown()) || (baseOk && baseAttr.IsUnknown()) {
        continue
    }
    _, _, baseLen, featureLen := kibanaPrivilegeCounts(obj)
    resp.Diagnostics.Append(validateKibanaPrivileges(baseLen, featureLen)...)
    if resp.Diagnostics.HasError() {
        return
    }
}
```

Files touched: `internal/kibana/security_role/validators.go` (~6 lines added).

## Migration / State

No state migration needed. This is a config-validation-only change with no effect
on stored state.
