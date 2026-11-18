## MODIFIED Requirements

### Requirement: State mapping from API response (REQ-022)

After any create, update, or read operation, the resource SHALL populate the following fields from the API response: `id`, `policy_id`, `name`, `namespace`, `description`, `integration_name`, `integration_version`, `output_id`, `additional_datastreams_permissions`. The resource SHALL populate `agent_policy_id` from the API response when `agent_policy_id` was the originally configured attribute, and `agent_policy_ids` from the API response when `agent_policy_ids` was the originally configured attribute. When `output_id` is not configured (null), the resource SHALL keep `output_id` null in state after create, update and read, whatever the API response contains (an output ID or an empty string), because the value may have been assigned outside Terraform and, from Kibana 9.0.0, the Fleet update API retains an omitted `output_id`. On import, the resource SHALL populate `output_id` from the API response. When `output_id` is not configured, the resource SHALL omit it from the update request; on Kibana 9.0.0 and later the server retains the previously assigned output, while earlier versions clear it. In both cases state remains null; explicitly clearing the server-side value is out of scope. When `space_ids` is returned by the API, the resource SHALL set it from the response; when not returned and `space_ids` was not originally set, the resource SHALL set it to null. When the API response omits `additional_datastreams_permissions` or returns an empty list, the resource SHALL set the attribute to null in state so that an unconfigured attribute does not produce a plan diff. The resource SHALL NOT map the API response's `enabled` field into Terraform state — the Kibana Fleet package-policy create/update API does not accept a top-level `enabled` value, the response field is always `true`, and the attribute is no longer part of the schema.

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

#### Scenario: output_id set outside Terraform

- **GIVEN** `output_id` is not configured and an output is assigned to the policy outside Terraform
- **WHEN** the unchanged configuration is applied or the resource is refreshed
- **THEN** apply SHALL succeed, `output_id` in state SHALL remain null, and on Kibana 9.0.0 and later the output SHALL remain assigned on the server

#### Scenario: Empty output_id on the server

- **GIVEN** `output_id` is not configured and the policy's `output_id` is `""` on the server
- **WHEN** the unchanged configuration is applied or the resource is refreshed
- **THEN** apply SHALL succeed and `output_id` in state SHALL remain null

#### Scenario: output_id removed from configuration

- **GIVEN** `output_id` was configured and applied
- **WHEN** `output_id` is removed from the configuration and applied
- **THEN** apply SHALL succeed, `output_id` in state SHALL be null, and on Kibana 9.0.0 and later the previously assigned output SHALL be retained on the server (earlier versions clear it)

#### Scenario: Import populates output_id

- **GIVEN** an existing policy with an output assigned
- **WHEN** the policy is imported and read runs
- **THEN** `output_id` in state SHALL be set from the API response
