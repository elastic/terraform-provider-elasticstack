## Context

The Terraform Plugin Framework requires that fields in a resource model struct that correspond to
attributes with `Computed` children or that can receive unknown values during planning must be
`types.Object` (or another framework-native type), not a pointer to a plain Go struct. When a plain
struct pointer is used and the framework needs to store an unknown for that attribute, it panics with
`Received unknown value, however the target type cannot handle unknown values`.

This problem manifested at the `analysis_config` level because `bucket_span` (Computed+Optional,
default `5m`) and `model_prune_window` (Computed) are nested inside it, and the v0.15.0 schema
addition of `custom_rules.scope` changed the coercion path so that Terraform must mark the whole
`analysis_config` object unknown during plan when the entire block is sourced from a variable.

The identical pattern was resolved for `detectors` in commit `bc23e6e2` (issue #2966) by changing
`Detectors []DetectorTFModel` → `types.List`. This change extends the same fix to:
1. `TFModel.AnalysisConfig *AnalysisConfigTFModel` → `types.Object`
2. `AnalysisConfigTFModel.PerPartitionCategorization *PerPartitionCategorizationTFModel` → `types.Object`

Point 2 is proactive: no user report exists yet for `per_partition_categorization`, but the same
structural limitation applies, and eliminating it now prevents the next report.

## Goals / Non-Goals

**Goals:**
- Eliminate the `Value Conversion Error` when `analysis_config` is sourced from a Terraform variable
  or `for_each`.
- Proactively remove the remaining struct-pointer (`PerPartitionCategorization`) from the
  `analysis_config` subtree.
- Follow the established conversion pattern used by `AnalysisLimits types.Object`,
  `DataDescription types.Object`, and `ModelPlotConfig types.Object` in the same resource.
- Add a regression acceptance test that would have caught both this issue and the prior #2966 issue.

**Non-Goals:**
- Changing the Terraform schema or any user-visible attribute names.
- Changing the Elastic ML API model (`models_api.go`).
- Modifying the `RequiresReplace()` plan modifier on `analysis_config` — that is correct API
  behaviour and is unrelated to the bug.
- Removing `RequiresReplace()` or `UseStateForUnknown()` plan modifiers from any attribute.

## Decisions

### 1. Change `TFModel.AnalysisConfig` to `types.Object`

Replace `*AnalysisConfigTFModel` with `types.Object`. This is the minimum fix for the reported
issue and is consistent with how the other `SingleNestedAttribute` fields in this resource are
typed (`AnalysisLimits`, `DataDescription`, `ModelPlotConfig`).

`getAnalysisConfigAttrTypes(ctx)` already exists and is derived directly from the schema, so no
schema helper work is required for this part.

### 2. Proactively change `PerPartitionCategorization` to `types.Object`

Extend the same fix to `AnalysisConfigTFModel.PerPartitionCategorization`. A new helper
`getPerPartitionCategorizationAttrTypes(ctx)` should be derived from
`getAnalysisConfigAttrTypes(ctx)` following the same pattern as `getDetectorAttrTypes(ctx)` which
navigates the schema from the top level.

This eliminates the last remaining struct-pointer in the `analysis_config` subtree. Without this
fix, a user who sources just `per_partition_categorization` from a variable would encounter the same
error in a future provider version.

### 3. Conversion pattern

All conversion sites must follow the framework-native pattern:

- **`toAPIModel()`**: Replace `plan.AnalysisConfig == nil` guard with
  `plan.AnalysisConfig.IsNull() || plan.AnalysisConfig.IsUnknown()`, then extract via
  `plan.AnalysisConfig.As(ctx, &analysisConfigTF, basetypes.ObjectAsOptions{})`.
  Inside that extracted struct, replace `analysisConfig.PerPartitionCategorization != nil` with
  `!analysisConfig.PerPartitionCategorization.IsNull() && !analysisConfig.PerPartitionCategorization.IsUnknown()`,
  then extract via `.As()`.

- **`convertAnalysisConfigFromAPI()`**: Change return type from `*AnalysisConfigTFModel` to
  `types.Object`. Build and return with
  `types.ObjectValueFrom(ctx, getAnalysisConfigAttrTypes(ctx), analysisConfigTF)`.
  For the `PerPartitionCategorization` field inside `AnalysisConfigTFModel`, change from storing
  a pointer to building a `types.Object` via
  `types.ObjectValueFrom(ctx, getPerPartitionCategorizationAttrTypes(ctx), perPartitionTF)`.
  When no `PerPartitionCategorization` should be stored, use
  `types.ObjectNull(getPerPartitionCategorizationAttrTypes(ctx))`.

- **`validateConfigCustomRules()`**: The existing guard `config.AnalysisConfig == nil` must change
  to `config.AnalysisConfig.IsNull() || config.AnalysisConfig.IsUnknown()`. Then extract
  `AnalysisConfigTFModel` via `.As()` before accessing `.Detectors`. No other logic changes are
  needed in `validate.go`.

- **`fromAPIModel()`**: The call site `plan.AnalysisConfig = plan.convertAnalysisConfigFromAPI(...)`
  must change to `plan.AnalysisConfig, d = plan.convertAnalysisConfigFromAPI(...)` to capture the
  returned `types.Object`.

### 4. Regression test

Add `TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig` with testdata at
`testdata/TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig/create/anomaly_detection.tf`.
The config should define an `analysis_config` variable typed as `object({...})` matching the full
`analysis_config` schema shape, assign `analysis_config = var.analysis_config`, run plan+apply, and
then run a plan-after-apply step to confirm the read path does not regress.

This mirrors the structure of `TestAccResourceAnomalyDetectionJobVariableSourcedDetectors`.

## Open Questions

- Does any other code path outside `models_tf.go` and `validate.go` access `plan.AnalysisConfig`
  fields directly, or is the grep result (`create.go` → `toAPIModel`, `read.go` → `fromAPIModel`)
  exhaustive?
- Does `objectplanmodifier.RequiresReplace()` on `analysis_config` interact with `types.Object`
  differently than with `*AnalysisConfigTFModel` during `ModifyPlan`? (Likely fine since other
  `types.Object` attributes use the same modifier, but worth verifying during implementation.)
- What exact Terraform/user config reproduces the unknown at the `analysis_config` level rather than
  at a child attribute level? A minimal repro would serve as the acceptance test foundation.

## Risks / Trade-offs

- [Conversion code complexity] The `convertAnalysisConfigFromAPI` function is already complex; the
  refactor requires careful handling of the prior-state access pattern where the plan's
  `AnalysisConfigTFModel` is used as the starting point for state-preserving round-trip logic.
  Mitigation: follow the pattern of `convertModelPlotConfigFromAPI` closely; write the new function
  incrementally and rely on the existing test matrix.
- [PerPartitionCategorization state-preservation logic] The existing logic uses
  `analysisConfigTF.PerPartitionCategorization` to check prior state for round-trip preservation.
  With `types.Object`, this must use `.IsNull()` and `.As()` instead of a nil check.
  Mitigation: map each existing `nil` check to `IsNull()` and each struct field access to an
  extracted variable.
- [No user report for PerPartitionCategorization] Fixing it now is proactive (YAGNI risk).
  Justification: the fix is symmetric with the `analysis_config` fix, adds minimal complexity,
  and prevents a follow-on report.

## Migration Plan

No migration steps for users. The change is internal to the Go representation; the Terraform schema
and state format are unchanged. Users who hit the bug will simply need to upgrade to the fixed
version.
