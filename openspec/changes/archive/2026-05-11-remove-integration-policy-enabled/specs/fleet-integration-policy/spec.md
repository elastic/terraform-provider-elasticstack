## REMOVED Schema Attribute

The top-level `enabled` attribute SHALL be removed from the
`elasticstack_fleet_integration_policy` resource schema. The Kibana Fleet
package-policy create/update API does not accept a top-level enabled flag, so
the attribute was a no-op. Per-input and per-stream `enabled` attributes
(`inputs.<id>.enabled`, `inputs.<id>.streams.<id>.enabled`) are unaffected and
remain the supported way to disable individual telemetry pipelines.

## MODIFIED Requirements

### Requirement: State mapping from API response (REQ-022)

After any create, update, or read operation, the resource SHALL populate the
following fields from the API response: `id`, `policy_id`, `name`, `namespace`,
`description`, `integration_name`, `integration_version`, `output_id`. The
resource SHALL populate `agent_policy_id` from the API response when
`agent_policy_id` was the originally configured attribute, and
`agent_policy_ids` from the API response when `agent_policy_ids` was the
originally configured attribute. When `space_ids` is returned by the API, the
resource SHALL set it from the response; when not returned and `space_ids` was
not originally set, the resource SHALL set it to null. The resource SHALL NOT
map the API response's `enabled` field into Terraform state (the attribute is
no longer part of the schema).

#### Scenario: agent_policy_id preserved when originally configured

- GIVEN a resource created with `agent_policy_id = "policy-abc"`
- WHEN read refreshes state
- THEN `agent_policy_id` in state SHALL be set from the API response and
  `agent_policy_ids` SHALL remain unconfigured

### Requirement: State upgrade — v0 to v3 (REQ-024)

The resource SHALL support state upgrade from schema version 0 to version 3 via
intermediate v1 and v3 conversions. During v0→v1: `vars_json` and all input
`vars_json`/`streams_json` string fields with empty string values SHALL be
converted to normalized JSON null; non-empty values SHALL be wrapped in
`jsontypes.Normalized`. The `agent_policy_ids` and `space_ids` fields absent in
v0 SHALL be initialized to null. During v1→v3: the `input` list block SHALL be
converted to an `inputs` map attribute keyed by `input_id`; each input's
`streams_json` normalized JSON string SHALL be parsed and converted to the
`streams` map structure; `vars_json` SHALL be migrated to the `VarsJSON` custom
type with integration context attached. The legacy `enabled` field present in
v0/v1/v2 state SHALL be dropped (the attribute no longer exists in v3).

#### Scenario: v0 empty vars_json becomes null

- GIVEN v0 state with `vars_json = ""`
- WHEN state upgrade to v3 runs
- THEN `vars_json` in v3 state SHALL be null

#### Scenario: v0 enabled field dropped

- GIVEN v0 state with `enabled = false`
- WHEN state upgrade to v3 runs
- THEN the resulting v3 state SHALL NOT contain an `enabled` attribute

### Requirement: State upgrade — v1 to v3 (REQ-025)

The resource SHALL support state upgrade from schema version 1 to version 3
directly. The v1→v3 upgrade SHALL apply the same `input` list to `inputs` map
conversion and `streams_json` expansion described in REQ-024. All other fields
(id, policy_id, name, namespace, agent_policy_id, agent_policy_ids,
description, force, integration_name, integration_version, output_id,
space_ids) SHALL be carried over unchanged. The legacy `enabled` field SHALL be
dropped.

#### Scenario: v1 to v3 direct upgrade

- GIVEN v1 state with an `input` list block containing one entry and `enabled = true`
- WHEN state upgrade to v3 runs directly (v1→v3 path)
- THEN `inputs` in v3 state SHALL be a map keyed by the entry's `input_id`, all
  other scalar fields SHALL be unchanged, and the resulting state SHALL NOT
  contain a top-level `enabled` attribute

## ADDED Requirements

### Requirement: State upgrade — v2 to v3 (REQ-026)

The resource SHALL support state upgrade from schema version 2 to version 3.
The upgrade SHALL decode prior state using the prior v2 schema (which contained
the now-removed top-level `enabled` attribute), drop the `enabled` value, and
write all other v2 fields (id, policy_id, name, namespace, agent_policy_id,
agent_policy_ids, description, force, integration_name, integration_version,
output_id, space_ids, vars_json, inputs) into the v3 model unchanged. The
upgrade SHALL NOT call the Fleet API.

#### Scenario: v2 to v3 drops enabled

- GIVEN v2 state with `enabled = true` and a populated `inputs` map
- WHEN state upgrade to v3 runs
- THEN the resulting v3 state SHALL preserve `inputs` and all other scalar
  fields and SHALL NOT contain an `enabled` attribute

#### Scenario: v2 to v3 with disabled flag

- GIVEN v2 state with `enabled = false` (a value that the provider previously
  ignored on writes)
- WHEN state upgrade to v3 runs
- THEN the resulting v3 state SHALL NOT contain an `enabled` attribute and the
  upgrade SHALL succeed without error
