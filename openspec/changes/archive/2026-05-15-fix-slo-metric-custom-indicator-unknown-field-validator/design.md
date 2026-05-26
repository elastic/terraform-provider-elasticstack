## Context

`RequiredIfDependentPathExpressionOneOf` (and `RequiredIfDependentPathOneOf`) in `internal/utils/validators/conditional.go` enforces that a field is set whenever a sibling dependent field matches one of the allowed values. The validator evaluates `attrValueIsUnsetForConditionalValidation(val)`, which returns `true` for `nil`, `IsNull()`, `IsUnknown()`, or an empty string.

Under Terraform 1.14, values that flow through a `for_each` map — even those derived from static `locals` — are marked as **potentially unknown** during `ValidateResourceConfig`. The framework defers resolution to apply time for any attribute that depends on the `for_each` key or value. This means `metric_custom_indicator.metrics.field = "path.${each.value.task_type}.success"` arrives at the validator as unknown, despite being fully resolvable at apply time.

Because the sibling `aggregation` attribute (e.g. `"sum"`) is already a known literal, the required condition fires correctly — but the validator then treats the unknown `field` as missing and emits a false-positive error.

The Plugin Framework guidance and the pattern established in PR #2981 (ML anomaly detection) both state that validators SHOULD return early without error when the attribute being validated is unknown, because unknown means "to be determined at apply time."

## Goals / Non-Goals

**Goals:**
- Eliminate the false-positive `"Attribute ... must be set"` error for `metric_custom_indicator.metrics.field` and any attribute guarded by `RequiredIfDependentPath*` when the attribute value is unknown at config-validation time.
- Keep `attrValueIsUnsetForConditionalValidation` intact for the `AllowedIf*` and `ForbiddenIf*` families, which already behave correctly for unknown values.

**Non-Goals:**
- Changing `AllowedIfDependentPath*` or `ForbiddenIfDependentPath*` — unknown already causes them to pass silently (`isSet = false`).
- Changing `attrValueIsUnsetForConditionalValidation` itself.
- Pinning or documenting Terraform version requirements.
- Changes to SLO API client models or API mapping.

## Decisions

### Decision 1: Inline unknown guard in `RequiredIf*` closures (Approach A)

Add `if val == nil || val.IsUnknown() { return diags }` at the top of the `validateValue` closure in both `RequiredIfDependentPathExpressionOneOf` and `RequiredIfDependentPathOneOf`, immediately after `var diags diag.Diagnostics`.

**Rationale:** Minimal diff. Two call sites in the same file. The guard is self-documenting (`IsUnknown()`) and directly mirrors PR #2981. No new helpers or exported surface.

**Alternative considered (Approach B):** Extract a separate `attrValueIsMissingForRequiredValidation` helper that excludes unknown. Rejected because it adds a nearly-identical unexported function without providing additional clarity beyond what `IsUnknown()` already communicates at the call site, and test changes are identical.

### Decision 2: Update existing test case, add parallel case

The `TestRequiredIfDependentPathExpressionOneOf` suite already has a `"invalid - current unknown, dependent matches required value"` case with `expectedError: true`. This must be flipped to `false`. An equivalent case must be added to `TestRequiredIfDependentPathOneOf` because that function uses the same pattern and is not currently tested for the unknown scenario.

## Risks / Trade-offs

- **[Accepted risk]** The guard defers validation entirely to apply time when the attribute is unknown. If a user genuinely does not set `field` (null, empty string), the apply-time validator will catch it. This is the correct behavior — validation of an unresolved value at plan time produces no signal.
- **[Accepted risk]** All other callers of `RequiredIfDependentPath*` (dashboard validators, other SLO indicators, etc.) also benefit from the fix. No regressions expected because the guard only suppresses a false-positive for unknown values; known null/empty values still trigger the error.

## Open Questions

- Does Terraform 1.14 also mark the **dependent** attribute (`aggregation`) as unknown in any real `for_each` pattern? If so, the existing `checkPathExpression` unknown-skip already handles it (line 125: `!pathValue.IsUnknown()`). Confirmation needed if users report a variant where `aggregation` itself is dynamic.
- Should `histogram_custom_indicator` `from`/`to` Float64 attributes get an explicit unit test case, or does fixing the shared validator suffice? The shared fix covers them, but an explicit test improves documentation.
- Should a Terraform 1.14 end-to-end acceptance test (with `for_each` and `each.value.*` interpolation) be added to the SLO acceptance suite, or only unit-level validator tests?
