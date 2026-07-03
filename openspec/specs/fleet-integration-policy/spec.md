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

After any create, update, or read operation, the resource SHALL populate the following fields from the API response: `id`, `policy_id`, `name`, `namespace`, `description`, `integration_name`, `integration_version`, `output_id`. The resource SHALL populate `agent_policy_id` from the API response when `agent_policy_id` was the originally configured attribute, and `agent_policy_ids` from the API response when `agent_policy_ids` was the originally configured attribute. When `space_ids` is returned by the API, the resource SHALL set it from the response; when not returned and `space_ids` was not originally set, the resource SHALL set it to null. The resource SHALL NOT map the API response's `enabled` field into Terraform state — the Kibana Fleet package-policy create/update API does not accept a top-level `enabled` value, the response field is always `true`, and the attribute is no longer part of the schema.

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

### Requirement: State upgrade — v0 to v3 (REQ-024)

The resource SHALL support state upgrade from schema version 0 to version 3 via intermediate v1 and v3 conversions. During v0→v1: `vars_json` and all input `vars_json`/`streams_json` string fields with empty string values SHALL be converted to normalized JSON null; non-empty values SHALL be wrapped in `jsontypes.Normalized`. The `agent_policy_ids` and `space_ids` fields absent in v0 SHALL be initialized to null. During v1→v3: the `input` list block SHALL be converted to an `inputs` map attribute keyed by `input_id`; each input's `streams_json` normalized JSON string SHALL be parsed and converted to the `streams` map structure; `vars_json` SHALL be migrated to the `VarsJSON` custom type with integration context attached. The legacy top-level `enabled` field present in v0/v1/v2 state SHALL be dropped, since the attribute no longer exists in v3.

#### Scenario: v0 empty vars_json becomes null

- GIVEN v0 state with `vars_json = ""`
- WHEN state upgrade to v3 runs
- THEN `vars_json` in v3 state SHALL be null

#### Scenario: v1 input list becomes v3 inputs map

- GIVEN v1 state with `input = [{ input_id = "logfile-1", enabled = true, vars_json = "{}", streams_json = null }]`
- WHEN state upgrade to v3 runs
- THEN `inputs` in v3 state SHALL be a map with key `"logfile-1"` containing `enabled = true` and `vars` populated from the v1 `vars_json`

#### Scenario: v0 enabled field dropped

- GIVEN v0 state with `enabled = false`
- WHEN state upgrade to v3 runs
- THEN the resulting v3 state SHALL NOT contain a top-level `enabled` attribute

### Requirement: State upgrade — v1 to v3 (REQ-025)

The resource SHALL support state upgrade from schema version 1 to version 3 directly. The v1→v3 upgrade SHALL apply the same `input` list to `inputs` map conversion and `streams_json` expansion described in REQ-024. All other fields (id, policy_id, name, namespace, agent_policy_id, agent_policy_ids, description, force, integration_name, integration_version, output_id, space_ids) SHALL be carried over unchanged. The legacy top-level `enabled` field SHALL be dropped.

#### Scenario: v1 to v3 direct upgrade

- GIVEN v1 state with an `input` list block containing one entry and `enabled = true`
- WHEN state upgrade to v3 runs directly (v1→v3 path)
- THEN `inputs` in v3 state SHALL be a map keyed by the entry's `input_id`, all other scalar fields SHALL be unchanged, and the resulting state SHALL NOT contain a top-level `enabled` attribute

### Requirement: State upgrade — v2 to v3 (REQ-026)

The resource SHALL support state upgrade from schema version 2 to version 3. The upgrade SHALL decode prior state using the prior v2 schema (which contained the now-removed top-level `enabled` attribute), drop the `enabled` value, and write all other v2 fields (id, policy_id, name, namespace, agent_policy_id, agent_policy_ids, description, force, integration_name, integration_version, output_id, space_ids, vars_json, inputs) into the v3 model unchanged. The upgrade SHALL NOT call the Fleet API.

#### Scenario: v2 to v3 drops enabled when true

- GIVEN v2 state with `enabled = true` and a populated `inputs` map
- WHEN state upgrade to v3 runs
- THEN the resulting v3 state SHALL preserve `inputs` and all other scalar fields and SHALL NOT contain a top-level `enabled` attribute

#### Scenario: v2 to v3 drops enabled when false

- GIVEN v2 state with `enabled = false` (a value that the provider previously ignored on writes)
- WHEN state upgrade to v3 runs
- THEN the resulting v3 state SHALL NOT contain a top-level `enabled` attribute and the upgrade SHALL succeed without error

### Requirement: Input-type package defaults extraction (REQ-NEW-INPUT-TYPE)

The defaults extractor SHALL handle both **integration-type** and **input-type** Fleet packages when extracting variable defaults from `PackageInfo`.

Integration-type packages declare `policyTemplates[].inputs[]` with a `type` and `vars` array per entry. Input-type packages declare a single top-level `input` string and a `vars` array at the policy-template level. For an input-type template, the extractor SHALL read `policyTemplate.input` as the input type, call `apiVars.defaults()` on `policyTemplate.vars`, and store the result keyed as `"{policyTemplate.name}-{policyTemplate.input}"` — the same key format used for integration-type packages. The resulting defaults SHALL then be combined with stream-level defaults from `apiDatastreams.defaults()` by the existing `packageInfoToDefaults()` function without further modification.

#### Scenario: Input-type package apply — no inconsistency error

- GIVEN an `elasticstack_fleet_integration_policy` resource targeting an input-type package (e.g. `gcp_pubsub`) where the user configures only a subset of the available variables
- WHEN Terraform applies the configuration
- THEN the provider SHALL NOT produce a `"Provider produced inconsistent result after apply"` error on `.inputs`
- AND the resulting Terraform state SHALL reflect the Kibana API response (including package-default vars) without a follow-up plan diff

#### Scenario: Input-type defaults extraction — known defaults present

- GIVEN a `PackageInfo` for an input-type package with at least one variable that carries a non-null `default` value (e.g. `subscription_type: "shared"`)
- WHEN `packageInfoToDefaults(pkg)` is called
- THEN the defaults map SHALL contain an entry for the expected input ID
- AND the entry's `Vars` JSON object SHALL include the defaulted variable
- AND the entry's `Vars` JSON object SHALL NOT include variables whose `default` is null and `multi` is false

#### Scenario: Input-type defaults extraction — non-defaulted vars omitted

- GIVEN a `PackageInfo` for an input-type package with at least one variable that has no `default` value and `multi: false` (e.g. `project_id`)
- WHEN `packageInfoToDefaults(pkg)` is called
- THEN the extracted defaults JSON SHALL NOT include that variable

### Requirement: Acceptance test coverage for input-type packages (REQ-NEW-INPUT-TYPE-ACC)

The acceptance test suite for `elasticstack_fleet_integration_policy` SHALL include at least one test case targeting an input-type package (e.g. `gcp_pubsub`) that configures a policy with only a subset of available vars, applies it, and then runs a plan that MUST produce no diff.

The test SHALL be skipped when the target Elastic Stack version is strictly below `8.10.0`.

#### Scenario: Acceptance test — apply and re-plan produce no diff

- GIVEN an input-type `gcp_pubsub` integration policy with only the required user-visible vars set (e.g. `project_id`, `subscription_name`, `topic`)
- WHEN the Terraform apply and a subsequent plan run
- THEN no `"inconsistent values for sensitive attribute"` error occurs during apply
- AND the subsequent plan SHALL show no changes

### Requirement: Stream vars semantic equality strips server-managed `data_stream.*` keys (REQ-DATASTREAM-VARS)

Fleet 9.5 injects server-managed `data_stream.type` and `data_stream.dataset` keys into the compiled stream `vars` of input-type packages. These keys are synthesised by Fleet's enrichment pipeline and are not user-configurable. The provider SHALL normalise stream `vars` by stripping all server-managed keys before performing semantic equality comparisons so that their presence in the API response does not trigger a `"Provider produced inconsistent result after apply"` error.

The server-managed keys that SHALL be stripped are:

- `data_stream.type`
- `data_stream.dataset`

Stripping SHALL occur on both sides of the comparison (plan-side and API-side stream `vars`) immediately before the `StringSemanticEquals` call in the `compareStreams` function. The stripped values are used only for comparison; they SHALL NOT be removed from the persisted state or from the value sent to the API. If the input JSON is null or unknown, the stripping helper SHALL return the input unchanged. Diag errors from the stripping helper SHALL abort the comparison and propagate to the caller.

The stripping is applied at stream level only. Input-level `vars` and the top-level `vars_json` attribute are unaffected.

#### Scenario: Stream vars with injected `data_stream.type` are semantically equal to plan-side vars without it

- GIVEN a stream whose plan-side `vars` JSON is `{"project_id":"my-project","subscription_name":"my-sub","tags":["forwarded"],"topic":"my-topic"}`
- AND the API-returned stream `vars` is `{"data_stream.dataset":"gcp_pubsub.generic","data_stream.type":"logs","project_id":"my-project","subscription_name":"my-sub","tags":["forwarded"],"topic":"my-topic"}`
- WHEN `compareStreams` evaluates semantic equality
- THEN the comparison SHALL return `true` (semantically equal)
- AND Terraform SHALL NOT produce a `"Provider produced inconsistent result after apply"` error

#### Scenario: Stream vars without server-managed keys are compared normally

- GIVEN a stream whose plan-side `vars` JSON is `{"threshold":42}` and the API-side `vars` is `{"threshold":99}`
- WHEN `compareStreams` evaluates semantic equality
- THEN the comparison SHALL return `false` (the user-defined key differs)

#### Scenario: Strip helper is a no-op on null/unknown vars

- GIVEN a stream `vars` value that is null or unknown
- WHEN the server-managed key stripping helper is called
- THEN the input SHALL be returned unchanged with no diagnostics

### Requirement: `defaults` attribute null ⇄ populated-object transition is semantically equal (REQ-DEFAULTS-COMPUTED)

Fleet 9.5 now populates a `defaults` block in the package policy GET response for input-type packages. Prior to 9.5, this block was not returned. The `defaults` attribute is purely computed from package information — the user never configures it directly. When the planned value of `defaults` is `null` (because the attribute was absent from the plan) and the API returns a populated `defaults` object, the provider SHALL treat this transition as semantically equal and SHALL NOT produce a `"Provider produced inconsistent result after apply"` error.

The semantic equality rule is applied in `InputValue.ObjectSemanticEquals`: if either side's `defaults` is null or unknown, the `defaults` component SHALL be skipped in the equality check (treated as equal). When both sides have a fully known `defaults`, the equality check SHALL also treat them as semantically equal (since `defaults` is purely server-managed and any value returned by the API is a valid resolved state for a null plan). This ensures robustness against future Fleet enrichment changes that alter the `defaults` content across applies.

No schema version bump is required — `defaults` is already `Computed: true` in the schema. No state upgrader is required. The fix is entirely within the semantic-equality comparison layer.

#### Scenario: `defaults` goes from null in plan to populated object after apply

- GIVEN a plan where `inputs["gcp-gcp-pubsub"].defaults` is `null`
- AND the Fleet API returns a populated `defaults` object after apply: `{"streams":{"gcp_pubsub.gcp":{"enabled":true,"vars":{"data_stream.dataset":"gcp_pubsub.generic","tags":["forwarded"]}}},"vars":null}`
- WHEN `InputValue.ObjectSemanticEquals` evaluates the input
- THEN the comparison SHALL return `true` (semantically equal)
- AND Terraform SHALL NOT produce a `"Provider produced inconsistent result after apply"` error on the `defaults` attribute

#### Scenario: Fully-known `defaults` on both sides does not block equality

- GIVEN a plan where `inputs["foo"].defaults` is a known populated object
- AND the API returns a `defaults` object with different content
- WHEN `InputValue.ObjectSemanticEquals` evaluates
- THEN the comparison SHALL treat the `defaults` component as semantically equal (since `defaults` is server-managed) and SHALL continue to compare `vars` and `streams`

#### Scenario: 9.4 behavior unchanged (defaults remains null)

- GIVEN a Fleet 9.4 API response where `defaults` is not returned
- WHEN the resource is read and `defaults` remains null in state
- THEN no plan diff SHALL appear on subsequent plans
- AND the stripping and defaults-equality fixes SHALL be no-ops (the keys are absent; null ⇄ null is already equal)

### Requirement: Acceptance test coverage for Fleet 9.5 enrichment (REQ-DATASTREAM-ACC)

The acceptance test suite SHALL include (or continue to pass, if already present) test cases that exercise both failure modes fixed by this change:

1. `TestAccResourceIntegrationPolicyGCPPubSub` — verifies that an input-type package policy can be applied and subsequently planned with no diff on a Fleet 9.5 stack, where stream `vars` contain injected `data_stream.*` keys and `defaults` is populated in the API response.
2. `TestAccResourceIntegrationPolicySecrets` (both subtests) — verifies that secret-valued input vars and multi-valued secrets on an input-type package policy apply cleanly on Fleet 9.5 without post-apply inconsistency.

Both tests SHALL be skipped when the target Elastic Stack version is strictly below `9.5.0`.

#### Scenario: GCP PubSub policy apply and re-plan on 9.5 produce no diff

- GIVEN `TestAccResourceIntegrationPolicyGCPPubSub` running against a 9.5.0+ stack
- WHEN the Terraform apply and a subsequent plan run
- THEN no `"Provider produced inconsistent result after apply"` error occurs
- AND the subsequent plan SHALL show no changes

