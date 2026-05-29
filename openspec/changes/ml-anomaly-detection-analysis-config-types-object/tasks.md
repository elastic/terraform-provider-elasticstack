## 1. Add schema helper and update TFModel fields

- [ ] 1.1 Add `getPerPartitionCategorizationAttrTypes(ctx context.Context) map[string]attr.Type`
  to `schema.go`, derived from the `per_partition_categorization` nested attribute inside
  `getAnalysisConfigAttrTypes(ctx)`.

- [ ] 1.2 In `models_tf.go`, change `TFModel.AnalysisConfig` from `*AnalysisConfigTFModel` to
  `types.Object`.

- [ ] 1.3 In `models_tf.go`, change `AnalysisConfigTFModel.PerPartitionCategorization` from
  `*PerPartitionCategorizationTFModel` to `types.Object`.

## 2. Update toAPIModel

- [ ] 2.1 In `toAPIModel()`, replace the `plan.AnalysisConfig == nil` guard with
  `plan.AnalysisConfig.IsNull() || plan.AnalysisConfig.IsUnknown()` and add a diagnostic error
  message for the null/unknown case (analysis_config is required).

- [ ] 2.2 In `toAPIModel()`, extract `AnalysisConfigTFModel` from `plan.AnalysisConfig` via
  `plan.AnalysisConfig.As(ctx, &analysisConfigTF, basetypes.ObjectAsOptions{})`.

- [ ] 2.3 In `toAPIModel()`, replace the `analysisConfig.PerPartitionCategorization != nil` guard
  with `!analysisConfig.PerPartitionCategorization.IsNull() && !analysisConfig.PerPartitionCategorization.IsUnknown()`
  and extract `PerPartitionCategorizationTFModel` via `.As()`.

## 3. Update convertAnalysisConfigFromAPI

- [ ] 3.1 Change the return type of `convertAnalysisConfigFromAPI` from `*AnalysisConfigTFModel`
  to `(types.Object, diag.Diagnostics)` and update the call site in `fromAPIModel()`.

- [ ] 3.2 In `convertAnalysisConfigFromAPI()`, replace the `apiConfig == nil || apiConfig.BucketSpan == ""`
  early return with a `types.ObjectNull(getAnalysisConfigAttrTypes(ctx))` return.

- [ ] 3.3 In `convertAnalysisConfigFromAPI()`, replace the `if plan.AnalysisConfig != nil` prior-state
  extraction with `plan.AnalysisConfig.As(ctx, &analysisConfigTF, basetypes.ObjectAsOptions{})`,
  guarded by `!plan.AnalysisConfig.IsNull() && !plan.AnalysisConfig.IsUnknown()`.

- [ ] 3.4 In `convertAnalysisConfigFromAPI()`, replace `PerPartitionCategorization` pointer assignment
  with a `types.ObjectValueFrom(ctx, getPerPartitionCategorizationAttrTypes(ctx), perPartitionTF)`
  call, and use `types.ObjectNull(getPerPartitionCategorizationAttrTypes(ctx))` when not present.

- [ ] 3.5 In `convertAnalysisConfigFromAPI()`, wrap the final return in
  `types.ObjectValueFrom(ctx, getAnalysisConfigAttrTypes(ctx), analysisConfigTF)` instead of
  returning a `*AnalysisConfigTFModel` pointer.

## 4. Update validate.go

- [ ] 4.1 In `validateConfigCustomRules()`, replace the `config.AnalysisConfig == nil` guard with
  `config.AnalysisConfig.IsNull() || config.AnalysisConfig.IsUnknown()`.

- [ ] 4.2 In `validateConfigCustomRules()`, extract `AnalysisConfigTFModel` from `config.AnalysisConfig`
  via `.As()` before accessing `.Detectors`, propagating any diagnostics.

## 5. Add regression acceptance test

- [ ] 5.1 Add testdata config at
  `internal/elasticsearch/ml/anomalydetectionjob/testdata/TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig/create/anomaly_detection.tf`
  that declares an `analysis_config` variable typed as `object({...})` matching the full
  `analysis_config` schema shape and assigns `analysis_config = var.analysis_config`.

- [ ] 5.2 Add `TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig` to `acc_test.go`
  with a create step (plan+apply) and a plan-after-apply step, mirroring
  `TestAccResourceAnomalyDetectionJobVariableSourcedDetectors`.

## 6. Validate and verify

- [ ] 6.1 Run `make build` and confirm the provider compiles without errors.

- [ ] 6.2 Run `go test ./internal/elasticsearch/ml/anomalydetectionjob/... -run TestAcc -v` with a
  running Elasticsearch instance to confirm the new regression test passes and no existing tests
  regress.
