# kibana-security-entity-store-entity-link Specification

## Purpose
TBD - created by archiving change kibana-security-entity-store-resolution-link. Update Purpose after archive.
## Requirements
### Requirement: Schema — identity (REQ-ESL-001)

The resource SHALL expose the following identity attributes:

- `id` (computed string, `UseStateForUnknown`): composite key `<space_id>/<target_id>`. Set on create; stable for the lifetime of the resource.
- `space_id` (optional, computed string, `RequiresReplace`): Kibana space identifier. Defaults to `"default"` when absent. Computed so that `terraform import` can populate it from the import ID.
- `target_id` (required string, `RequiresReplace`): entity identifier that linked entities resolve to.

#### Scenario: id is computed from space_id and target_id

- GIVEN `space_id = "default"` and `target_id = "user-123"`
- WHEN create succeeds
- THEN `id` SHALL equal `"default/user-123"`

#### Scenario: space_id defaults to "default"

- GIVEN `space_id` is not configured
- WHEN apply runs
- THEN the provider SHALL treat `space_id` as `"default"` in all API calls and in the computed `id`

### Requirement: Schema — entity_ids (REQ-ESL-002)

The resource SHALL expose `entity_ids` as a required `schema.SetAttribute` of string. The set SHALL contain between 1 and 1000 items (inclusive). The provider SHALL enforce these bounds at plan time via schema validators.

#### Scenario: entity_ids below minimum

- GIVEN `entity_ids` is an empty set
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic indicating at least 1 entity ID is required

#### Scenario: entity_ids above maximum

- GIVEN `entity_ids` contains 1001 or more items
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic indicating the maximum is 1000

#### Scenario: entity_ids valid range

- GIVEN `entity_ids` contains between 1 and 1000 items
- WHEN Terraform validates configuration
- THEN the provider SHALL accept the configuration

### Requirement: Validation — self-link guard (REQ-ESL-003)

The provider SHALL reject at plan time any configuration where `target_id` appears in `entity_ids`. A custom plan-time validator SHALL check membership and return a diagnostic describing the constraint.

#### Scenario: self-link rejected

- GIVEN `target_id = "user-123"` and `entity_ids = ["user-123", "user-456"]`
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic stating that `target_id` must not appear in `entity_ids`

#### Scenario: no self-link passes validation

- GIVEN `target_id = "user-123"` and `entity_ids = ["user-456", "user-789"]`
- WHEN Terraform validates configuration
- THEN the provider SHALL accept the configuration

### Requirement: Schema — resolution_group_json (REQ-ESL-004)

The resource SHALL expose `resolution_group_json` as a computed `jsontypes.NormalizedType{}` attribute. On every successful read the provider SHALL populate this attribute with the normalized JSON body returned by `GET /api/security/entity_store/resolution/group?entity_id=<target_id>`.

#### Scenario: resolution_group_json populated after create

- GIVEN a resource is created with `target_id = "user-123"` and `entity_ids = ["user-456"]`
- WHEN create completes and read is called
- THEN `resolution_group_json` SHALL be a non-empty, valid JSON string

### Requirement: Schema — kibana_connection (REQ-ESL-005)

The resource SHALL expose `kibana_connection` as an optional single nested block using `schema.GetKbFWConnectionBlock()`, following the existing Plugin Framework convention for Kibana resources.

#### Scenario: Resource uses provider-level Kibana connection by default

- GIVEN no `kibana_connection` block is configured on the resource
- WHEN the provider resolves the Kibana client
- THEN the provider SHALL use the provider-level Kibana connection defaults for all API calls

#### Scenario: Resource uses entity-local Kibana connection when configured

- GIVEN a `kibana_connection` block is configured with a URL and credentials
- WHEN the provider resolves the Kibana client
- THEN the provider SHALL use the entity-local connection for all link, unlink, and read API calls

### Requirement: Create (REQ-ESL-006)

On create, the provider SHALL:
1. Enforce `EnforceMinVersion("9.1.0")` — return an error diagnostic if the connected Elastic Stack is below 9.1.0.
2. Call `POST /api/security/entity_store/resolution/link` with body `{"target_id": "<target_id>", "entity_ids": [<entity_ids>]}`, applying `kibanautil.SpaceAwarePathRequestEditor(spaceID)`.
3. On a non-2xx response, return an error diagnostic with the HTTP status and response body.
4. Call the internal `read()` function to populate final state (including `resolution_group_json`).

#### Scenario: Create links entities to target

- GIVEN `target_id = "user-123"` and `entity_ids = ["user-456", "user-789"]`
- WHEN create runs against a 9.1.0+ stack with an enterprise license
- THEN the link request SHALL be sent with `entity_ids = ["user-456", "user-789"]` and `target_id = "user-123"`, and state SHALL reflect a successful link

#### Scenario: Create fails on unsupported version

- GIVEN the connected Elastic Stack is version 8.17.0
- WHEN create runs
- THEN the provider SHALL return a version-gate error diagnostic before calling the API

### Requirement: Read (REQ-ESL-007)

On read, the provider SHALL:
1. Call `GET /api/security/entity_store/resolution/group?entity_id=<target_id>`, applying `kibanautil.SpaceAwarePathRequestEditor(spaceID)`.
2. When the read is executed immediately after a successful link or unlink operation to populate final state, retry the GET with exponential back-off until the expected changes are visible or a bounded timeout of approximately 2 seconds is reached.
3. On a 404, call `resp.State.RemoveResource(ctx)` to remove the resource from state (out-of-band deletion).
4. On a non-2xx non-404 response, return an error diagnostic.
5. On success, store the raw response body (normalized) in `resolution_group_json`.
6. Emit a warning diagnostic if any managed `entity_ids` from state are absent from the API response (indicating out-of-band removal), without removing the resource from state.

#### Scenario: Read removes resource on 404

- GIVEN the resolution group for `target_id` no longer exists in Kibana
- WHEN read runs
- THEN the provider SHALL remove the resource from state without error

#### Scenario: Read warns on missing entity_ids

- GIVEN one of the managed `entity_ids` was removed from the resolution group out-of-band
- WHEN read runs
- THEN the provider SHALL emit a warning diagnostic and retain the resource in state with the API-returned `resolution_group_json`

#### Scenario: Read retries after link or unlink before final state is available

- GIVEN a link or unlink operation has just completed successfully
- AND the Entity Store change is not yet visible because the next index refresh has not completed
- WHEN the provider reads the resolution group to populate final state
- THEN the provider SHALL retry the GET with exponential back-off until the expected changes are reflected or a bounded timeout of approximately 2 seconds is reached

### Requirement: Update (REQ-ESL-008)

On update, the provider SHALL compute the set difference between the new `entity_ids` (plan) and the current `entity_ids` (state):
- Added IDs: call `POST /api/security/entity_store/resolution/link` with `{target_id, added_ids}`.
- Removed IDs: call `POST /api/security/entity_store/resolution/unlink` with `{entity_ids: removed_ids}`.

After link/unlink operations, call the internal `read()` function to populate final state.

If either operation fails, return an error diagnostic. The provider SHALL NOT treat a partial update as success.

#### Scenario: Update links new and unlinks removed IDs

- GIVEN state has `entity_ids = ["user-456"]` and plan has `entity_ids = ["user-456", "user-789"]`
- WHEN update runs
- THEN the provider SHALL call link with `["user-789"]` and SHALL NOT call unlink, then read final state

#### Scenario: Update unlinks removed IDs only

- GIVEN state has `entity_ids = ["user-456", "user-789"]` and plan has `entity_ids = ["user-456"]`
- WHEN update runs
- THEN the provider SHALL call unlink with `["user-789"]` and SHALL NOT call link, then read final state

### Requirement: Delete (REQ-ESL-009)

On delete, the provider SHALL call `POST /api/security/entity_store/resolution/unlink` with `{entity_ids: <managed entity_ids>}`, applying `kibanautil.SpaceAwarePathRequestEditor(spaceID)`. The provider SHALL treat a 404 response as already-deleted (no error). The provider SHALL NOT delete entities from the Entity Store, only remove their resolution link.

#### Scenario: Delete unlinks managed entity_ids

- GIVEN a resource with `entity_ids = ["user-456", "user-789"]`
- WHEN destroy runs
- THEN the provider SHALL call unlink with `entity_ids = ["user-456", "user-789"]`

#### Scenario: Delete treats 404 as success

- GIVEN the resolution group was removed out-of-band before destroy
- WHEN destroy runs and the API returns 404
- THEN the provider SHALL not return an error

### Requirement: Import (REQ-ESL-010)

The resource SHALL support `terraform import` using the ID format `<space_id>/<target_id>` (e.g. `default/user-123`). On import, the provider SHALL:
1. Parse `space_id` and `target_id` from the import ID.
2. Call `GetSecurityEntityStoreResolutionGroup` with `entity_id = target_id`.
3. Populate `resolution_group_json` from the response body.
4. Populate `entity_ids` from the entity identifiers returned in the resolution group.

#### Scenario: Import reconstructs full state

- GIVEN `terraform import elasticstack_kibana_security_entity_store_entity_link.r default/user-123`
- WHEN import runs and the resolution group exists
- THEN state SHALL contain `space_id = "default"`, `target_id = "user-123"`, `entity_ids` populated from the API, and `resolution_group_json`

### Requirement: Space routing (REQ-ESL-011)

All API calls (link, unlink, get resolution group) SHALL apply `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a `RequestEditorFn`. For `space_id = "default"`, the path SHALL remain unchanged (no `/s/default/` prefix, per `BuildSpaceAwarePath` semantics).

#### Scenario: Non-default space routes correctly

- GIVEN `space_id = "security-team"`
- WHEN link is called
- THEN the HTTP request path SHALL be `/s/security-team/api/security/entity_store/resolution/link`

#### Scenario: Default space omits space prefix

- GIVEN `space_id = "default"` (or absent)
- WHEN link is called
- THEN the HTTP request path SHALL be `/api/security/entity_store/resolution/link`

### Requirement: Acceptance tests for entity-link enforce entity store isolation (REQ-ESL-TEST-ISOLATION-001)

Every acceptance test in the entity-link package that manages the entity store resource SHALL
register a `t.Cleanup` function that uninstalls the entity store and waits for `not_installed`
state (using the same shared helper as the entity store package, as specified in
`kibana-security-entity-store` delta REQ-TEST-ISOLATION-001). This ensures that entity-link
tests do not leave residual entity types in the singleton store that contaminate subsequent tests.

#### Scenario: Entity-link test cleanup leaves store in not_installed state

- GIVEN `TestAccResourceSecurityEntityStoreEntityLink` registers `t.Cleanup(cleanupEntityStore)`
- WHEN the test body completes (success or failure)
- THEN `cleanupEntityStore` SHALL run and poll until the entity store reports `not_installed`
- AND the next test in the suite SHALL find the store in a clean state

