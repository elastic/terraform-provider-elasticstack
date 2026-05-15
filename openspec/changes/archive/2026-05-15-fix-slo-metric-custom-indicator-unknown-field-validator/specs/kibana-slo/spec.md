## MODIFIED Requirements

### Requirement: Conditional validator deferral for unknown attribute values (REQ-038)

When `RequiredIfDependentPathExpressionOneOf` or `RequiredIfDependentPathOneOf` is evaluating whether a field is required, if the attribute being validated is **unknown** at config-validation time (i.e. `val.IsUnknown()` is true), the validator SHALL return without error and defer the required-field check to apply time. This preserves correctness under Terraform 1.14 and later, where attribute values sourced from `for_each` map entries (via `each.value.*` interpolation) are marked unknown during `ValidateResourceConfig` even when the source data is statically known.

This behavior applies to the following attributes in `elasticstack_kibana_slo`:

- `metric_custom_indicator.good.metrics.field` — guarded by `RequiredIfDependentPathExpressionOneOf` when sibling `aggregation` is one of the non-`doc_count` aggregation values
- `metric_custom_indicator.total.metrics.field` — same
- `histogram_custom_indicator.good.from` and `histogram_custom_indicator.good.to` — guarded by `RequiredIfDependentPathExpressionOneOf` when sibling `aggregation` is `"range"`
- `histogram_custom_indicator.total.from` and `histogram_custom_indicator.total.to` — same

#### Scenario: Unknown field in for_each resource — no false-positive error

- GIVEN an `elasticstack_kibana_slo` resource configured with `for_each`
- AND `metric_custom_indicator.good.metrics.field` or `total.metrics.field` is set via `each.value.*` interpolation
- AND Terraform 1.14+ marks the field value as unknown at `ValidateResourceConfig` time
- WHEN `RequiredIfDependentPathExpressionOneOf` runs
- THEN the provider SHALL NOT emit a `"Attribute ... must be set"` diagnostic
- AND validation SHALL be deferred to apply time

#### Scenario: Known null field still fires error

- GIVEN `metric_custom_indicator.good.metrics.field` is null or empty string (not unknown)
- AND the sibling `aggregation` matches a value that makes `field` required
- WHEN `RequiredIfDependentPathExpressionOneOf` runs
- THEN the provider SHALL emit a `"Attribute ... must be set"` diagnostic
