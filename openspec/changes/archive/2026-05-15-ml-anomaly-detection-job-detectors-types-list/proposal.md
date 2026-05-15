## Why

`elasticstack_elasticsearch_ml_anomaly_detection_job` fails to plan when
`analysis_config.detectors` flows from a Terraform module input variable (or any
expression whose value carries unknown-origin metadata at config-evaluation time).
The error — `Value Conversion Error … Target Type: []anomalydetectionjob.DetectorTFModel
… Suggested Type: basetypes.ListValue` — appears on every plan, with no viable userland
workaround. The same configuration plans cleanly on v0.14.5; v0.15.0 introduced
`ValidateConfig` which calls `req.Config.Get(ctx, &config)` at plan time. That decode
path rejects a `[]DetectorTFModel` field when the list's elements carry framework
unknown-origin metadata (which Terraform 1.14 marks on any value referenced via a
variable).

**Root cause**: `AnalysisConfigTFModel.Detectors` is typed `[]DetectorTFModel`. The
Plugin Framework cannot decode a list carrying unknown-origin metadata into a raw Go
slice; it requires `types.List` (`basetypes.ListValue`), consistent with the framework's
own suggestion in the error message and with other list fields in the same struct
(`CategorizationFilters types.List`, `Influencers types.List`).

## What Changes

- **`AnalysisConfigTFModel.Detectors`** (`models_tf.go:81`): change field type from
  `[]DetectorTFModel` to `types.List`. The schema (`schema.go`) is **not** changed —
  `schema.ListNestedAttribute` for `detectors` remains correct.
- **`toAPIModel`** (`models_tf.go` detectors loop): replace the direct range over the
  slice with `Detectors.ElementsAs(ctx, &detectorsTF, false)` before iterating.
- **`convertAnalysisConfigFromAPI`** (`models_tf.go` detectors loop): replace
  `analysisConfigTF.Detectors = detectorsTF` with
  `types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getDetectorAttrTypes(ctx)}, detectorsTF)`.
- **`validateConfigCustomRules`** (`validate.go`): replace the direct range over the
  slice with a `Detectors.IsUnknown()` guard followed by `Detectors.ElementsAs` before
  the loop.
- **Acceptance test**: add a new test case (or sub-test) for
  `TestAccResourceAnomalyDetectionJob` that assigns `analysis_config.detectors` from a
  Terraform `variable` (matching the minimal repro in the issue) to prevent future
  regressions of this class.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-ml-anomaly-detection-job`**: `AnalysisConfigTFModel.Detectors` field
  type changed from `[]DetectorTFModel` to `types.List`; all call sites updated; new
  acceptance test exercises variable-sourced `detectors`.

## Impact

- **Users**: Regression introduced in v0.15.0 is fixed; configurations that assign
  `analysis_config.detectors` from a variable or module input now plan successfully.
  No HCL schema changes; no state migration required (the schema type exposed to
  Terraform users is unchanged — `schema.ListNestedAttribute` was and remains correct).
- **Code**: `internal/elasticsearch/ml/anomalydetectionjob/models_tf.go`,
  `internal/elasticsearch/ml/anomalydetectionjob/validate.go`,
  `internal/elasticsearch/ml/anomalydetectionjob/acc_test.go` (and associated testdata
  if applicable).
- **Maintenance**: `Detectors types.List` is now consistent with `CategorizationFilters`
  and `Influencers` in the same struct, reducing the risk of the same failure class
  recurring in future plan-time decode paths.
