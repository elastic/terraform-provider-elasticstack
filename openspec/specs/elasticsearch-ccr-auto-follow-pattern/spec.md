# elasticsearch-ccr-auto-follow-pattern Specification

## Purpose
TBD - created by archiving change elasticsearch-ccr-resources. Update Purpose after archive.
## Requirements
### Requirement: Schema — identity and pattern configuration (REQ-CCR-AFP-001)

The resource SHALL expose:

- `name` (string, required, ForceNew): the pattern name. Changing forces resource replacement.
- `remote_cluster` (string, required): the alias of the remote cluster. Updates are in-place.
- `leader_index_patterns` (list of string, required, min 1): one or more simple index patterns to
  match against indices in the remote cluster. The schema SHALL enforce at least one entry via a
  list validator (`listvalidator.SizeAtLeast(1)`).
- `leader_index_exclusion_patterns` (list of string, optional): patterns that exclude indices from
  being auto-followed even if they match `leader_index_patterns`.
- `follow_index_pattern` (string, optional, nullable): name template for follower indices; the
  placeholder `{{leader_index}}` is substituted with the matched leader index name.
- Standard Elasticsearch connection block.

#### Scenario: All required attributes accepted

- GIVEN `name = "etl-logs"`, `remote_cluster = "dc2"`, and `leader_index_patterns = ["logs-*"]`
- WHEN Terraform validates the configuration
- THEN the provider SHALL accept the configuration without error

#### Scenario: Empty leader_index_patterns rejected at plan time

- GIVEN `leader_index_patterns = []`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error indicating at least one pattern is required

#### Scenario: Exclusion patterns limit auto-following

- GIVEN `leader_index_patterns = ["logs-*"]` and `leader_index_exclusion_patterns = ["logs-debug-*"]`
- WHEN the provider creates or updates the pattern
- THEN the request body SHALL include both `leader_index_patterns` and `leader_index_exclusion_patterns` arrays

#### Scenario: follow_index_pattern uses template substitution

- GIVEN `follow_index_pattern = "{{leader_index}}-replica"`
- WHEN the provider creates the pattern
- THEN the request SHALL include `follow_index_pattern` with the configured template value

### Requirement: Schema — index settings override (REQ-CCR-AFP-002)

The resource SHALL expose `settings_raw` (string, optional, nullable): a JSON-encoded object of
Elasticsearch index settings to apply to auto-created follower indices. Changes trigger in-place
update. This attribute is write-only: it is sent to Elasticsearch on create and update but is not
returned by `GET /_ccr/auto_follow/{name}`, so it will be empty in state after `terraform import`.

#### Scenario: Valid settings_raw accepted

- GIVEN `settings_raw = jsonencode({"index.refresh_interval": "30s"})`
- WHEN Terraform validates the configuration
- THEN the provider SHALL accept it without diagnostic errors

#### Scenario: Invalid JSON in settings_raw rejected at apply

- GIVEN `settings_raw = "not-valid-json"`
- WHEN the provider attempts to apply the configuration
- THEN the provider SHALL return an error diagnostic describing the JSON parse failure

### Requirement: Schema — CCR tuning parameters (REQ-CCR-AFP-003)

The resource SHALL expose optional tuning attributes that override Elasticsearch defaults. All are
nullable. Duration attributes are strings in ES time format (e.g. `"10s"`); byte-size attributes are
strings in ES byte format (e.g. `"100mb"`); count attributes are int64. Changes trigger in-place update.

Attributes: `max_outstanding_read_requests` (int64), `max_outstanding_write_requests` (int64),
`max_read_request_operation_count` (int64), `max_read_request_size` (string, byte format),
`max_retry_delay` (string, time format), `max_write_buffer_count` (int64),
`max_write_buffer_size` (string, byte format), `max_write_request_operation_count` (int64),
`max_write_request_size` (string, byte format), `read_poll_timeout` (string, time format).

**API read limitation**: `GET /_ccr/auto_follow/{name}` returns only `max_outstanding_read_requests`
from this set. The remaining nine are accepted by the PUT API but never returned by the GET API.
All ten attributes are `Optional` only (not `Optional/Computed`). During Read the provider updates
`max_outstanding_read_requests` from the API and preserves prior-state values for the other nine
unchanged. This prevents perpetual diffs while allowing normal plan/apply management of all params.

Where the go-elasticsearch typed client uses `*int` for count fields, the provider SHALL narrow int64 schema values to int when building API requests and widen int to int64 when reading back.

#### Scenario: Tuning params included in create request

- GIVEN `max_outstanding_read_requests = 10` and `max_retry_delay = "30s"` are configured
- WHEN the resource is created
- THEN the provider SHALL include both values in the `PUT /_ccr/auto_follow/{name}` request body

#### Scenario: Tuning params updated in-place

- GIVEN the resource exists with `max_outstanding_read_requests = 10`
- WHEN `max_outstanding_read_requests` is changed to `20` in the plan
- THEN the provider SHALL update in-place via `PUT /_ccr/auto_follow/{name}` without recreating the resource

### Requirement: Schema — active attribute (REQ-CCR-AFP-004)

The resource SHALL expose `active` (bool, optional, default `true`): the desired state of the
auto-follow pattern. Practitioners MAY set `active = false` to intentionally pause the pattern.
When Elasticsearch pauses the pattern automatically, Read writes the actual state, the plan shows
the diff against the configured value, and Apply reconciles via the Update state machine.

#### Scenario: Active defaults to true

- GIVEN a configuration with no explicit `active` attribute
- WHEN the resource is created
- THEN `active` SHALL be `true` in state

#### Scenario: Active can be set to false

- GIVEN a configuration with `active = false`
- WHEN the resource is created
- THEN the provider SHALL create the pattern and immediately pause it
- AND state SHALL record `active = false`

#### Scenario: Elasticsearch-initiated pause shows as plan drift

- GIVEN the resource has `active = true` in config and Elasticsearch has automatically paused the pattern
- WHEN Read executes and a plan is generated
- THEN `terraform plan` SHALL show a change from `active = false` to `active = true`

### Requirement: Create behavior (REQ-CCR-AFP-005)

When the resource is created, the provider SHALL call `PUT /_ccr/auto_follow/{name}` with
`remote_cluster`, `leader_index_patterns`, and all other configured attributes. After a successful
create, if `active = false` is configured, the provider SHALL immediately call
`POST /_ccr/auto_follow/{name}/pause`. The provider SHALL store the configured `active` value in state.

#### Scenario: Create sends full configuration to API

- GIVEN all required and optional attributes are configured with `active = true` (default)
- WHEN the resource is created
- THEN the provider SHALL call `PUT /_ccr/auto_follow/{name}` with all configured values in the request body
- AND `active = true` SHALL be stored in state

#### Scenario: Create with active = false pauses immediately after creation

- GIVEN `active = false` is configured
- WHEN the resource is created
- THEN the provider SHALL call `PUT /_ccr/auto_follow/{name}` followed by `POST /_ccr/auto_follow/{name}/pause`
- AND `active = false` SHALL be stored in state

### Requirement: Read behavior (REQ-CCR-AFP-006)

The provider SHALL call `GET /_ccr/auto_follow/{name}` during Read. It SHALL map the following
`AutoFollowPatternSummary` fields to state: `active`, `remote_cluster`, `leader_index_patterns`,
`leader_index_exclusion_patterns`, `follow_index_pattern`, and `max_outstanding_read_requests`.
For the remaining nine tuning parameters not returned by the API, the provider SHALL preserve
their prior-state values unchanged. If the pattern is not found (404), the provider SHALL remove
it from state without error.

#### Scenario: Read maps API-returned fields to state

- GIVEN an auto-follow pattern with `active = true`, `remote_cluster = "dc2"`, `leader_index_patterns = ["logs-*"]`, and `max_outstanding_read_requests = 10` in Elasticsearch
- WHEN Read executes
- THEN all four values SHALL be reflected in Terraform state

#### Scenario: Read preserves prior-state for unreadable tuning params

- GIVEN the pattern exists in state with `max_retry_delay = "30s"` and the API does not return `max_retry_delay`
- WHEN Read executes
- THEN `max_retry_delay` SHALL remain `"30s"` in state

#### Scenario: Read removes state for out-of-band deleted pattern

- GIVEN the auto-follow pattern has been deleted outside of Terraform
- WHEN Read executes (the API returns 404)
- THEN the provider SHALL remove the resource from state and return without error

### Requirement: Update behavior (REQ-CCR-AFP-007)

The provider SHALL always call `PUT /_ccr/auto_follow/{name}` (idempotent upsert) with the full
plan configuration. After the PUT, if the `active` value changed, the provider SHALL call the
appropriate pause or resume API:

- Prior `active = true`, desired `active = false`: `POST /_ccr/auto_follow/{name}/pause`
- Prior `active = false`, desired `active = true`: `POST /_ccr/auto_follow/{name}/resume`
- No change in `active`: no additional call

The resource SHALL NOT be destroyed and recreated unless `name` changes.

#### Scenario: Update via idempotent PUT

- GIVEN the resource exists with `leader_index_patterns = ["logs-*"]`
- WHEN the configuration is changed to add `"metrics-*"` to `leader_index_patterns`
- THEN the provider SHALL call `PUT /_ccr/auto_follow/{name}` with the updated patterns array
- AND the auto-follow pattern SHALL NOT be destroyed and recreated

#### Scenario: Update active from true to false pauses the pattern

- GIVEN the resource exists with `active = true`
- WHEN `active` is changed to `false`
- THEN the provider SHALL call `PUT /_ccr/auto_follow/{name}` then `POST /_ccr/auto_follow/{name}/pause`
- AND state SHALL record `active = false`

#### Scenario: Update active from false to true resumes the pattern

- GIVEN the resource exists with `active = false`
- WHEN `active` is changed to `true`
- THEN the provider SHALL call `PUT /_ccr/auto_follow/{name}` then `POST /_ccr/auto_follow/{name}/resume`
- AND state SHALL record `active = true`

### Requirement: Destroy behavior (REQ-CCR-AFP-008)

When the resource is destroyed, the provider SHALL call `DELETE /_ccr/auto_follow/{name}`.
The provider SHALL NOT attempt to pause the pattern or delete any follower indices previously
created by the pattern.

#### Scenario: Destroy deletes the auto-follow pattern

- GIVEN the auto-follow pattern exists in Elasticsearch
- WHEN the resource is destroyed
- THEN the provider SHALL call `DELETE /_ccr/auto_follow/{name}` and remove state
- AND no follower indices previously created by this pattern SHALL be affected

### Requirement: Import support (REQ-CCR-AFP-009)

The resource SHALL support import by pattern name. After import, `active`, `remote_cluster`, `leader_index_patterns`, `leader_index_exclusion_patterns`, `follow_index_pattern`, and `max_outstanding_read_requests` SHALL be populated from the API response. The remaining nine tuning parameters are not returned by `GET /_ccr/auto_follow/{name}` and SHALL remain null in state after import unless the practitioner configures them. `settings_raw` is write-only and will be empty in state after import; practitioners MUST add it to their configuration manually if required.

#### Scenario: Import by pattern name

- GIVEN the auto-follow pattern `etl-logs` exists in Elasticsearch
- WHEN the practitioner runs `terraform import elasticstack_elasticsearch_ccr_auto_follow_pattern.example etl-logs`
- THEN Terraform state SHALL reflect the pattern's current `active`, `remote_cluster`, `leader_index_patterns`, and tuning parameters from the API
- AND `settings_raw` SHALL be empty in state

