## Why

Since v0.15.0, `elasticstack_elasticsearch_ml_anomaly_detection_job` fails at plan time with a
`Value Conversion Error` when `analysis_config` is assigned from a dynamic source such as
`for_each` / `each.value` or a Terraform variable typed as `object({...})`:

```
Received unknown value, however the target type cannot handle unknown values.
Path: analysis_config
Target Type: *anomalydetectionjob.AnalysisConfigTFModel
Suggested Type: basetypes.ObjectValue
```

The Plugin Framework is explicitly reporting that `*AnalysisConfigTFModel` (a pointer to a plain Go
struct) cannot hold an **unknown** value during plan. The same pattern was already fixed one level
deeper in commit `bc23e6e2` (issue #2966), where `Detectors []DetectorTFModel` was changed to
`types.List`. This change applies the same fix to `analysis_config` and, proactively, to
`per_partition_categorization` — the last remaining struct-pointer in the `analysis_config` subtree.

## What Changes

- Change `TFModel.AnalysisConfig` from `*AnalysisConfigTFModel` to `types.Object` so the top-level
  `analysis_config` attribute can hold null, unknown, and known values.
- Change `AnalysisConfigTFModel.PerPartitionCategorization` from `*PerPartitionCategorizationTFModel`
  to `types.Object`, eliminating the only remaining struct-pointer limitation in the subtree.
- Update `toAPIModel()`, `convertAnalysisConfigFromAPI()`, and `validateConfigCustomRules()` to use
  the `.As()` / `.IsNull()` / `.IsUnknown()` framework-native patterns.
- Add a `getPerPartitionCategorizationAttrTypes(ctx)` schema helper derived from the existing
  `getAnalysisConfigAttrTypes(ctx)` pattern.
- Add a regression acceptance test
  `TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig` that assigns the full
  `analysis_config` block from a Terraform variable, mirroring the existing
  `TestAccResourceAnomalyDetectionJobVariableSourcedDetectors` test.

## Capabilities

### Modified Capabilities
- `elasticstack_elasticsearch_ml_anomaly_detection_job`: fix Value Conversion Error for variable-sourced `analysis_config`

## Impact

- Internal Go representation of `TFModel.AnalysisConfig` and
  `AnalysisConfigTFModel.PerPartitionCategorization` changes from struct-pointer to `types.Object`.
- No change to the Terraform schema visible to users.
- No change to the Elastic ML API surface.
- Existing acceptance tests continue to pass; a new regression test is added.
