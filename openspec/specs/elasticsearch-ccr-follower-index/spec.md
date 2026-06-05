# elasticsearch-ccr-follower-index Specification

## Purpose
TBD - created by archiving change elasticsearch-ccr-resources. Update Purpose after archive.
## Requirements
### Requirement: Schema — identity and connectivity (REQ-CCR-FI-001)

The resource SHALL expose: `name` (string, required, ForceNew — follower index name), `remote_cluster`
(string, required, ForceNew — remote cluster alias), `leader_index` (string, required, ForceNew —
leader index on the remote cluster), `data_stream_name` (string, optional, ForceNew, nullable — local
data stream name when following a data stream; write-only, not returned by `GET /{index}/_ccr/info`),
and the standard Elasticsearch connection block. All ForceNew attributes force resource replacement
when changed.

#### Scenario: All required attributes specified

- GIVEN a configuration with `name`, `remote_cluster`, and `leader_index` all set to non-empty strings
- WHEN Terraform validates the configuration
- THEN the provider SHALL accept the configuration without error

#### Scenario: Missing required attribute

- GIVEN a configuration with `leader_index` omitted
- WHEN Terraform validates the configuration
- THEN the provider SHALL return an error diagnostic for the missing required attribute

### Requirement: Schema — index settings override (REQ-CCR-FI-002)

The resource SHALL expose `settings_raw` (string, optional, nullable): a JSON-encoded object of
Elasticsearch index settings to override from the leader index. Changes to this attribute trigger
an in-place update (pause → resume). Mirrors the `settings_raw` attribute of
`elasticstack_elasticsearch_index`. This attribute is write-only: it is sent to Elasticsearch on
create and update but is not returned by `GET /{index}/_ccr/info`, so it will be empty in state
after `terraform import`.

#### Scenario: Valid settings_raw accepted

- GIVEN `settings_raw = jsonencode({"index.refresh_interval": "30s"})`
- WHEN Terraform validates the configuration
- THEN the provider SHALL accept it without diagnostic errors

#### Scenario: Invalid JSON in settings_raw rejected at apply

- GIVEN `settings_raw = "not-valid-json"`
- WHEN the provider attempts to apply the configuration
- THEN the provider SHALL return an error diagnostic describing the JSON parse failure

### Requirement: Schema — CCR tuning parameters (REQ-CCR-FI-003)

The resource SHALL expose optional tuning attributes that override Elasticsearch defaults. All are
nullable. Duration attributes are strings in ES time format (e.g. `"10s"`); byte-size attributes are
strings in ES byte format (e.g. `"100mb"`); count attributes are int64. Changes to any tuning
attribute trigger an in-place update (no ForceNew).

Attributes: `max_outstanding_read_requests` (int64), `max_outstanding_write_requests` (int64),
`max_read_request_operation_count` (int64), `max_read_request_size` (string, byte format),
`max_retry_delay` (string, time format), `max_write_buffer_count` (int64),
`max_write_buffer_size` (string, byte format), `max_write_request_operation_count` (int64),
`max_write_request_size` (string, byte format), `read_poll_timeout` (string, time format).

The go-elasticsearch typed client represents `max_outstanding_read_requests` as `*int64` and the
other four count fields as `*int`. The provider SHALL narrow int64 schema values to int when
building API requests for those four fields, and widen int to int64 when reading them back.

#### Scenario: Tuning params included in create request

- GIVEN `max_outstanding_read_requests = 12` and `read_poll_timeout = "10m"` are configured
- WHEN the resource is created
- THEN the provider SHALL include both values in the `PUT /{index}/_ccr/follow` request body

#### Scenario: Tuning params updated without ForceNew

- GIVEN the resource exists with `max_outstanding_read_requests = 12`
- WHEN `max_outstanding_read_requests` is changed to `24` in the plan
- THEN the provider SHALL update the resource in-place (pause → resume) without destroying and recreating it

### Requirement: Schema — status attribute (REQ-CCR-FI-004)

The resource SHALL expose `status` (string, optional, default `"active"`): the desired replication
status, either `"active"` or `"paused"`. Practitioners MAY set `status = "paused"` to intentionally
pause replication. When Elasticsearch automatically pauses a follower index, Read writes the actual
state, the plan shows the diff against the configured value, and Apply reconciles via the Update
state machine.

#### Scenario: Status defaults to active

- GIVEN a configuration with no explicit `status` attribute
- WHEN the resource is created
- THEN `status` SHALL be `"active"` in state

#### Scenario: Status can be set to paused

- GIVEN a configuration with `status = "paused"`
- WHEN the resource is created
- THEN the provider SHALL create the follower index and immediately pause it
- AND state SHALL record `status = "paused"`

#### Scenario: Elasticsearch-initiated pause shows as plan drift

- GIVEN the resource has `status = "active"` in config and Elasticsearch has automatically paused the follower index
- WHEN Read executes and a plan is generated
- THEN `terraform plan` SHALL show a change from `status = "paused"` to `status = "active"`

### Requirement: Schema — destroy control (REQ-CCR-FI-005)

The resource SHALL expose `delete_index_on_destroy` (bool, optional, default `false`), controlling
what happens to the underlying index when the Terraform resource is destroyed.

- When `false` (default): destroy runs pause (if active) → close → unfollow → open, leaving a usable regular (non-follower) index in Elasticsearch. The open step is required because Elasticsearch leaves the index closed after unfollow; omitting it would strand the index in a closed state.
- When `true`: destroy runs pause (if active) → close → unfollow → delete. The open step is skipped.

#### Scenario: destroy with delete_index_on_destroy = false leaves open regular index

- GIVEN `delete_index_on_destroy = false`
- WHEN the resource is destroyed
- THEN the provider SHALL pause the follower (if active), close the index, unfollow, and then open the index
- AND the underlying index SHALL exist in Elasticsearch as an open, regular (non-follower) index

#### Scenario: destroy with delete_index_on_destroy = true removes index

- GIVEN `delete_index_on_destroy = true`
- WHEN the resource is destroyed
- THEN the provider SHALL pause (if active), close, unfollow, and then delete the underlying index
- AND the index SHALL NOT exist in Elasticsearch after destroy completes

### Requirement: Create behavior (REQ-CCR-FI-006)

When the resource is created, the provider SHALL call `PUT /{name}/_ccr/follow` with `remote_cluster`,
`leader_index`, any configured tuning parameters, and `settings` when `settings_raw` is set.
After a successful create, if `status = "paused"` is configured, the provider SHALL immediately call
`POST /{name}/_ccr/pause_follow`. The provider SHALL store the configured `status` in state.

#### Scenario: Create calls follow API with all configured attributes

- GIVEN `name = "my-follower"`, `remote_cluster = "dc2"`, `leader_index = "source-index"`, and `max_outstanding_read_requests = 10`
- WHEN the resource is created
- THEN the provider SHALL send `PUT /my-follower/_ccr/follow` with a body containing `remote_cluster`, `leader_index`, and `max_outstanding_read_requests`
- AND state SHALL record `status = "active"`

#### Scenario: Create with settings_raw includes settings in request

- GIVEN `settings_raw = jsonencode({"index.refresh_interval": "30s"})`
- WHEN the resource is created
- THEN the provider SHALL deserialise `settings_raw` and include the result as the `settings` field in the create request

#### Scenario: Create with status = paused pauses immediately after creation

- GIVEN `status = "paused"` is configured
- WHEN the resource is created
- THEN the provider SHALL call `PUT /{name}/_ccr/follow` followed by `POST /{name}/_ccr/pause_follow`
- AND state SHALL record `status = "paused"`

### Requirement: Read behavior (REQ-CCR-FI-007)

The provider SHALL call `GET /{name}/_ccr/info` during Read. It SHALL map `FollowerIndex.Status`,
`FollowerIndex.RemoteCluster`, and `FollowerIndex.LeaderIndex` to state. When `Parameters != nil`,
it SHALL map all tuning fields from `Parameters` back to state attributes. When `Parameters` is nil
(follower is paused), the provider SHALL preserve the prior-state tuning parameter values unchanged
rather than zeroing them out. If the index is not found (404), the provider SHALL remove it from
state without error (out-of-band deletion).

#### Scenario: Read maps all follower index fields to state

- GIVEN a follower index with `status = "active"`, `remote_cluster = "dc2"`, and `max_outstanding_read_requests = 12` in Elasticsearch
- WHEN Read executes
- THEN all three values SHALL be reflected in Terraform state

#### Scenario: Read preserves tuning params when paused

- GIVEN the follower index is paused (Parameters omitted from API response) and prior state has `max_outstanding_read_requests = 12`
- WHEN Read executes
- THEN `max_outstanding_read_requests` SHALL remain `12` in state
- AND `status` SHALL be `"paused"` in state

#### Scenario: Read removes state for out-of-band deleted index

- GIVEN the follower index has been deleted outside of Terraform
- WHEN Read executes (the API returns 404)
- THEN the provider SHALL remove the resource from state and return without error

### Requirement: Update behavior (REQ-CCR-FI-008)

The provider SHALL implement the following state machine in the Update callback, driven by the prior
state `status` and the desired plan `status`. Note that `resumefollow.Request` accepts only tuning
parameters — it has no `Settings` field. When `settings_raw` changes, the provider SHALL apply the
new settings via `PUT /{name}/_settings` while the follower is paused, before resuming.

- Prior `active`, desired `active`, tuning params and/or `settings_raw` changed: `POST /{name}/_ccr/pause_follow`, then `PUT /{name}/_settings` if `settings_raw` changed, then `POST /{name}/_ccr/resume_follow` with updated tuning params
- Prior `active`, desired `paused`: `POST /{name}/_ccr/pause_follow` only; tuning and `settings_raw` changes are stored in state and applied on the next active transition
- Prior `paused`, desired `active`: `PUT /{name}/_settings` if `settings_raw` changed, then `POST /{name}/_ccr/resume_follow` with plan tuning params
- Prior `paused`, desired `paused`: no API call; changes stored in state

#### Scenario: Update tuning params via pause/resume when active

- GIVEN the resource exists with `status = "active"` and `max_outstanding_read_requests = 12`
- WHEN the configuration is changed to `max_outstanding_read_requests = 24`
- THEN the provider SHALL call `POST /{name}/_ccr/pause_follow` followed by `POST /{name}/_ccr/resume_follow` with `max_outstanding_read_requests = 24`
- AND data in the follower index SHALL be preserved

#### Scenario: Update status from active to paused

- GIVEN the resource exists with `status = "active"`
- WHEN the configuration is changed to `status = "paused"`
- THEN the provider SHALL call `POST /{name}/_ccr/pause_follow`
- AND state SHALL record `status = "paused"`

#### Scenario: Update status from paused to active

- GIVEN the resource exists with `status = "paused"` and `max_outstanding_read_requests = 12`
- WHEN the configuration is changed to `status = "active"`
- THEN the provider SHALL call `POST /{name}/_ccr/resume_follow` with `max_outstanding_read_requests = 12`
- AND state SHALL record `status = "active"`

#### Scenario: Tuning param update while paused defers to next resume

- GIVEN the resource exists with `status = "paused"` and `max_outstanding_read_requests = 12`
- WHEN `max_outstanding_read_requests` is changed to `24` with `status` remaining `"paused"`
- THEN the provider SHALL make no API calls
- AND state SHALL record `max_outstanding_read_requests = 24` for application on the next resume

### Requirement: Import support (REQ-CCR-FI-009)

The resource SHALL support import by follower index name. After import, `status`, `remote_cluster`,
`leader_index`, and all readable tuning parameters SHALL be populated from the API response.
`delete_index_on_destroy` SHALL default to `false`. `settings_raw` and `data_stream_name` are
write-only and will be empty in state after import; practitioners MUST add them to their
configuration manually. In particular, omitting `data_stream_name` after importing a data stream
follower will produce a plan diff that forces resource replacement.

#### Scenario: Import by follower index name

- GIVEN the follower index `my-follower-index` exists in Elasticsearch
- WHEN the practitioner runs `terraform import elasticstack_elasticsearch_ccr_follower_index.example my-follower-index`
- THEN Terraform state SHALL reflect the follower index's current `status`, `remote_cluster`, `leader_index`, and tuning parameters from the API
- AND `delete_index_on_destroy` SHALL be `false`
- AND `settings_raw` and `data_stream_name` SHALL be empty in state

#### Scenario: Setting data_stream_name after import triggers replacement

- GIVEN a data stream follower index has been imported and `data_stream_name` is null in state
- WHEN the practitioner adds `data_stream_name = "my-data-stream"` to their configuration
- THEN `terraform plan` SHALL show a ForceNew replacement plan for the resource
- AND no in-place update path SHALL exist (the attribute is ForceNew and write-only)

