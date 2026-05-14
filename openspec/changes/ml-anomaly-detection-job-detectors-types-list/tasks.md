## 1. Fix `AnalysisConfigTFModel.Detectors` field type

- [ ] 1.1 In `internal/elasticsearch/ml/anomalydetectionjob/models_tf.go`, change
  `AnalysisConfigTFModel.Detectors` from `[]DetectorTFModel` to `types.List` (line ~81).
- [ ] 1.2 Reuse the existing `getDetectorAttrTypes(ctx context.Context) map[string]attr.Type`
  helper from `internal/elasticsearch/ml/anomalydetectionjob/schema.go` for the full
  `DetectorTFModel` `attr.Type` map (including the `custom_rules` nested object type via
  `getCustomRuleAttrTypes`) when updating `convertAnalysisConfigFromAPI` and any
  `types.ListValueFrom` call. Do not add a duplicate helper in `models_tf.go`; if the
  helper must be relocated, move/centralize it rather than maintaining two copies.

## 2. Update `toAPIModel` detectors loop

- [ ] 2.1 In `models_tf.go` `toAPIModel` method, replace the direct `make([]DetectorAPIModel, len(analysisConfig.Detectors))` / `range analysisConfig.Detectors` pattern with:
  ```go
  var detectorsTF []DetectorTFModel
  diags.Append(analysisConfig.Detectors.ElementsAs(ctx, &detectorsTF, false)...)
  if diags.HasError() {
      return nil, diags
  }
  apiDetectors := make([]DetectorAPIModel, len(detectorsTF))
  for i, detector := range detectorsTF { ... }
  ```

## 3. Update `convertAnalysisConfigFromAPI` detectors assignment

- [ ] 3.1 In `models_tf.go` `convertAnalysisConfigFromAPI`, replace the final
  `analysisConfigTF.Detectors = detectorsTF` assignment (where `detectorsTF` is a
  `[]DetectorTFModel` slice) with:
  ```go
  detectorsListVal, d := types.ListValueFrom(ctx,
      types.ObjectType{AttrTypes: getDetectorAttrTypes(ctx)},
      detectorsTF)
  diags.Append(d...)
  analysisConfigTF.Detectors = detectorsListVal
  ```
- [ ] 3.2 Update the existing code that reads from `analysisConfigTF.Detectors` as a
  slice (e.g. `if len(analysisConfigTF.Detectors) > i` guard for preserving
  `originalDetector`): extract the prior `[]DetectorTFModel` from the incoming
  `analysisConfigTF.Detectors` using `ElementsAs` before the detectors loop, or use a
  local slice variable throughout.

## 4. Update `validateConfigCustomRules`

- [ ] 4.1 In `internal/elasticsearch/ml/anomalydetectionjob/validate.go`, change the
  `for i := range ac.Detectors` loop to guard with an `IsUnknown` / `IsNull` check on
  `ac.Detectors` first, then extract `[]DetectorTFModel` via `ElementsAs`:
  ```go
  if ac.Detectors.IsUnknown() || ac.Detectors.IsNull() {
      return diags
  }
  var detectors []DetectorTFModel
  diags.Append(ac.Detectors.ElementsAs(ctx, &detectors, false)...)
  if diags.HasError() {
      return diags
  }
  for i := range detectors {
      cr := detectors[i].CustomRules
      ...
  }
  ```

## 5. Acceptance test

- [ ] 5.1 In `internal/elasticsearch/ml/anomalydetectionjob/acc_test.go` (or associated
  testdata), add a test case (or `resource.TestStep`) that:
  - Declares a `variable "detectors"` with a `list(object({...}))` type and the minimal
    repro default from issue #2966.
  - Assigns `analysis_config = { detectors = var.detectors }` at the resource site.
  - Verifies that plan and apply succeed without a `Value Conversion Error`.
  - Includes a `resource.TestCheckResourceAttr` assertion confirming at least one
    detector attribute (e.g. `analysis_config.detectors.0.function`) reflects the
    expected value.
- [ ] 5.2 Ensure the new test is grouped under or named consistently with existing
  `TestAccResourceAnomalyDetectionJob*` tests.

## 6. Build and verify

- [ ] 6.1 Run `make build` and confirm the provider compiles without errors.
- [ ] 6.2 Run the targeted acceptance tests for the anomaly detection job resource to
  confirm existing and new tests pass (requires a running Elasticsearch stack; see
  [`dev-docs/high-level/testing.md`](../../../dev-docs/high-level/testing.md)).

## 7. OpenSpec

- [ ] 7.1 Keep delta spec
  `openspec/changes/ml-anomaly-detection-job-detectors-types-list/specs/elasticsearch-ml-anomaly-detection-job/spec.md`
  aligned with implementation; add normative requirements for the `types.List` field type
  and variable-sourced detectors test.
- [ ] 7.2 After merge: sync delta into
  `openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md` and run
  `make check-openspec`.
