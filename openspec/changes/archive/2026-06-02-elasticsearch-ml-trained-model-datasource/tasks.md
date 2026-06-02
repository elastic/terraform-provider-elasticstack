## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-ml-trained-model-datasource --type change` (or `make check-openspec` after sync).
- [x] 1.2 Resolve open questions on `create_time` wire format and alias vs. canonical `model_id` semantics (see `design.md`); update delta spec accordingly.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/elasticsearch-ml-trained-model/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Create package `internal/elasticsearch/ml/trainedmodel/` with the following files: `data_source.go`, `read.go`, `models.go`, `schema.go`.
- [x] 2.2 In `models.go`, define `trainedModelData` struct with `tfsdk:"..."` field tags. `model_id` is a required string input; all remaining fields (`id`, `description`, `model_type`, `model_size_bytes`, `fully_defined`, `tags`, `create_time`, `created_by`, `version`, `platform_architecture`, `license_level`, `input_json`, `inference_config_json`, `metadata_json`, `default_field_map`) are computed outputs.
- [x] 2.3 In `schema.go`, define the schema factory: `model_id` as required string input; all other fields as computed. Do not add an `elasticsearch_connection` block — it is injected by the entitycore envelope.
- [x] 2.4 In `read.go`, implement the read callback: call `client.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)`, handle 404 / empty-results by returning no error diagnostics and setting `id` to `""` with computed attributes null, map `TrainedModelConfig` fields to the TF model, marshal struct fields (`Input`, `InferenceConfig`, `Metadata`) to JSON strings.
- [x] 2.5 In `data_source.go`, register the data source via `entitycore.NewElasticsearchDataSource` (or equivalent envelope constructor), using the schema from `schema.go` and read callback from `read.go`.
- [x] 2.6 Add a client helper (or inline the call) in `internal/clients/elasticsearch/` for `GetTrainedModel(ctx, modelID string) (*types.TrainedModelConfig, bool, error)` — returns `(config, found, err)`.
- [x] 2.7 Register the new data source in `provider/plugin_framework.go` in the `dataSources()` list.
- [x] 2.8 Generate or update provider documentation for the new data source.

## 3. Testing

- [x] 3.1 Add acceptance test `TestAccDataSourceMLTrainedModel_basic` in `internal/elasticsearch/ml/trainedmodel/acc_test.go`: pre-condition on an existing trained model (use a built-in model such as `lang_ident_model_current` if available in the test cluster); read it via the data source; assert computed fields are populated.
- [x] 3.2 Add acceptance test `TestAccDataSourceMLTrainedModel_notFound`: attempt to read a non-existent `model_id`; assert the data source returns not-found (empty result, no error, or a plan-time error with a clear message depending on the chosen not-found policy).
- [ ] 3.3 (Optional) Add acceptance test `TestAccDataSourceMLTrainedModel_alias`: if the test cluster has a model alias, read by alias and verify computed fields match the underlying model.
- [x] 3.4 Gate acceptance tests behind a skip function if no suitable trained model is available in the cluster (e.g. check for `lang_ident_model_current` and skip if absent).
- [x] 3.5 Add unit tests for the `TrainedModelConfig` → TF model mapping in `models_test.go`: cover JSON marshaling of `input_json`, `inference_config_json`, `metadata_json`; nil fields mapping to null; `default_field_map` round-trip.
