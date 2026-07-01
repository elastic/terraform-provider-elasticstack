## ADDED Requirements

### Requirement: Resource identity and composite ID

The `elasticstack_fleet_agentless_policy` resource SHALL set its `id` to the composite string `"<space_id>/<policy_id>"` after every Create and Update. `policy_id` SHALL be Optional+Computed with `UseStateForUnknown`: when omitted from config, the API-assigned package policy ID SHALL be populated into state; when supplied, it SHALL be sent as the `id` field on the create request. `space_ids` SHALL be Optional+Computed defaulting to `["default"]`. Changing `policy_id`, `space_ids`, `name`, `namespace`, `package.*`, `policy_template`, or `cloud_connector.*` SHALL force resource replacement.

#### Scenario: Create with auto-assigned policy_id
- **WHEN** `policy_id` is not set in config and the resource is created
- **THEN** `policy_id` SHALL be populated from the `id` field of the `POST /api/fleet/agentless_policies` response
- **AND** `id` in state SHALL equal `"default/<policy_id>"`

#### Scenario: Create with explicit policy_id
- **WHEN** `policy_id = "my-agentless-cspm"` is set in config and the resource is created
- **THEN** the create request body SHALL include `"id": "my-agentless-cspm"`
- **AND** `id` in state SHALL equal `"default/my-agentless-cspm"`

#### Scenario: name change forces replacement
- **WHEN** `name` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: package version change forces replacement
- **WHEN** `package.version` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Schema attributes

The resource SHALL expose the following schema:

**Identity and naming (all force replacement on change unless noted):**
- `policy_id` — Optional+Computed string; server-assigned if omitted; forces replacement on change.
- `id` — Computed string; equals `"<space_id>/<policy_id>"`.
- `name` — Required string; forces replacement on change.
- `description` — Optional string; updatable in-place.
- `namespace` — Optional+Computed string; forces replacement on change.
- `space_ids` — Optional+Computed set of strings; forces replacement on change.

**Package (all force replacement on change):**
- `package` — Required object with:
  - `name` — Required string.
  - `version` — Required string.
  - `title` — Optional string.

**Policy template:**
- `policy_template` — Optional string; forces replacement on change.

**Variables and inputs:**
- `vars_json` — Optional+Computed sensitive JSON string; integration-level variables; updatable in-place.
- `var_group_selections` — Optional map(string); updatable in-place.
- `inputs` — Optional+Computed map(object) keyed by input type ID; updatable in-place. Each element:
  - `enabled` — Optional+Computed bool.
  - `condition` — Optional string.
  - `vars_json` — Optional sensitive JSON string; input-level variables.
  - `streams` — Optional map(object) keyed by stream ID. Each element:
    - `enabled` — Optional+Computed bool.
    - `vars_json` — Optional sensitive JSON string.

**Cloud connector (all force replacement on change):**
- `cloud_connector` — Optional object with:
  - `enabled` — Optional bool.
  - `cloud_connector_id` — Optional string; forces replacement on change.
  - `name` — Optional string.
  - `target_csp` — Optional string; one of `"aws"`, `"azure"`, `"gcp"`; forces replacement on change.

**Extras:**
- `global_data_tags` — Optional list of objects with `name` (string) and `value` (string); updatable in-place.
- `additional_datastreams_permissions` — Optional list of strings; updatable in-place.
- `create_dataset_templates` — Optional bool; updatable in-place.

**Operation flags:**
- `force` — Optional bool (default `false`); passed to the create request only; not persisted in read state.
- `force_delete` — Optional bool (default `false`); passed to the delete request as `?force=true` when true.

**Server-populated:**
- `created_at` — Computed string (ISO 8601 timestamp).
- `updated_at` — Computed string (ISO 8601 timestamp).

#### Scenario: vars_json stored and retrieved
- **WHEN** `vars_json = jsonencode({"cloud.account_type" = "single-account"})` is set in config
- **THEN** the create request SHALL include the parsed vars object under `vars`
- **AND** state SHALL contain the normalized JSON string in `vars_json` after Create

#### Scenario: global_data_tags passed to API
- **WHEN** `global_data_tags = [{name = "env", value = "prod"}]` is set in config
- **THEN** the create request body SHALL contain `"global_data_tags": [{"name": "env", "value": "prod"}]`

### Requirement: Create

The resource SHALL call `POST /api/fleet/agentless_policies` (space-aware) with the full create body derived from config. The response body SHALL be decoded and state SHALL be set from the response. `policy_id` and `id` SHALL be set from the response `id` field.

#### Scenario: Successful create with CSPM package
- **WHEN** a resource with `package.name = "cloud_security_posture"` and `package.version = "1.14.0"` is applied
- **THEN** `POST /api/fleet/agentless_policies` SHALL be called with a request body containing `"package": {"name": "cloud_security_posture", "version": "1.14.0"}`
- **AND** state SHALL contain `policy_id` populated from the API response
- **AND** state SHALL contain `created_at` and `updated_at` from the API response

#### Scenario: Create with cloud connector
- **WHEN** `cloud_connector = {enabled = true, cloud_connector_id = "cc-abc123", target_csp = "aws"}` is set
- **THEN** the create request body SHALL contain `"cloud_connector": {"enabled": true, "cloud_connector_id": "cc-abc123", "target_csp": "aws"}`

#### Scenario: Create with force flag
- **WHEN** `force = true` is set in config
- **THEN** the create request body SHALL include `"force": true`

#### Scenario: API error on create
- **WHEN** the API returns a non-2xx response
- **THEN** the resource SHALL surface the API error in diagnostics
- **AND** no state SHALL be saved

### Requirement: Read

The resource SHALL call `GET /api/fleet/package_policies/{policy_id}` (space-aware) to read the current state. On HTTP 404 the resource SHALL be removed from state without error. State SHALL be updated from the response on every successful Read.

#### Scenario: Successful read
- **WHEN** `GET /api/fleet/package_policies/{policy_id}` returns a 200 response
- **THEN** state SHALL be updated with the response values for all API-populated fields
- **AND** `updated_at` SHALL reflect the API's current value

#### Scenario: Resource deleted out of band
- **WHEN** the API returns HTTP 404 on Read
- **THEN** the resource SHALL be removed from state without error
- **AND** Terraform SHALL plan to recreate the resource on the next apply

#### Scenario: Read preserves force_delete
- **WHEN** Read succeeds
- **THEN** `force_delete` SHALL retain its config value (it is not returned by the API)
- **AND** `force` SHALL NOT be read from the API

### Requirement: Update

The resource SHALL call `PUT /api/fleet/package_policies/{policy_id}` (space-aware) for attributes in the in-place-updatable allowlist: `description`, `vars_json`, `var_group_selections`, `inputs`, `global_data_tags`, `additional_datastreams_permissions`, `create_dataset_templates`. After Update, state SHALL be repopulated from the PUT response.

#### Scenario: Description updated in-place
- **WHEN** `description` is changed in config
- **THEN** `PUT /api/fleet/package_policies/{policy_id}` SHALL be called with the new description
- **AND** the resource SHALL NOT be replaced

#### Scenario: inputs updated in-place
- **WHEN** an input's `vars_json` is changed in config
- **THEN** `PUT /api/fleet/package_policies/{policy_id}` SHALL be called with the updated inputs
- **AND** the resource SHALL NOT be replaced

#### Scenario: API error on update
- **WHEN** the API returns a non-2xx response during Update
- **THEN** the resource SHALL surface the API error in diagnostics
- **AND** the prior state SHALL be preserved

### Requirement: Delete

The resource SHALL call `DELETE /api/fleet/agentless_policies/{policy_id}` (space-aware). HTTP 404 SHALL be treated as success. When `force_delete = true`, the request SHALL include `?force=true`. When `force_delete = false` and the API returns a conflict error, the provider SHALL surface a helpful error message.

#### Scenario: Successful delete
- **WHEN** `force_delete = false` and the resource is destroyed
- **THEN** `DELETE /api/fleet/agentless_policies/{policy_id}` SHALL be called without the `force` query parameter
- **AND** no error SHALL be returned on 2xx

#### Scenario: Force delete
- **WHEN** `force_delete = true` and the resource is destroyed
- **THEN** `DELETE /api/fleet/agentless_policies/{policy_id}?force=true` SHALL be called

#### Scenario: Already-deleted resource
- **WHEN** the API returns HTTP 404 during Delete
- **THEN** no error SHALL be returned
- **AND** the resource SHALL be removed from state

### Requirement: Import

The resource SHALL support import with both a plain `<policy_id>` and a composite `<space_id>/<policy_id>` import ID using the existing `SpaceImporter` pattern. After import, Read SHALL populate all state attributes from the Fleet API.

#### Scenario: Import by composite ID
- **GIVEN** an agentless policy exists in Kibana space `"my-space"` with policy ID `"abc-123"`
- **WHEN** `terraform import` is run with `"my-space/abc-123"`
- **THEN** `policy_id` SHALL be `"abc-123"` and `space_ids` SHALL contain `"my-space"`
- **AND** a subsequent refresh SHALL populate all state fields from the API

#### Scenario: Import by plain policy ID
- **GIVEN** an agentless policy exists in the default Kibana space with policy ID `"abc-123"`
- **WHEN** `terraform import` is run with `"abc-123"`
- **THEN** `policy_id` SHALL be `"abc-123"`
- **AND** a subsequent refresh SHALL populate all state fields from the API

### Requirement: Version gating

The resource SHALL enforce a minimum Kibana version of 9.3.0 using the existing `EnforceMinVersion` pattern (see `internal/fleet/agentpolicy/capabilities.go`). Requests against older Kibana versions SHALL fail with a clear diagnostic naming the minimum version.

#### Scenario: Kibana version too old
- **WHEN** the connected Kibana is version 9.2.x or older
- **THEN** any Create, Update, or Delete operation SHALL return an error diagnostic stating the minimum version requirement
- **AND** no API call SHALL be made

### Requirement: Deployment topology preflight check

The resource SHALL detect self-managed (non-cloud) Kibana deployments and refuse agentless policy operations with a clear error message directing users to Elastic Cloud Hosted or Serverless. The check SHALL run at Create time.

#### Scenario: Self-managed stack rejected
- **WHEN** the resource is applied against a self-managed (on-premises) Kibana
- **THEN** Create SHALL return an error diagnostic explaining agentless policies require Elastic Cloud Hosted or Serverless
- **AND** no `POST /api/fleet/agentless_policies` call SHALL be made

#### Scenario: Cloud-hosted stack accepted
- **WHEN** the resource is applied against an Elastic Cloud Hosted or Serverless Kibana
- **THEN** the preflight check SHALL pass and the Create call SHALL proceed normally

### Requirement: Resource marked experimental

The resource description SHALL include a clear note that the Fleet agentless policy API is experimental (added in Kibana 9.3.0) and that the behavior may change in future Kibana releases. This follows the convention used by other experimental resources in the provider.

#### Scenario: Resource description contains experimental notice
- **WHEN** `terraform providers schema -json` is run
- **THEN** the `elasticstack_fleet_agentless_policy` resource description SHALL contain the word "experimental"
