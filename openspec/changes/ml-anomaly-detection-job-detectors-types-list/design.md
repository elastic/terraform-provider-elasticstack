## Context

Canonical requirements for this resource live in
[`openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md`](../../specs/elasticsearch-ml-anomaly-detection-job/spec.md).
Implementation lives in
[`internal/elasticsearch/ml/anomalydetectionjob/`](../../../internal/elasticsearch/ml/anomalydetectionjob/).

The regression was introduced by PR #2877 (`feat(elasticsearch): custom_rules.scope for
ML anomaly detection job`) which added `ValidateConfig` to the resource. That method
calls `req.Config.Get(ctx, &config)` at plan time — the first code path that decodes the
full `TFModel` during planning. Terraform 1.14 marks all values that flow through a
variable as potentially-unknown at the config phase, including fully-concrete defaults.
The Framework reflection layer rejects assignment of such a value to `[]DetectorTFModel`
because a raw Go slice cannot represent unknown state; `types.List` can.

v0.14.5 had no `ValidateConfig` so full config decoding only happened at apply time
(when all values are fully concrete).

## Goals / Non-Goals

**Goals:**

- Fix the plan-time `Value Conversion Error` for `analysis_config.detectors` sourced
  from a Terraform variable or module input.
- Make `AnalysisConfigTFModel.Detectors` consistent with `CategorizationFilters` and
  `Influencers` (both already `types.List`) in the same struct.
- Add an acceptance test that exercises `var.detectors` assignment to prevent future
  regressions.

**Non-goals:**

- Changing `DetectorTFModel` itself (its fields already use `types.String`, `types.Bool`,
  `types.List` correctly).
- Changing `CustomRuleTFModel` or `ScopeEntryTFModel` — not implicated.
- Modifying the Terraform schema (`schema.go`) — `schema.ListNestedAttribute` is correct.
- Shipping a v0.15.1 patch release (fix targets `main`).
- State migration — the Terraform schema type exposed to users is unchanged; only the
  internal Go model field type changes.

## Decisions

- **Field type**: `AnalysisConfigTFModel.Detectors` changes from `[]DetectorTFModel` to
  `types.List`. Element type is
  `types.ObjectType{AttrTypes: getDetectorAttrTypes(ctx)}`.

- **`toAPIModel` (detectors loop)**: replace direct slice range with:
  ```go
  var detectorsTF []DetectorTFModel
  diags.Append(analysisConfig.Detectors.ElementsAs(ctx, &detectorsTF, false)...)
  ```
  `IsUnknown` guard is not needed here because `toAPIModel` is only called at apply time
  (after the framework has resolved all values).

- **`convertAnalysisConfigFromAPI` (detectors loop)**: replace direct slice assignment
  with `types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getDetectorAttrTypes(ctx)}, detectorsTF)`.
  A helper `getDetectorAttrTypes(ctx)` is needed; it can be derived from
  `schema.ListNestedAttribute` or declared as a function returning the `attr.Types` map
  for `DetectorTFModel`. Other helpers (`getCustomRuleAttrTypes`, `getRuleConditionAttrTypes`)
  already follow this pattern in `internal/elasticsearch/ml/anomalydetectionjob/schema.go`
  and are reused from `models_tf.go`.

- **`validateConfigCustomRules` guard**: add `ac.Detectors.IsUnknown()` check before
  `ElementsAs`. If unknown, return early (validation is deferred to apply time, matching
  existing behavior for `CustomRules.IsUnknown()`).

- **Acceptance test**: add a sub-test or separate test function for the variable-based
  repro. Use `resource.TestStep` with a Terraform configuration that mirrors the minimal
  repro from the issue (`variable "detectors"` with the same default, assigned to
  `analysis_config = { detectors = var.detectors }`). The test should verify the resource
  plans and applies without error.

## Risks / Trade-offs

- **Three method changes**: `toAPIModel`, `convertAnalysisConfigFromAPI`, and
  `validateConfigCustomRules` all need updating. Forgetting one would leave a latent
  decode issue. Tasks are explicit about each site.
- **`getDetectorAttrTypes` helper**: must be complete (include `custom_rules` nested
  object type). If the map is incomplete, `ElementsAs` will return a diagnostic error.
  Review against the full `DetectorTFModel` struct before finalizing.
- **Import path**: the `convertAnalysisConfigFromAPI` change must preserve the existing
  state-diff–safe semantics (preserving `null` vs empty-list distinction for `custom_rules`,
  preserving `null` for `detector_description` when the user omitted it). The only change
  is the final assignment; the slice population logic is unchanged.

## Open Questions

- Does Terraform 1.14 mark all variable-referenced values as potentially-unknown during
  the config phase even when a concrete default is present? (**Answered by @tobio**:
  yes — the minimal repro variable reference is not fully known at plan time. The
  acceptance test should treat the config as confirming the fix works for this case.)
- Should the fix be bundled into a v0.15.1 patch, or merged to main only for the next
  minor? (**Answered by @tobio**: main only.)
- Are there acceptance test fixtures that would exercise `var.detectors` assignment?
  (**Answered by @tobio**: add one based on the TF config in the issue body.)
