# `elasticstack_fleet_integration_policy` — Schema and Functional Requirements

Resource implementation: `internal/fleet/integration_policy`

## Purpose

Manage Fleet integration policies (package policies), which configure a specific integration package for use within one or more Fleet agent policies. The resource uses the Kibana Fleet package policy API to create, read, update, and delete integration policies. It supports configuring integration-level variables, per-input variables and streams, space assignment, and agent policy association, including multi-policy assignment via `agent_policy_ids` (available from Elastic Stack 8.15.0) and output routing via `output_id` (available from 8.16.0).

## Schema

```hcl
resource "elasticstack_fleet_integration_policy" "example" {
  id                  = <computed, string>             # same as policy_id; UseStateForUnknown
  policy_id           = <optional+computed, string>    # force new; UseStateForUnknown; import key
  name                = <required, string>
  namespace           = <required, string>

  agent_policy_id     = <optional, string>             # conflicts with agent_policy_ids
  agent_policy_ids    = <optional, list(string)>       # conflicts with agent_policy_id; size >= 1; requires server >= 8.15.0

  description         = <optional, string>
  enabled             = <optional+computed, bool>      # default true
  force               = <optional, bool>
  integration_name    = <required, string>
  integration_version = <required, string>
  output_id           = <optional, string>             # requires server >= 8.16.0

  vars_json           = <optional+computed, json string>   # integration-level variables; sensitive; UseStateForUnknown
  space_ids           = <optional+computed, set(string)>   # UseStateForUnknown

  inputs = <optional+computed, map(object)> {          # keyed by input ID; UseStateForUnknown
    enabled  = <optional+computed, bool>               # default true
    vars     = <optional, json string (normalized)>    # input-level variables; sensitive
    defaults = <computed, object> {                    # populated from package info
      vars    = <computed, json string (normalized)>
      streams = <computed, map(object)> {
        enabled = <computed, bool>
        vars    = <computed, json string (normalized)>
      }
    }
    streams = <optional, map(object)> {                # keyed by stream ID
      enabled = <optional+computed, bool>              # default true
      vars    = <optional, json string (normalized)>   # stream-level variables; sensitive
    }
  }
}
```
## Requirements
### Requirement: Fleet package policy CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Kibana Fleet create package policy API to create integration policies. The resource SHALL use the Kibana Fleet update package policy API to update integration policies. The resource SHALL use the Kibana Fleet get package policy API to read integration policies. The resource SHALL use the Kibana Fleet delete package policy API to delete integration policies. When the Fleet API returns a non-success response for any of these operations (other than not found on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN a failing Fleet API response during package policy creation
- WHEN create runs
- THEN diagnostics SHALL include the API error and the operation SHALL be aborted

### Requirement: Identity (REQ-005)

The resource SHALL expose a computed `id` attribute equal to the `policy_id` returned by the Fleet API. The resource SHALL also expose a computed `policy_id` attribute, set to the same value as `id`. Both `id` and `policy_id` SHALL be persisted with `UseStateForUnknown` so they are stable across plan/apply cycles.

#### Scenario: id equals policy_id

- GIVEN a successful create
- WHEN the response is stored in state
- THEN `id` SHALL equal `policy_id`

### Requirement: Import (REQ-006)

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<policy_id>` (as
parsed by `clients.CompositeIDFromStrFw`), the resource SHALL set `policy_id` to the
parsed resource-ID segment and SHALL set `space_ids` to a single-element set containing the
space-ID segment. The subsequent read SHALL query the package-policy API in the named space,
so that policies created in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` separator
that `clients.CompositeIDFromStrFw` recognises as a composite ID — the resource SHALL treat
the entire string as `policy_id` and SHALL NOT set `space_ids` from the import ID. This
preserves existing behaviour for default-space imports.

When the import ID contains a `/` separator but either the space-ID segment or the
policy-ID segment is empty (e.g. `"/policy-id"` or `"space-id/"`), the resource SHALL
return an error diagnostic describing the expected format and SHALL NOT partially populate
`policy_id` or `space_ids`.

On the subsequent read after import (regardless of ID form), the resource SHALL populate all
attributes from the Fleet API response, including inputs.

#### Scenario: Import by composite space/policy ID

- GIVEN a package policy that exists in the Kibana space `"my-space"` with policy ID
  `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `policy_id` SHALL be `"abc-123"`, `space_ids` SHALL contain `"my-space"`, and a
  subsequent refresh SHALL populate all state fields from the API

#### Scenario: Import by plain policy ID (default space)

- GIVEN a package policy that exists in the default Kibana space with policy ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `policy_id` SHALL be `"abc-123"`, `space_ids` SHALL NOT be set from the import ID,
  and a subsequent refresh SHALL populate all state fields from the API

### Requirement: Lifecycle — policy_id requires replacement (REQ-007)

When `policy_id` changes (e.g. a user provides an explicit value that differs from the computed one), the resource SHALL require replacement.

#### Scenario: Explicit policy_id change

- GIVEN `policy_id` changes in configuration
- WHEN Terraform plans
- THEN resource replacement SHALL be required

### Requirement: Connection (REQ-008)

The resource SHALL use the provider-level Fleet client obtained from provider configuration by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the Fleet client derived from the scoped connection for all CRUD operations.

#### Scenario: Provider Fleet client used by default

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** the resource SHALL obtain its Fleet client from the provider configuration

#### Scenario: Scoped Fleet client used when overridden

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** the resource SHALL obtain its effective Fleet client from the scoped connection for create, read, update, and delete

### Requirement: Compatibility — agent_policy_ids (REQ-009)

When `agent_policy_ids` is configured with a known value, the resource SHALL verify the server version is at least 8.15.0. If the server version is lower, the resource SHALL return an attribute-level error diagnostic with "Unsupported Elasticsearch version" and SHALL not call the Fleet API.

#### Scenario: agent_policy_ids on old server

- GIVEN `agent_policy_ids` is set and the server version is below 8.15.0
- WHEN create or update runs
- THEN the provider SHALL return an attribute-level error diagnostic and SHALL not call the create/update API

### Requirement: Compatibility — output_id (REQ-010)

When `output_id` is configured with a known value, the resource SHALL verify the server version is at least 8.16.0. If the server version is lower, the resource SHALL return an attribute-level error diagnostic with "Unsupported Elasticsearch version" and SHALL not call the Fleet API.

#### Scenario: output_id on old server

- GIVEN `output_id` is set and the server version is below 8.16.0
- WHEN create or update runs
- THEN the provider SHALL return an attribute-level error diagnostic and SHALL not call the create/update API

### Requirement: Create — API request body (REQ-011)

On create, the resource SHALL construct a `PackagePolicyRequest` from the plan model and submit it to the Fleet create package policy API. The request body SHALL include `name`, `namespace`, `description` (if set), `force` (if set), `integration_name` and `integration_version` as the package reference, `agent_policy_id` or `policy_ids` based on which attribute is configured, `output_id` if set, `vars` from `vars_json` (with provider-internal context keys stripped before sending), and `inputs` derived from the `inputs` attribute. When `space_ids` is configured with a known value, the first element SHALL be used as the space context for the create API call.

#### Scenario: space context from space_ids

- GIVEN `space_ids` is set to `["my-space"]`
- WHEN create runs
- THEN the package policy SHALL be created in the "my-space" Kibana space

### Requirement: Create — policy_id in request (REQ-012)

When `policy_id` is configured with a known value at create time, the resource SHALL include it as the `id` field in the create request body to create the policy with that specific ID.

#### Scenario: Explicit policy_id propagated to create body

- GIVEN `policy_id` is set to a known value in the plan
- WHEN create runs
- THEN the create request body SHALL include that value as the `id` field

### Requirement: Create — read-back after create (REQ-013)

After a successful create, the resource SHALL retrieve package info for the created policy's package (name and version) from the Fleet registry cache. The resource SHALL call `populateFromAPI` to set all state fields from the API response. When `inputs` was null or empty in the plan, the resource SHALL not populate `inputs` from the API response (to avoid provider-produced inconsistent result errors).

#### Scenario: Inputs omitted in plan

- GIVEN `inputs` is null or empty in the plan
- WHEN create completes
- THEN `inputs` in state SHALL be null (not populated from API)

### Requirement: Update — space-aware operation (REQ-014)

On update, the resource SHALL derive the operational space from the prior state using `GetOperationalSpaceFromState`, and SHALL submit the update request in that space context. The API will handle adding or removing the policy from spaces based on the `space_ids` field in the request body.

#### Scenario: Update uses operational space from state

- GIVEN `space_ids` was set to `["my-space"]` in prior state
- WHEN update runs
- THEN the update API call SHALL use "my-space" as the Kibana space context

### Requirement: Update — inputs state preservation (REQ-015)

On update, when neither the prior state nor the plan had inputs configured (both null/empty), the resource SHALL not populate `inputs` from the API response.

#### Scenario: Inputs not added during update when unconfigured

- GIVEN `inputs` is null in both prior state and plan
- WHEN update completes
- THEN `inputs` in the updated state SHALL remain null

### Requirement: Read — not found removes resource (REQ-016)

On read, the resource SHALL derive the `policy_id` from state and use `GetOperationalSpaceFromState` to determine the Kibana space. If the Fleet get package policy API returns nil (not found), the resource SHALL remove itself from Terraform state.

#### Scenario: Policy not found on refresh

- GIVEN the integration policy was deleted outside Terraform
- WHEN read runs
- THEN the resource SHALL be removed from state

### Requirement: Read — import detection and inputs population (REQ-017)

During read, when `policy_id` has a value but `name` is null or unknown (indicating an import operation), the resource SHALL always populate `inputs` from the API response. When `inputs` was previously known and non-empty in state, the resource SHALL populate `inputs` from the API. When `inputs` was previously known and null/empty (user did not configure inputs), the resource SHALL not populate `inputs` from the API, leaving it null.

#### Scenario: Import populates all inputs

- GIVEN an import where only `policy_id` is set in state
- WHEN read runs after import
- THEN `inputs` SHALL be fully populated from the API response

### Requirement: Secrets handling (REQ-018)

The resource SHALL use provider private state to store a secret store mapping secret reference IDs to their original plaintext values. On create and update, `HandleReqRespSecrets` SHALL be called to map secret references in the API response back to the original plaintext values from the request, and store the mapping in private state. On read, `HandleRespSecrets` SHALL be called to replace secret references in the API response with the stored original values. The secret store SHALL be filtered on each read to remove entries whose reference IDs are no longer present in the API response.

#### Scenario: Secret value preserved across read

- GIVEN a variable whose value was stored as a secret reference by Fleet
- WHEN read refreshes state
- THEN the variable value in state SHALL be the original plaintext value, not the secret reference

### Requirement: vars_json — sanitization before API call (REQ-019)

Before submitting `vars_json` to the Fleet API, the resource SHALL strip any provider-internal context keys (such as `__tf_provider_context`) from the vars map using `SanitizedValue`. This prevents Fleet API 400 errors caused by unrecognized internal variables.

#### Scenario: Internal context keys stripped before API call

- GIVEN `vars_json` contains a `__tf_provider_context` key added by the provider
- WHEN create or update runs
- THEN the request body sent to Fleet SHALL not include the `__tf_provider_context` key

### Requirement: inputs — map-keyed structure (REQ-020)

The `inputs` attribute SHALL be a map keyed by input ID (e.g. `"logfile-1"`). Each entry SHALL contain `enabled`, `vars` (JSON), `defaults` (computed from package info), and `streams` (a map keyed by stream ID containing `enabled` and `vars`). When the API returns streams, the resource SHALL populate them in state. When there are no streams, the resource SHALL store `streams` as null in state.

#### Scenario: Streams null when API returns no streams

- GIVEN an integration input that has no streams in the API response
- WHEN read populates the inputs map
- THEN `streams` for that input SHALL be null in state

### Requirement: agent_policy_id and agent_policy_ids mutual exclusion (REQ-021)

The `agent_policy_id` and `agent_policy_ids` attributes SHALL be mutually exclusive: configuring both simultaneously SHALL be rejected at plan time via schema-level `ConflictsWith` validators.

#### Scenario: Both agent policy fields set

- GIVEN both `agent_policy_id` and `agent_policy_ids` are configured
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: State mapping from API response (REQ-022)

After any create, update, or read operation, the resource SHALL populate the following fields from the API response: `id`, `policy_id`, `name`, `namespace`, `description`, `enabled`, `integration_name`, `integration_version`, `output_id`. The resource SHALL populate `agent_policy_id` from the API response when `agent_policy_id` was the originally configured attribute, and `agent_policy_ids` from the API response when `agent_policy_ids` was the originally configured attribute. When `space_ids` is returned by the API, the resource SHALL set it from the response; when not returned and `space_ids` was not originally set, the resource SHALL set it to null.

#### Scenario: agent_policy_id preserved when originally configured

- GIVEN a resource created with `agent_policy_id = "policy-abc"`
- WHEN read refreshes state
- THEN `agent_policy_id` in state SHALL be set from the API response and `agent_policy_ids` SHALL remain unconfigured

### Requirement: Package info caching (REQ-023)

The resource SHALL cache package info retrieved from the Fleet registry using a process-level in-memory cache keyed by `<name>-<version>`. Package info requests SHALL be scoped to the operational Kibana space (using the same space context as the package policy create/read/update operation) so that space-restricted API keys can retrieve package details. When the exact version is not found in the registry, the resource SHALL fall back to querying without a version (returning the installed package), and SHALL emit a warning diagnostic. When neither lookup finds the package, the resource SHALL emit a warning and proceed without package info defaults.

#### Scenario: Version not found falls back to installed package

- GIVEN an integration policy whose `integration_version` is no longer available in the registry
- WHEN the resource fetches package info
- THEN the resource SHALL fall back to querying the installed package version and SHALL emit a warning diagnostic

### Requirement: State upgrade — v0 to v2 (REQ-024)

The resource SHALL support state upgrade from schema version 0 to version 2 via an intermediate v1 conversion. During v0→v1: `vars_json` and all input `vars_json`/`streams_json` string fields with empty string values SHALL be converted to normalized JSON null; non-empty values SHALL be wrapped in `jsontypes.Normalized`. The `agent_policy_ids` and `space_ids` fields absent in v0 SHALL be initialized to null. During v1→v2: the `input` list block SHALL be converted to an `inputs` map attribute keyed by `input_id`; each input's `streams_json` normalized JSON string SHALL be parsed and converted to the `streams` map structure; `vars_json` SHALL be migrated to the `VarsJSON` custom type with integration context attached.

#### Scenario: v0 empty vars_json becomes null

- GIVEN v0 state with `vars_json = ""`
- WHEN state upgrade to v2 runs
- THEN `vars_json` in v2 state SHALL be null

#### Scenario: v1 input list becomes v2 inputs map

- GIVEN v1 state with `input = [{ input_id = "logfile-1", enabled = true, vars_json = "{}", streams_json = null }]`
- WHEN state upgrade to v2 runs
- THEN `inputs` in v2 state SHALL be a map with key `"logfile-1"` containing `enabled = true` and `vars` populated from the v1 `vars_json`

### Requirement: State upgrade — v1 to v2 (REQ-025)

The resource SHALL support state upgrade from schema version 1 to version 2 directly. The v1→v2 upgrade SHALL apply the same `input` list to `inputs` map conversion and `streams_json` expansion described in REQ-024. All other fields (id, policy_id, name, namespace, agent_policy_id, agent_policy_ids, description, enabled, force, integration_name, integration_version, output_id, space_ids) SHALL be carried over unchanged.

#### Scenario: v1 to v2 direct upgrade

- GIVEN v1 state with an `input` list block containing one entry
- WHEN state upgrade to v2 runs directly (v1→v2 path)
- THEN `inputs` in v2 state SHALL be a map keyed by the entry's `input_id` and all scalar fields SHALL be unchanged

