## ADDED Requirements

### Requirement: additional_datastreams_permissions attribute (REQ-027)

The resource SHALL expose an optional `additional_datastreams_permissions` attribute of type `list(string)`. Each element SHALL be a data stream name or pattern (for example `logs-custom-*` or `metrics-elastic_agent.my_dataset-default`) for which the integration's Elasticsearch API key is granted write access in addition to the data streams the integration package declares. The attribute grants the privileges required by an ingest pipeline `reroute` processor whose destination lies outside the package's own data streams; in the Kibana UI this control is labelled "Add a reroute processor permission" under a policy's advanced options.

The attribute SHALL NOT be `Computed`: when it is absent from configuration it SHALL be null in state rather than populated with a server-side default.

The attribute SHALL be rejected at plan time when configured as an empty list, via a schema-level minimum-size validator, in the same way `agent_policy_ids` is validated. Removing the attribute from configuration — not setting it to `[]` — is the supported way to revoke previously granted permissions.

#### Scenario: Permissions configured for a reroute destination

- **GIVEN** `additional_datastreams_permissions = ["metrics-elastic_agent.my_dataset-default"]`
- **WHEN** Terraform applies the configuration
- **THEN** the integration policy SHALL be created with that data stream granted in addition to the package's own data streams

#### Scenario: Attribute omitted

- **GIVEN** a configuration that does not set `additional_datastreams_permissions`
- **WHEN** Terraform applies the configuration and then refreshes state
- **THEN** `additional_datastreams_permissions` SHALL be null in state and no plan diff SHALL appear on a subsequent plan

### Requirement: additional_datastreams_permissions — clearing removes server-side permissions (REQ-028)

When `additional_datastreams_permissions` was previously set in state and is removed from configuration, the resource SHALL send an explicit empty array on update so that Kibana revokes the previously granted permissions. The resource SHALL NOT omit the field in this case, because omission would leave the existing permissions in place and state would no longer describe the policy.

#### Scenario: Removing the attribute revokes permissions

- **GIVEN** an applied integration policy with `additional_datastreams_permissions = ["logs-custom-*"]`
- **WHEN** the attribute is removed from configuration and Terraform applies
- **THEN** the update request SHALL send an empty array for `additional_datastreams_permissions`
- **AND** a subsequent read SHALL show no additional data stream permissions on the policy

#### Scenario: Empty list rejected at plan time

- **GIVEN** `additional_datastreams_permissions = []`
- **WHEN** Terraform validates the configuration
- **THEN** a validation error SHALL be returned directing the user to remove the attribute instead

#### Scenario: Clearing is not attempted against an unsupported server

- **GIVEN** the attribute is not configured and the server version is below 9.1.0
- **WHEN** create or update runs
- **THEN** the request body SHALL NOT include an `additional_datastreams_permissions` field

### Requirement: Compatibility — additional_datastreams_permissions (REQ-029)

When `additional_datastreams_permissions` is configured with a known value, the resource SHALL verify the server version is at least 9.1.0. If the server version is lower, the resource SHALL return an attribute-level error diagnostic with "Unsupported Elasticsearch version" and SHALL not call the Fleet API. The Fleet package policy API accepted this field starting in Kibana 9.1.0 and it was not backported to 8.x, so no 8.x server supports it.

#### Scenario: additional_datastreams_permissions on old server

- **GIVEN** `additional_datastreams_permissions` is set and the server version is below 9.1.0
- **WHEN** create or update runs
- **THEN** the provider SHALL return an attribute-level error diagnostic and SHALL not call the create/update API

#### Scenario: Attribute unset on old server

- **GIVEN** `additional_datastreams_permissions` is not configured and the server version is below 9.1.0
- **WHEN** create or update runs
- **THEN** the provider SHALL not return a version diagnostic and the operation SHALL proceed

### Requirement: additional_datastreams_permissions — server-side validation errors surfaced (REQ-030)

Kibana validates `additional_datastreams_permissions` entries against the space's allowed namespace prefixes and rejects values it does not accept. The resource SHALL NOT re-implement that validation locally. When the Fleet API rejects a create or update because of the configured permissions, the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: Kibana rejects a permission entry

- **GIVEN** `additional_datastreams_permissions` contains a value that Kibana rejects for the target space
- **WHEN** create or update runs
- **THEN** diagnostics SHALL include the Fleet API error and the operation SHALL be aborted

## MODIFIED Requirements

### Requirement: Create — API request body (REQ-011)

On create, the resource SHALL construct a `PackagePolicyRequest` from the plan model and submit it to the Fleet create package policy API. The request body SHALL include `name`, `namespace`, `description` (if set), `force` (if set), `integration_name` and `integration_version` as the package reference, `agent_policy_id` or `policy_ids` based on which attribute is configured, `output_id` if set, `additional_datastreams_permissions` if set, `vars` from `vars_json` (with provider-internal context keys stripped before sending), and `inputs` derived from the `inputs` attribute. When `space_ids` is configured with a known value, the first element SHALL be used as the space context for the create API call.

#### Scenario: space context from space_ids

- GIVEN `space_ids` is set to `["my-space"]`
- WHEN create runs
- THEN the package policy SHALL be created in the "my-space" Kibana space

#### Scenario: additional_datastreams_permissions included in create body

- **GIVEN** `additional_datastreams_permissions = ["logs-custom-*"]` in the plan
- **WHEN** create runs
- **THEN** the create request body SHALL include `additional_datastreams_permissions` with that single element

#### Scenario: Empty array sent when unset against a supporting server

- **GIVEN** `additional_datastreams_permissions` is null in the plan and the server version is at least 9.1.0
- **WHEN** create or update runs
- **THEN** the request body SHALL include `additional_datastreams_permissions` as an empty array, so that any previously granted permissions are revoked

#### Scenario: Field omitted when unset against an older server

- **GIVEN** `additional_datastreams_permissions` is null in the plan and the server version is below 9.1.0
- **WHEN** create or update runs
- **THEN** the request body SHALL NOT include an `additional_datastreams_permissions` field

### Requirement: State mapping from API response (REQ-022)

After any create, update, or read operation, the resource SHALL populate the following fields from the API response: `id`, `policy_id`, `name`, `namespace`, `description`, `integration_name`, `integration_version`, `output_id`, `additional_datastreams_permissions`. The resource SHALL populate `agent_policy_id` from the API response when `agent_policy_id` was the originally configured attribute, and `agent_policy_ids` from the API response when `agent_policy_ids` was the originally configured attribute. When `space_ids` is returned by the API, the resource SHALL set it from the response; when not returned and `space_ids` was not originally set, the resource SHALL set it to null. When the API response omits `additional_datastreams_permissions` or returns an empty list, the resource SHALL set the attribute to null in state so that an unconfigured attribute does not produce a plan diff. The resource SHALL NOT map the API response's `enabled` field into Terraform state — the Kibana Fleet package-policy create/update API does not accept a top-level `enabled` value, the response field is always `true`, and the attribute is no longer part of the schema.

#### Scenario: agent_policy_id preserved when originally configured

- GIVEN a resource created with `agent_policy_id = "policy-abc"`
- WHEN read refreshes state
- THEN `agent_policy_id` in state SHALL be set from the API response and `agent_policy_ids` SHALL remain unconfigured

#### Scenario: additional_datastreams_permissions read back from API

- **GIVEN** a policy whose API response contains `additional_datastreams_permissions: ["logs-custom-*", "metrics-acc-*"]`
- **WHEN** read refreshes state
- **THEN** `additional_datastreams_permissions` in state SHALL be a two-element list preserving the API order

#### Scenario: Empty API list becomes null in state

- **GIVEN** a policy whose API response omits `additional_datastreams_permissions` or returns an empty array
- **WHEN** read refreshes state
- **THEN** `additional_datastreams_permissions` SHALL be null in state

#### Scenario: Import populates permissions

- **GIVEN** an existing policy with additional data stream permissions granted outside Terraform
- **WHEN** the policy is imported and read runs
- **THEN** `additional_datastreams_permissions` in state SHALL reflect the API response
