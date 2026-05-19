## Why

`RequiredIfDependentPathExpressionOneOf` fires a false-positive error during `ValidateResourceConfig` when `metric_custom_indicator.{good,total}.metrics.field` receives an `each.value.*` interpolation inside a `for_each` resource. Under Terraform 1.14, values flowing through `for_each` — even from static `locals` — are marked **unknown** at config-validation time because the framework defers resolution to apply time.

The validator calls `attrValueIsUnsetForConditionalValidation(val)`, which returns `true` when `val.IsUnknown()`. Because the sibling `aggregation` attribute is already known (e.g. `"sum"`), the condition is met and the validator incorrectly concludes the required `field` attribute is missing. The field is not missing — it is simply not yet resolved.

The identical pattern was fixed for ML anomaly detection detectors in PR #2981 by returning early when the validated value is unknown.

## What Changes

- **Validator fix**: Add an early-return guard `if val == nil || val.IsUnknown() { return diags }` immediately after `var diags diag.Diagnostics` in the `validateValue` closure of both `RequiredIfDependentPathExpressionOneOf` and `RequiredIfDependentPathOneOf` in `internal/utils/validators/conditional.go`. This defers the required-field check to apply time when the attribute value is not yet resolved.
- **Unit test fix**: Update the existing `"invalid - current unknown, dependent matches required value"` test case in `TestRequiredIfDependentPathExpressionOneOf` (`conditional_test.go` ≈ line 799) to expect `expectedError: false` instead of `true`. Add a parallel test case to `TestRequiredIfDependentPathOneOf` for the same unknown scenario.
- **Requirements update**: Update the `kibana-slo` delta spec to document that `metric_custom_indicator.{good,total}.metrics.field` and `histogram_custom_indicator.{good,total}.from`/`to` validators SHALL defer to apply time when the attribute value is unknown at config-validation time.

## Capabilities

### New Capabilities
*(none)*

### Modified Capabilities

- `kibana-slo`: The validator on `metric_custom_indicator.good.metrics.field`, `metric_custom_indicator.total.metrics.field`, and (by shared validator fix) `histogram_custom_indicator.good.from`, `histogram_custom_indicator.good.to`, `histogram_custom_indicator.total.from`, `histogram_custom_indicator.total.to` SHALL skip the required-field check and defer to apply time when the attribute value is unknown at config-validation time. This resolves false-positive errors when `each.value.*` interpolation is used in `for_each` resources under Terraform 1.14.

## Impact

- `internal/utils/validators/conditional.go` — add unknown guard in `RequiredIfDependentPathExpressionOneOf.validateValue` (line ~541) and `RequiredIfDependentPathOneOf.validateValue` (line ~373)
- `internal/utils/validators/conditional_test.go` — flip `expectedError` for the unknown test case in `TestRequiredIfDependentPathExpressionOneOf` (~line 799); add parallel unknown case to `TestRequiredIfDependentPathOneOf`
- `openspec/specs/kibana-slo/spec.md` — update the canonical spec to incorporate this delta and document that these validators defer the required-field check to apply time when the attribute value is unknown during config validation
