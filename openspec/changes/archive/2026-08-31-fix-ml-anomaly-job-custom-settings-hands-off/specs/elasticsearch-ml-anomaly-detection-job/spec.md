## MODIFIED Requirements

### Requirement: Mapping — config to API model (REQ-025–REQ-026)

On create and update, optional fields that are null or unknown SHALL be omitted from the API request body. On create, the resource SHALL serialize `analysis_config.detectors[*].custom_rules[*].actions` as a JSON array of strings and SHALL serialize `analysis_config.detectors[*].custom_rules[*].conditions[*]` as objects containing `applies_to`, `operator`, and `value`. The resource SHALL serialize `analysis_config.detectors[*].custom_rules[*].scope` as a JSON object mapping analysis field names to objects containing `filter_id` (required) and `filter_type` (omitted when not configured or empty). The `custom_settings` field SHALL be validated as a JSON string and SHALL be decoded into a `map[string]any` for the API request. When `custom_settings` is not valid JSON, the resource SHALL return an error diagnostic and SHALL not call the API.

`custom_settings` is atomic and hands-off when omitted: when `custom_settings` is null in the plan, the resource SHALL NOT send `custom_settings` in the create or update request body (this already follows from the general "null or unknown fields are omitted" rule above, and is restated here because it is the load-bearing behavior for this attribute). When `custom_settings` decodes to a JSON object with zero keys (i.e. the config is exactly `"{}"`), the resource SHALL still send `custom_settings` as an explicit empty JSON object (`{"custom_settings":{}}`) in the update request body — this is the only way to clear an existing `custom_settings` bag on the server, and the wire encoding SHALL NOT rely on Go's `encoding/json` `omitempty` behavior for this field on the update path, since `omitempty` drops a zero-length map and would silently turn the wipe into a no-op. When `custom_settings` decodes to a JSON object with one or more keys, the resource SHALL send exactly that object, replacing the server's entire `custom_settings` bag (the Update Job API replaces this field; it does not merge), including any keys previously written by Kibana or other tooling that are absent from the configured object.

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

#### Scenario: Omitted custom_settings is never sent

- GIVEN a configuration where `custom_settings` is omitted (null)
- WHEN create or update builds the API request body
- THEN `custom_settings` SHALL NOT appear in the request body, regardless of any `custom_settings` value already present on the server

#### Scenario: custom_settings = "{}" sends an explicit empty object on update

- GIVEN an update where `custom_settings` changes to `"{}"` from a prior non-empty value
- WHEN update builds the Update Job request body
- THEN the request body SHALL include `"custom_settings":{}` (not omit the field)

#### Scenario: Non-empty custom_settings replaces the whole bag

- GIVEN a job whose server-side `custom_settings` contains keys not present in the configured `custom_settings` object (for example a Kibana-authored `created_by` key)
- WHEN update sends the configured `custom_settings` object
- THEN the request body SHALL contain exactly the configured object and the server-side extra keys SHALL be dropped by the API

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
- `custom_settings` SHALL be governed by the atomic hands-off contract below rather than being mirrored unconditionally from the API response.

`custom_settings` read contract: the resource SHALL determine the incoming/prior `custom_settings` value — the plan's configured value on the create/update read-after-write path, or the previously persisted state value on a plain read/refresh (including immediately after `terraform import`) — before applying the API response, and SHALL do the following based on that incoming value, irrespective of what the Get Jobs API returns for `custom_settings`:
- When the incoming value is null, the resource SHALL store `custom_settings` as null in state and SHALL NOT copy any API-returned `custom_settings` value into state.
- When the incoming value is a JSON object with zero keys (`"{}"`), the resource SHALL store `custom_settings` as `"{}"` in state, even when the Get Jobs API omits `custom_settings` entirely or returns it as absent/empty for that job.
- When the incoming value is a JSON object with one or more keys, the resource SHALL store `custom_settings` as the JSON-marshaled form of the API response's `custom_settings` map when the API returns a non-nil value for it (this may include keys not present in the incoming value, surfacing drift for the next plan); when the API returns no value at all for a previously-owned bag, the resource SHALL store `"{}"` in state rather than re-using the incoming value, since the server no longer holds it.

#### Scenario: Empty description stored as null

- GIVEN a job where description is empty string on the server
- WHEN read runs
- THEN `description` SHALL be null (not empty string) in state

#### Scenario: custom_settings omitted from config is never populated from the server

- GIVEN a job whose server-side `custom_settings` is populated (for example `{"created_by":"advanced-wizard"}` or `{"custom_urls":[]}`) by Kibana or another tool outside Terraform
- AND the Terraform configuration omits `custom_settings` (null)
- WHEN create, update, or plain read runs and the envelope refreshes state from the Get Jobs API
- THEN `custom_settings` SHALL remain null in state and the apply SHALL NOT fail with "Provider produced inconsistent result after apply"

#### Scenario: custom_settings stays null through terraform import

- GIVEN a job whose server-side `custom_settings` is populated outside Terraform
- AND the Terraform configuration used to manage the imported resource omits `custom_settings`
- WHEN `terraform import` runs followed by the subsequent read that populates the rest of state
- THEN `custom_settings` SHALL be null in state, matching the omitted configuration

#### Scenario: custom_settings "{}" persists even when the API omits the field

- GIVEN a prior `custom_settings` value of `"{}"` (a previously-applied wipe)
- WHEN read runs and the Get Jobs API response omits `custom_settings` entirely for that job
- THEN `custom_settings` SHALL remain `"{}"` in state (not null and not converted to any other value)

#### Scenario: custom_settings nil stored as null

- GIVEN a job where custom_settings is not set on the server
- AND the Terraform configuration also has `custom_settings` unset
- WHEN read runs
- THEN `custom_settings` SHALL be null (not empty string) in state

#### Scenario: Owned custom_settings shows drift when the server adds keys

- GIVEN a job with configured `custom_settings = jsonencode({department = "ops"})`, applied successfully
- AND Elasticsearch/Kibana subsequently adds a key to the server-side `custom_settings` (for example `created_by`) outside Terraform
- WHEN the next `terraform plan` runs
- THEN the plan SHALL show a diff for `custom_settings` reflecting the server's superset value, and the next `terraform apply` SHALL replace the server-side bag with exactly the configured object

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
