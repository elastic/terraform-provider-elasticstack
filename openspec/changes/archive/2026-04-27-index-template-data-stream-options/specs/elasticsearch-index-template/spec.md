# Delta Spec: `elasticstack_elasticsearch_index_template` — `data_stream_options` Support

Base spec: `openspec/specs/elasticsearch-index-template/spec.md`
Last requirement in base spec: REQ-031
This delta introduces: REQ-032 to REQ-037

---

This delta defines the target behavior introduced by change `index-template-data-stream-options`. It describes the requirements the implementation will satisfy once the change is applied; it is not intended to claim that the current implementation already meets these behaviors.

## ADDED Requirements

### Requirement: Schema — `template.data_stream_options` block (REQ-032)

The `template` block SHALL support an optional `data_stream_options` sub-block. The `data_stream_options` block SHALL contain at most one optional `failure_store` sub-block. If `data_stream_options` is configured without a `failure_store` sub-block, the provider SHALL reject the configuration at plan time with an error diagnostic. The `failure_store` block SHALL contain:

- `enabled` — required boolean; activates or deactivates document redirection to the failure store on newly created matching data streams.
- `lifecycle` — optional sub-block containing `data_retention`, a required string specifying how long failure store documents are retained (e.g. `"30d"`).

When `data_stream_options` is omitted from the configuration, the provider SHALL not include the field in API requests and SHALL leave it unset in Terraform state.

**Example HCL:**

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  name           = "my-index-template"
  index_patterns = ["my-datastream-*"]

  template {
    data_stream_options {
      failure_store {
        enabled = true
        lifecycle {
          data_retention = "30d"
        }
      }
    }
  }

  data_stream {}
}
```

#### Scenario: `data_stream_options` omitted

- **WHEN** `data_stream_options` is not configured
- **THEN** the provider SHALL not include `data_stream_options` in the Put index template API request body

#### Scenario: `failure_store.enabled` without lifecycle

- **WHEN** `failure_store.enabled = true` is configured and `lifecycle` is omitted
- **THEN** the provider SHALL send `{"failure_store": {"enabled": true}}` inside `template.data_stream_options` in the API request

#### Scenario: `failure_store` with lifecycle retention

- **WHEN** `failure_store.enabled = true` and `failure_store.lifecycle.data_retention = "10d"` are configured
- **THEN** the provider SHALL send `{"failure_store": {"enabled": true, "lifecycle": {"data_retention": "10d"}}}` inside `template.data_stream_options`

---

### Requirement: Compatibility — version gate for `data_stream_options` (REQ-033)

When `data_stream_options` is configured and the Elasticsearch server version is below `9.1.0`, the provider SHALL return an error diagnostic and SHALL not call the Put index template API.

#### Scenario: Feature on unsupported cluster version

- **GIVEN** `data_stream_options` is configured
- **AND** the connected Elasticsearch server version is below `9.1.0`
- **WHEN** create or update runs
- **THEN** the provider SHALL return an error diagnostic without calling the Put index template API

#### Scenario: Feature on supported cluster version

- **GIVEN** `data_stream_options` is configured
- **AND** the connected Elasticsearch server version is `9.1.0` or above
- **WHEN** create or update runs
- **THEN** the provider SHALL include `data_stream_options` in the API request normally

---

### Requirement: Create/update — expand `data_stream_options` into API request (REQ-034)

On create and update, when `template.data_stream_options` is configured, the provider SHALL construct a `DataStreamOptions` model from the Terraform configuration and include it in the `template` field of the Put index template API request body.

#### Scenario: `failure_store.enabled` round-trip on create

- **GIVEN** `failure_store.enabled = true` configured
- **WHEN** create runs and the template is read back
- **THEN** state SHALL contain `template.data_stream_options.failure_store.enabled = true`

#### Scenario: Update changes `enabled` value

- **GIVEN** an existing template with `failure_store.enabled = true`
- **WHEN** configuration changes `failure_store.enabled` to `false` and apply runs
- **THEN** the provider SHALL send `enabled: false` in the updated API request
- **AND** state SHALL reflect `failure_store.enabled = false` after the read-back

---

### Requirement: Read — flatten `data_stream_options` from API response (REQ-035)

On read, when the API response includes `data_stream_options` inside the `template` object, the provider SHALL populate `template.data_stream_options` in Terraform state, including the `failure_store.enabled` value and `failure_store.lifecycle.data_retention` if present.

When the API response does not include `data_stream_options` (or `data_stream_options` is null), the provider SHALL leave `template.data_stream_options` unset in state.

#### Scenario: Read-back with `data_stream_options` present

- **GIVEN** the API response includes `template.data_stream_options.failure_store.enabled = true`
- **WHEN** read runs
- **THEN** state SHALL contain `template.data_stream_options.failure_store.enabled = true`

#### Scenario: Read-back with `data_stream_options` absent

- **GIVEN** the API response does not include `template.data_stream_options`
- **WHEN** read runs
- **THEN** `template.data_stream_options` SHALL be unset in state

#### Scenario: Read-back with `lifecycle.data_retention` present

- **GIVEN** the API response includes `template.data_stream_options.failure_store.lifecycle.data_retention = "10d"`
- **WHEN** read runs
- **THEN** state SHALL contain `template.data_stream_options.failure_store.lifecycle.data_retention = "10d"`

---

### Requirement: Model — `DataStreamOptions` struct in `models.Template` (REQ-036)

The internal `models.Template` struct SHALL include a `DataStreamOptions` field typed as `*DataStreamOptions` and serialized as `"data_stream_options"` in JSON. The field SHALL be `omitempty` so that templates without `data_stream_options` serialize correctly. New structs `DataStreamOptions`, `FailureStoreOptions`, and `FailureStoreLifecycle` SHALL be added to `internal/models/models.go`.

The `DataStreamOptions` struct SHALL contain:
- `FailureStore *FailureStoreOptions json:"failure_store,omitempty"`

The `FailureStoreOptions` struct SHALL contain:
- `Enabled bool json:"enabled"`
- `Lifecycle *FailureStoreLifecycle json:"lifecycle,omitempty"`

The `FailureStoreLifecycle` struct SHALL contain:
- `DataRetention string json:"data_retention,omitempty"`

Adding these fields to the shared `models.Template` struct SHALL not affect the `elasticstack_elasticsearch_component_template` resource, because the `data_stream_options` field will never be populated by that resource and `omitempty` ensures the key is absent from component template API payloads.

#### Scenario: Component template payloads are not affected

- **GIVEN** a component template create or update operation where `data_stream_options` is not configured
- **WHEN** the `models.Template` struct is serialized to JSON
- **THEN** the JSON payload SHALL NOT include the `data_stream_options` key

#### Scenario: Index template payload includes `data_stream_options`

- **GIVEN** an index template create or update operation where `data_stream_options` is configured
- **WHEN** the `models.Template` struct is serialized to JSON
- **THEN** the JSON payload SHALL include `"data_stream_options": { "failure_store": { "enabled": true } }`

---

### Requirement: Acceptance tests — `data_stream_options` coverage (REQ-037)

Acceptance tests for `elasticstack_elasticsearch_index_template` SHALL include coverage for:

- Creating a template with `data_stream_options.failure_store.enabled = true` and verifying state after create.
- Updating the template to change `failure_store.enabled` and verifying state after update.
- Creating a template with `failure_store.lifecycle.data_retention` set and verifying state after create.
- Verifying that omitting `data_stream_options` produces no drift in plan after apply.

These tests SHALL only run against Elasticsearch >= 9.1.0 and SHALL be skipped or guarded appropriately when a lower version is detected.

A unit test (not an acceptance test) SHALL verify the version-gate logic for the error path: when `data_stream_options` is configured and the detected Elasticsearch version is below `9.1.0`, the provider function under test SHALL return an error diagnostic without invoking the Put index template API.

#### Scenario: Acceptance test create with failure store enabled

- **GIVEN** an acceptance test configuration with `failure_store.enabled = true`
- **WHEN** the test creates the template and refreshes state
- **THEN** the acceptance test SHALL assert `template.0.data_stream_options.0.failure_store.0.enabled` equals `true` in state

#### Scenario: Acceptance test update failure store enabled value

- **GIVEN** an acceptance test that first creates a template with `failure_store.enabled = true`
- **WHEN** the configuration is updated to `failure_store.enabled = false` and applied
- **THEN** the acceptance test SHALL assert that the state reflects `failure_store.enabled = false` after the update

#### Scenario: Acceptance test with data_retention

- **GIVEN** an acceptance test configuration with `failure_store.lifecycle.data_retention = "14d"`
- **WHEN** the test creates the template and refreshes state
- **THEN** the acceptance test SHALL assert `template.0.data_stream_options.0.failure_store.0.lifecycle.0.data_retention` equals `"14d"` in state

#### Scenario: Unit test — version-gate error path

- **GIVEN** a unit test that simulates `data_stream_options` configured with Elasticsearch version `9.0.0`
- **WHEN** the create or update function is invoked
- **THEN** the function SHALL return an error diagnostic containing the minimum version requirement
- **AND** the Put index template API SHALL NOT be called
