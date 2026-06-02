# elasticsearch-ml-trained-model Specification

## Purpose
TBD - created by archiving change elasticsearch-ml-trained-model-datasource. Update Purpose after archive.
## Requirements
### Requirement: API — Read (REQ-001)

The data source SHALL call `GET _ml/trained_models/<model_id>` via the typed Elasticsearch client (`client.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)`) to retrieve model configuration. The request SHALL use `include=definition_status` (or equivalent) if needed for the `fully_defined` field, but SHALL NOT request the full `definition` or `compressed_definition` to keep the response small.

Required cluster privilege: `monitor_ml`.

When the Elasticsearch API returns HTTP 404, or when the response `trained_model_configs` array is empty, the data source SHALL return no error diagnostics and set `id` to an empty string with all computed attributes null. In all other API error cases, the data source SHALL surface the error.

#### Scenario: Read an existing trained model

- GIVEN a trained model with `model_id = "lang_ident_model_current"` exists in the cluster
- WHEN the data source is read
- THEN `GET _ml/trained_models/lang_ident_model_current` is called
- AND computed attributes are populated from the API response
- AND `id` is set to `"<cluster_uuid>/lang_ident_model_current"`

#### Scenario: Model not found (404)

- GIVEN no trained model with the specified `model_id` exists
- WHEN the data source is read
- THEN the data source returns no error diagnostics
- AND `id` is set to an empty string and all computed attributes are null

#### Scenario: Empty results array

- GIVEN the API returns an empty `trained_model_configs` array
- WHEN the data source is read
- THEN the data source returns no error diagnostics
- AND `id` is set to an empty string and all computed attributes are null

### Requirement: Identity (REQ-002)

The `model_id` attribute SHALL be a required string input. It accepts the canonical model ID or a model alias.

The computed `id` attribute SHALL be set to `"<cluster_uuid>/<model_id>"` using the provider's standard composite ID helper (`client.ID(ctx, modelID)`), where `model_id` is the value supplied by the practitioner (not the API-resolved canonical ID).

#### Scenario: Composite id is set after read

- GIVEN a data source configuration with `model_id = "my-model"`
- WHEN the data source is successfully read
- THEN `id` is set to `"<cluster_uuid>/my-model"` in state

### Requirement: Scalar field mapping (REQ-003)

After a successful API read, the data source SHALL map `TrainedModelConfig` fields to Terraform state as follows:

| Terraform attribute    | Source field                           | Null when                        |
|------------------------|----------------------------------------|----------------------------------|
| `description`          | `TrainedModelConfig.Description`       | field is absent or empty string  |
| `model_type`           | `TrainedModelConfig.ModelType`         | field is absent                  |
| `model_size_bytes`     | `TrainedModelConfig.ModelSizeBytes`    | field is absent or zero          |
| `fully_defined`        | `TrainedModelConfig.FullyDefined`      | field is absent (default false)  |
| `create_time`          | `TrainedModelConfig.CreateTime`        | field is absent                  |
| `created_by`           | `TrainedModelConfig.CreatedBy`         | field is absent                  |
| `version`              | `TrainedModelConfig.Version`           | field is absent                  |
| `platform_architecture`| `TrainedModelConfig.PlatformArchitecture` | field is absent               |
| `license_level`        | `TrainedModelConfig.LicenseLevel`      | field is absent                  |

The `create_time` value SHALL be stored as an ISO-8601 string. If the API returns an epoch-millisecond integer, it SHALL be converted to ISO-8601 UTC before storing.

#### Scenario: All scalar fields populated

- GIVEN the API response contains all scalar fields
- WHEN the data source maps the response
- THEN each Terraform attribute reflects the corresponding API value

#### Scenario: Optional scalar field absent

- GIVEN the API response omits `description`
- WHEN the data source maps the response
- THEN `description` is null in state

### Requirement: Collection field mapping (REQ-004)

The `tags` attribute SHALL be mapped as a set of strings from `TrainedModelConfig.Tags`. When the API returns an empty or absent tags list, `tags` SHALL be an empty set (not null) if the field is present in the API response, or null if the field is absent.

The `default_field_map` attribute SHALL be mapped as a `map(string)` from `TrainedModelConfig.DefaultFieldMap`. When the API returns an empty or absent map, `default_field_map` SHALL be null.

#### Scenario: Tags populated

- GIVEN the API response contains `tags: ["nlp", "text"]`
- WHEN the data source maps the response
- THEN `tags` contains exactly `{"nlp", "text"}`

#### Scenario: Tags absent

- GIVEN the API response omits the `tags` field
- WHEN the data source maps the response
- THEN `tags` is null

### Requirement: JSON computed fields (REQ-005)

The `input_json`, `inference_config_json`, and `metadata_json` attributes SHALL each serialize the corresponding API struct field to a compact JSON string:

- `input_json`: JSON of `TrainedModelConfig.Input` (the model's input definition, including field names).
- `inference_config_json`: JSON of `TrainedModelConfig.InferenceConfig` (the default inference configuration union).
- `metadata_json`: JSON of `TrainedModelConfig.Metadata` (arbitrary model metadata map).

When the corresponding source field is nil or absent, the attribute SHALL be null in state.

The JSON serialization SHALL produce a stable, compact (no extra whitespace) representation. Key order need not be guaranteed but SHALL be deterministic within a single read.

#### Scenario: input_json populated

- GIVEN the API response contains a non-nil `Input` field with field names `["text_field"]`
- WHEN the data source maps the response
- THEN `input_json` is a non-empty JSON string encoding the input definition

#### Scenario: inference_config_json null when absent

- GIVEN the API response has a nil `InferenceConfig`
- WHEN the data source maps the response
- THEN `inference_config_json` is null in state

#### Scenario: metadata_json populated

- GIVEN the API response contains a non-nil metadata map
- WHEN the data source maps the response
- THEN `metadata_json` is a valid JSON object string

### Requirement: Acceptance test prerequisites (REQ-006)

Acceptance tests SHALL be gated with a skip function that checks for the presence of a suitable trained model (e.g. `lang_ident_model_current`) in the test cluster. If no model is available, the test SHALL be skipped with a clear message rather than failing.

Tests that verify the not-found path SHALL use a `model_id` that is guaranteed not to exist (e.g. a UUID-based name).

#### Scenario: Skip when no model available

- GIVEN no trained model exists in the test cluster
- WHEN the acceptance test is run
- THEN the test is skipped (not failed) with a message indicating the prerequisite is missing

