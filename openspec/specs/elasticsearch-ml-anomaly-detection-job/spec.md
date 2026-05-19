# `elasticstack_elasticsearch_ml_anomaly_detection_job` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/anomalydetectionjob`

## Purpose

Define schema and behavior for the Elasticsearch ML anomaly detection job resource: API usage, identity and import, connection, lifecycle (force-new attributes), create/read/update/delete flows (including job close before delete), and mapping between Terraform configuration and the Elasticsearch Machine Learning Jobs API, including state consistency when Elasticsearch normalizes or omits fields on read.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "example" {
  id     = <computed, string>  # internal identifier: <cluster_uuid>/<job_id>
  job_id = <required, string>  # force new; 1–64 chars; lowercase alphanumeric, hyphens, underscores; must start and end with alphanumeric

  description = <optional, string>  # minimum 1 char
  groups      = <optional, set(string)>

  analysis_config {                   # required, force new
    bucket_span                = <optional+computed, string>  # default: "5m"; must match /^\d+[nsumdh]$/
    categorization_field_name  = <optional, string>
    categorization_filters     = <optional, list(string)>

    detectors {                       # required, min 1 element
      function             = <required, string>  # one of: count, high_count, low_count, non_zero_count, high_non_zero_count, low_non_zero_count, distinct_count, high_distinct_count, low_distinct_count, info_content, high_info_content, low_info_content, min, max, median, high_median, low_median, mean, high_mean, low_mean, metric, varp, high_varp, low_varp, sum, high_sum, low_sum, non_null_sum, high_non_null_sum, low_non_null_sum, rare, freq_rare, time_of_day, time_of_week, lat_long
      field_name           = <optional, string>
      by_field_name        = <optional, string>
      over_field_name      = <optional, string>
      partition_field_name = <optional, string>
      detector_description = <optional, string>
      exclude_frequent     = <optional, string>  # one of: all, none, by, over
      use_null             = <optional+computed, bool>  # default: false

      custom_rules {                  # optional
        actions    = <optional, list(string)>  # values: skip_result, skip_model_update
        conditions {                  # optional
          applies_to = <required, string>  # one of: actual, typical, diff_from_typical, time
          operator   = <required, string>  # one of: gt, gte, lt, lte
          value      = <required, float64>
        }
        scope = <optional, map of scope entries>  # map keys are analysis field names (e.g. by_field_name, over_field_name, partition_field_name); min 1 entry; no null keys; each entry has:
          filter_id   = <required, string>  # minimum 1 char; references an existing ML filter
          filter_type = <optional, string>  # one of: include, exclude
      }
    }

    influencers                  = <optional, list(string)>
    latency                      = <optional, string>
    model_prune_window           = <optional+computed, string>
    multivariate_by_fields       = <optional, bool>
    per_partition_categorization {    # optional
      enabled      = <optional+computed, bool>
      stop_on_warn = <optional, bool>
    }
    summary_count_field_name = <optional, string>
  }

  analysis_limits {                   # optional+computed
    categorization_examples_limit = <optional+computed, int64>  # >= 0
    model_memory_limit            = <optional, string>          # memory size (custom type)
  }

  data_description {                  # required, force new
    time_field  = <optional, string>
    time_format = <optional, string>
  }

  model_plot_config {                 # optional
    enabled             = <optional+computed, bool>
    annotations_enabled = <optional+computed, bool>
    terms               = <optional, string>
  }

  allow_lazy_open                           = <optional+computed, bool>
  background_persist_interval               = <optional, string>
  custom_settings                           = <optional, string>  # JSON (normalized) string
  daily_model_snapshot_retention_after_days = <optional+computed, int64>  # >= 0
  model_snapshot_retention_days             = <optional+computed, int64>  # >= 0
  renormalization_window_days               = <optional, int64>   # >= 0
  results_index_name                        = <optional+computed, string>  # force new
  results_retention_days                    = <optional, int64>   # >= 0

  # Read-only computed attributes
  create_time      = <computed, string>
  job_type         = <computed, string>
  job_version      = <computed, string>
  model_snapshot_id = <computed, string>

  elasticsearch_connection {          # optional, deprecated
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    headers                  = <optional, map(string)>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    key_file                 = <optional, string>
    cert_data                = <optional, string>
    key_data                 = <optional, string>
  }
}
```
## Requirements
### Requirement: Anomaly Detection Job CRUD APIs (REQ-001–REQ-005)

The resource SHALL use the Elasticsearch Put Anomaly Detection Job API to create jobs ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-job.html)). The resource SHALL use the Elasticsearch Update Anomaly Detection Job API to update jobs ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-update-job.html)). The resource SHALL use the Elasticsearch Get Anomaly Detection Jobs API to read job definitions ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-job.html)). The resource SHALL use the Elasticsearch Close Anomaly Detection Job API before deleting a job ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-close-job.html)). The resource SHALL use the Elasticsearch Delete Anomaly Detection Job API to delete jobs ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-job.html)). When Elasticsearch returns a non-success status for any API call (except 404 on read), the resource SHALL surface the API error as a Terraform diagnostic.

#### Scenario: API failure on create

- GIVEN a non-success response from the Put Anomaly Detection Job API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

#### Scenario: API failure on update

- GIVEN a non-success response from the Update Anomaly Detection Job API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

#### Scenario: API failure on delete

- GIVEN a non-success response from the Delete Anomaly Detection Job API
- WHEN the provider handles the response
- THEN Terraform diagnostics SHALL include the error

### Requirement: Identity and import (REQ-006–REQ-008)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<job_id>`. During create, the resource SHALL derive `id` by calling `r.client.ID(ctx, jobID)` to obtain the cluster UUID and `job_id`, and SHALL set `id` in state after a successful Put Job call. The resource SHALL support import by accepting an `id` in the format `<cluster_uuid>/<job_id>`, parsing it with `clients.CompositeIDFromStrFw`, and persisting both `id` and `job_id` to state. When the import `id` format is invalid (not parseable as a composite id), the resource SHALL return an error diagnostic.

#### Scenario: Import with valid composite id

- GIVEN import with a valid `<cluster_uuid>/<job_id>` id
- WHEN import completes
- THEN `id` and `job_id` SHALL be stored in state and read SHALL populate all remaining attributes

#### Scenario: Import with invalid id format

- GIVEN import with an id that is not in `<cluster_uuid>/<job_id>` format
- WHEN import runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Lifecycle — force-new attributes (REQ-009–REQ-012)

Changing `job_id` SHALL require resource replacement. Changing `analysis_config` SHALL require resource replacement. Changing `data_description` SHALL require resource replacement. Changing `results_index_name` SHALL require resource replacement.

#### Scenario: job_id change triggers replacement

- GIVEN an existing anomaly detection job
- WHEN the `job_id` attribute is changed in configuration
- THEN Terraform SHALL plan a destroy-and-recreate (force new)

#### Scenario: analysis_config change triggers replacement

- GIVEN an existing anomaly detection job
- WHEN any attribute within `analysis_config` is changed
- THEN Terraform SHALL plan a destroy-and-recreate (force new)

### Requirement: Connection (REQ-013)

The resource SHALL resolve a `*clients.ElasticsearchScopedClient` from the provider client factory and call `GetESClient()` to perform Elasticsearch operations. When `elasticsearch_connection` is absent, the factory SHALL return a typed client built from provider-level defaults. When `elasticsearch_connection` is configured, the factory SHALL return a typed scoped client rebuilt from that connection for all API calls (create, read, update, delete).

#### Scenario: Resource-level client override

- GIVEN `elasticsearch_connection` is set with specific endpoints and credentials
- WHEN any API call is made
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Create and read-after-write (REQ-014–REQ-015)

On create, the resource SHALL call the Put Anomaly Detection Job API with a JSON body built from the plan, then SHALL call read to refresh state from the server response. If the job is not found after creation, the resource SHALL return an error diagnostic ("Failed to read created job").

#### Scenario: State refreshed after create

- GIVEN a successful Put Job API call
- WHEN create completes
- THEN the resource SHALL call read to populate state from the API response

#### Scenario: Job not found after creation

- GIVEN a successful Put Job API call followed by a not-found read response
- WHEN create runs
- THEN the resource SHALL return an error diagnostic

### Requirement: Update — partial update with only mutable fields (REQ-016–REQ-018)

On update, the envelope SHALL invoke the resource update callback with `WriteRequest[TFModel]` (where `req.Prior` is a non-nil pointer to the prior-state model), supplying planned and prior models plus raw Terraform config. The callback SHALL only send fields that have changed and that are permitted by the Elasticsearch Update Job API. The following fields MAY be updated in place: `description`, `groups`, `model_plot_config`, `analysis_limits` (specifically `model_memory_limit`), `allow_lazy_open`, `background_persist_interval`, `custom_settings`, `daily_model_snapshot_retention_after_days`, `model_snapshot_retention_days`, `renormalization_window_days`, and `results_retention_days`. When no updateable fields have changed, the callback SHALL log a warning and SHALL NOT call the Update Job API.

After the update callback returns successfully — including the case where no Update Job API call was made because no updatable fields changed — the envelope SHALL refresh the resource by reading the job back from the API via the shared read callback and SHALL persist state from that read result.

#### Scenario: No updateable fields changed

- GIVEN an update where only non-updateable fields (e.g. analysis_config) have changed according to plan
- WHEN update runs
- THEN the resource SHALL not call the Update Job API and SHALL log a warning
- AND the envelope SHALL still perform read-after-write to refresh state from the API

#### Scenario: Updateable fields changed

- GIVEN an update where description has changed
- WHEN update runs
- THEN the resource SHALL call the Update Job API with only the changed fields
- AND the envelope SHALL refresh state via read-after-write afterward

### Requirement: Read — not found handling (REQ-019–REQ-020)

On read, when the Get Anomaly Detection Jobs API returns HTTP 404, the resource SHALL remove itself from state. When the API returns an empty jobs list, the resource SHALL remove itself from state. When the API returns multiple jobs for a single job ID, the resource SHALL use the first result and SHALL emit a warning diagnostic.

#### Scenario: Job not found on refresh

- GIVEN a job that has been deleted outside of Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state without error

#### Scenario: Multiple jobs returned

- GIVEN a Get Jobs API response containing more than one job for a specific job ID
- WHEN read runs
- THEN the resource SHALL emit a warning and SHALL use the first job in the response

### Requirement: Delete — close before delete (REQ-021–REQ-022)

On delete, the resource SHALL first attempt to close the job by calling the Close Anomaly Detection Job API with `force=true` and `allow_no_match=true`. If the close call fails, the resource SHALL log a warning and continue. After the Close Job API call returns (whether it succeeded or failed), the resource SHALL poll the job's state via the Get Job Stats API until the job reports `closed` state or is no longer found, before calling the Delete Job API. This polling SHALL be bounded by the Terraform operation context (i.e. the delete timeout). If polling fails, the resource SHALL log a warning and continue to deletion.

The resource SHALL then call the Delete Anomaly Detection Job API. If the first delete attempt fails, the resource SHALL retry once with `force=true`. If the retry also fails, the error SHALL be surfaced as a Terraform diagnostic.

#### Scenario: Normal delete succeeds

- **WHEN** the job is `closed` (or not found) before delete is called
- **THEN** the resource SHALL call Delete Job and it SHALL succeed without `force`

#### Scenario: First delete fails — retry with force succeeds

- **WHEN** the initial Delete Job call fails (e.g. job is still open due to polling timeout)
- **THEN** the resource SHALL retry Delete Job with `force=true` and treat a success response as the job being deleted

#### Scenario: Both delete attempts fail

- **WHEN** both the initial Delete Job and the `force=true` retry fail
- **THEN** the resource SHALL surface the retry error as a Terraform diagnostic

#### Scenario: Delete called regardless of polling outcome

- **WHEN** the polling wait fails for any reason
- **THEN** the resource SHALL still call the Delete Job API (not skip it)

### Requirement: job_id validation (REQ-023)

The `job_id` attribute SHALL be validated to be between 1 and 64 characters, contain only lowercase alphanumeric characters (a–z and 0–9), hyphens, and underscores, and start and end with an alphanumeric character.

#### Scenario: Invalid job_id rejected

- GIVEN a `job_id` that starts with a hyphen or contains uppercase characters
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error and SHALL not call the API

### Requirement: analysis_config.detectors validation (REQ-024)

The `analysis_config.detectors` list SHALL contain at least one element. Each detector `function` SHALL be one of the enumerated values. Each detector `exclude_frequent` SHALL be one of: `all`, `none`, `by`, `over`. Each detector `custom_rules[*].actions` value SHALL be one of: `skip_result`, `skip_model_update`. Each detector `custom_rules[*].conditions[*].applies_to` SHALL be one of: `actual`, `typical`, `diff_from_typical`, `time`. Each detector `custom_rules[*].conditions[*].operator` SHALL be one of: `gt`, `gte`, `lt`, `lte`. Each detector `custom_rules[*].scope` entry SHALL have a `filter_id` that is at least 1 character. Each detector `custom_rules[*].scope[*].filter_type` SHALL be one of: `include`, `exclude`. Each detector `custom_rules` entry SHALL have either a non-empty `scope` or at least one `conditions` entry; both may be set simultaneously. When both `scope` and `conditions` are unknown at plan time, the validation SHALL be skipped (to avoid false positives).

#### Scenario: Empty detectors list rejected

- GIVEN an `analysis_config` with no detectors
- WHEN the configuration is applied
- THEN the provider SHALL return a validation error

### Requirement: Mapping — config to API model (REQ-025–REQ-026)

On create and update, optional fields that are null or unknown SHALL be omitted from the API request body. On create, the resource SHALL serialize `analysis_config.detectors[*].custom_rules[*].actions` as a JSON array of strings and SHALL serialize `analysis_config.detectors[*].custom_rules[*].conditions[*]` as objects containing `applies_to`, `operator`, and `value`. The resource SHALL serialize `analysis_config.detectors[*].custom_rules[*].scope` as a JSON object mapping analysis field names to objects containing `filter_id` (required) and `filter_type` (omitted when not configured or empty). The `custom_settings` field SHALL be validated as a JSON string and SHALL be decoded into a `map[string]any` for the API request. When `custom_settings` is not valid JSON, the resource SHALL return an error diagnostic and SHALL not call the API.

#### Scenario: Invalid custom_settings JSON

- GIVEN an invalid JSON string in `custom_settings`
- WHEN create or update runs
- THEN the provider SHALL return an error diagnostic and SHALL not call the Put Job or Update Job API

#### Scenario: Custom rules are sent on create

- GIVEN a detector with `custom_rules` containing `actions` and `conditions`
- WHEN create builds the Put Job request body
- THEN the request SHALL include those `custom_rules` entries with their configured values

#### Scenario: Custom rules with scope are sent on create

- GIVEN a detector with `custom_rules` containing `scope` entries
- WHEN create builds the Put Job request body
- THEN the request SHALL include the `scope` map with each entry containing `filter_id` and, when configured, `filter_type`

#### Scenario: Custom rules scope with omitted filter_type

- GIVEN a detector with `custom_rules` containing a `scope` entry where `filter_type` is not set
- WHEN create or update builds the API request body
- THEN the `filter_type` field SHALL be omitted from that scope entry's JSON object

### Requirement: Mapping — API response to state (REQ-027–REQ-031)

On read, the resource SHALL set the following state attributes from the Get Jobs API response:
- `job_id`, `description`, `job_type`, `job_version`, `create_time`, `model_snapshot_id` from the corresponding API fields.
- `groups` SHALL be set to null in state when the API returns an empty or nil groups list; otherwise it SHALL be set to the returned set of strings.
- `analysis_config.bucket_span`, `categorization_field_name`, `latency`, `model_prune_window`, `multivariate_by_fields`, and `summary_count_field_name` from the corresponding `analysis_config` API fields.
- `analysis_config.categorization_filters` SHALL use the API values when Elasticsearch returns a non-empty list. When Elasticsearch omits the list or returns it empty, the resource SHALL preserve the prior configured value so server-side normalization into `categorization_analyzer` does not create drift.
- `analysis_config.influencers` SHALL use the API values when Elasticsearch returns a non-empty list. When Elasticsearch omits the list or returns it empty, the resource SHALL preserve the prior configured value, including an explicit empty list.
- `analysis_config.detectors[*]` SHALL be set from the corresponding detector in the API response. When the prior detector configuration omitted `detector_description` and Elasticsearch returns an auto-generated description, the resource SHALL keep `detector_description` null in state instead of storing the generated value. `custom_rules[*].actions` and `custom_rules[*].conditions` SHALL be populated from the API response; when Elasticsearch omits an empty `actions` or `conditions` list, the resource SHALL preserve a previously configured empty list rather than converting it to null. `custom_rules[*].scope` SHALL be populated from the API response when non-empty; when Elasticsearch omits an empty `scope` map, the resource SHALL preserve the prior configured `scope` value rather than converting it to null. Within each scope entry, when the API returns an empty or absent `filter_type`, the resource SHALL store `filter_type` as null in state.
- `analysis_config.per_partition_categorization` SHALL be populated only when the block was previously configured or when Elasticsearch reports `enabled = true`. When the block exists in prior state and Elasticsearch omits `stop_on_warn`, the resource SHALL preserve the prior `stop_on_warn` value.
- Empty or nil string fields in the API response SHALL be stored as null in state (not as empty string), using `typeutils.NonEmptyStringishValue`.
- `results_index_name` SHALL be stored after stripping a `custom-` prefix from the API response value.
- `custom_settings` SHALL be JSON-marshaled from the API response `map[string]any` when non-nil; when nil it SHALL be stored as null in state.

#### Scenario: Empty description stored as null

- GIVEN a job where description is empty string on the server
- WHEN read runs
- THEN `description` SHALL be null (not empty string) in state

#### Scenario: custom_settings nil stored as null

- GIVEN a job where custom_settings is not set on the server
- WHEN read runs
- THEN `custom_settings` SHALL be null (not empty string) in state

#### Scenario: Empty influencers list remains empty

- GIVEN configuration that sets `analysis_config.influencers = []`
- WHEN read runs and Elasticsearch returns no influencers
- THEN `analysis_config.influencers` SHALL remain an empty list in state

#### Scenario: Categorization filters survive Elasticsearch normalization

- GIVEN configuration that sets `analysis_config.categorization_filters`
- WHEN read runs and Elasticsearch does not return `categorization_filters` because it normalized them internally
- THEN the prior configured `analysis_config.categorization_filters` SHALL remain in state

#### Scenario: Auto-generated detector description does not create drift

- GIVEN a detector without `detector_description` in configuration
- WHEN read runs and Elasticsearch returns an auto-generated detector description
- THEN `analysis_config.detectors[*].detector_description` SHALL remain null in state

#### Scenario: Custom rule conditions round-trip from API to state

- GIVEN a detector with `custom_rules` containing conditions
- WHEN create succeeds and read refreshes state
- THEN the configured `actions` and `conditions` SHALL be present in state

#### Scenario: Custom rule scope round-trip from API to state

- GIVEN a detector with `custom_rules` containing `scope` entries referencing ML filters
- WHEN create succeeds and read refreshes state
- THEN the configured `scope` entries SHALL be present in state with correct `filter_id` and `filter_type` values

#### Scenario: Custom rule scope with absent filter_type stored as null

- GIVEN a detector with `custom_rules` containing `scope` entries where `filter_type` is not configured
- WHEN read runs and the API omits `filter_type`
- THEN `filter_type` SHALL be null in state (not empty string)

#### Scenario: Custom rule with empty scope preserved from prior state

- GIVEN a detector with `custom_rules` where the API returns an empty `scope` map
- WHEN read runs and the prior state has a configured `scope` value
- THEN the prior configured `scope` value SHALL be preserved in state

#### Scenario: Custom rule with both scope and conditions accepted

- GIVEN a detector with `custom_rules` containing both a non-empty `scope` and at least one `conditions` entry
- WHEN create runs
- THEN the provider SHALL send both `scope` and `conditions` to the Elasticsearch API and SHALL not reject the configuration

#### Scenario: Custom rule with neither scope nor conditions rejected

- GIVEN a detector with `custom_rules` containing an entry with no `scope` and no `conditions`
- WHEN the configuration is validated
- THEN the provider SHALL return a validation error indicating a rule must have either a non-empty `scope` or at least one condition

#### Scenario: Disabled per-partition categorization preserves configured stop_on_warn

- GIVEN configuration that sets `analysis_config.per_partition_categorization.enabled = false` and `stop_on_warn = false`
- WHEN read runs and Elasticsearch omits `stop_on_warn`
- THEN the resource SHALL keep the configured `stop_on_warn` value in state

#### Scenario: results_index_name strips custom- prefix

- GIVEN a job where the API returns `results_index_name = "custom-my-index"`
- WHEN read runs
- THEN `results_index_name` in state SHALL be `"my-index"`

### Requirement: Plan/State — UseStateForUnknown (REQ-032)

The following attributes SHALL use `UseStateForUnknown` plan modifier to preserve prior state values when the plan value is unknown: `id`, `analysis_config.bucket_span`, `analysis_config.detectors[*].use_null`, `analysis_config.model_prune_window`, `analysis_config.per_partition_categorization.enabled`, `analysis_limits`, `model_plot_config.enabled`, `model_plot_config.annotations_enabled`, `allow_lazy_open`, `daily_model_snapshot_retention_after_days`, `model_snapshot_retention_days`, `results_index_name`, `create_time`, `job_type`, `job_version`, `model_snapshot_id`.

#### Scenario: id preserved across plan

- GIVEN an existing job with a known id in state
- WHEN a plan is generated without changing job_id
- THEN `id` SHALL remain known (not unknown) in the plan

### Requirement: `analysis_config.detectors` internal model type (REQ-033)

The Go field `AnalysisConfigTFModel.Detectors` SHALL be typed `types.List` (not
`[]DetectorTFModel`). The `types.List` type is the Plugin Framework-idiomatic holder
for list attributes that can carry null/unknown state, consistent with
`AnalysisConfigTFModel.CategorizationFilters` and `AnalysisConfigTFModel.Influencers`
in the same struct.

All code that reads `Detectors` as a Go slice SHALL use `ElementsAs(ctx,
&detectorSlice, false)` to convert the `types.List` value. All code that writes
`Detectors` from a `[]DetectorTFModel` slice SHALL use
`types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getDetectorAttrTypes(ctx)},
slice)`.

The Terraform schema declaration (`schema.ListNestedAttribute` for
`analysis_config.detectors`) SHALL NOT change. This requirement applies only to the
internal Go model field type.

#### Scenario: Variable-sourced detectors plan succeeds

- GIVEN a Terraform configuration that assigns `analysis_config.detectors` from a
  `variable` of type `list(object({...}))` (including when the variable has a concrete
  default value that Terraform marks as potentially-unknown at the config phase)
- WHEN `terraform plan` runs
- THEN the provider SHALL NOT return a `Value Conversion Error` and the plan SHALL
  succeed

#### Scenario: Hardcoded detectors continue to work

- GIVEN a Terraform configuration with `analysis_config.detectors` defined as a
  hardcoded list literal in the resource block
- WHEN `terraform plan` and `terraform apply` run
- THEN all existing create, update, read, and delete behaviors SHALL be preserved
  without regression

### Requirement: `ValidateConfig` plan-time unknown guard for detectors (REQ-034)

The `ValidateConfig` implementation SHALL handle the case where
`analysis_config.detectors` is unknown at plan time (for example, when the value flows
from a module input or variable). Specifically:

- When `Detectors.IsUnknown()` is true, validation of `custom_rules` SHALL be skipped
  without returning an error diagnostic. Validation is deferred to apply time when all
  values are concrete.
- When `Detectors` is known and non-null, validation of `custom_rules` SHALL proceed as
  before: each rule that has neither a non-empty `scope` nor at least one `conditions`
  entry SHALL produce an attribute-level error diagnostic.

#### Scenario: Unknown detectors skips custom_rules validation

- GIVEN an `analysis_config.detectors` that is unknown at plan time
- WHEN `ValidateConfig` runs
- THEN no `Invalid detector "custom_rules" entry` error SHALL be emitted and the plan
  SHALL proceed

#### Scenario: Known detectors with empty custom_rules still fails validation

- GIVEN a known `analysis_config.detectors` containing a custom rule with neither a
  `scope` nor any `conditions`
- WHEN `ValidateConfig` runs
- THEN the resource SHALL return an error diagnostic identifying the offending custom
  rule

### Requirement: Acceptance test for variable-sourced detectors (REQ-035)

The acceptance test suite SHALL include at least one test case for
`elasticstack_elasticsearch_ml_anomaly_detection_job` that exercises assigning
`analysis_config { detectors = var.detectors }` where `var.detectors` is a Terraform
`variable` of type `list(object({...}))`. The test SHALL use the minimal repro shape
from issue #2966 (`function` as required string; `field_name`, `by_field_name`,
`detector_description` as optional string). The test SHALL assert that plan and apply
complete without a `Value Conversion Error` and SHALL verify at least one detector
attribute in state (e.g. `analysis_config.detectors.0.function`). The test SHALL be
named consistently with the existing `TestAccResourceAnomalyDetectionJob*` convention.

#### Scenario: Acceptance test — variable-sourced detectors plan and apply

- GIVEN an acceptance test configuration that assigns `analysis_config.detectors` from
  a Terraform `variable` with a concrete default
- WHEN the acceptance test runs `terraform plan` and `terraform apply`
- THEN the resource SHALL be created without a `Value Conversion Error` and state SHALL
  reflect the expected detector function value

