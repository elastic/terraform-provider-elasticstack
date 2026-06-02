# `elasticstack_kibana_security_entity_store_resolution_group` — Requirements

Data source implementation: `internal/kibana/security_entity_store_resolution_group`

## Purpose

Read the resolution group for a given entity in the Kibana Entity Store. A resolution group is the set of all entity identifiers linked to a single target ("golden") entity. This data source allows practitioners to reference resolution group membership in other Terraform resources or to verify current link state. Requires an enterprise license and `securitySolution` + `securitySolution-entity-analytics` route privileges.

## ADDED Requirements

### Requirement: Schema — identity (REQ-ESG-001)

The data source SHALL expose the following attributes:

- `id` (computed string): composite key `<space_id>/<entity_id>`. Set on read.
- `space_id` (optional, computed string): Kibana space identifier. Defaults to `"default"` when absent.
- `entity_id` (required string): entity identifier to look up the resolution group for. May be any entity in the group (target or alias).

#### Scenario: id is computed from space_id and entity_id

- GIVEN `space_id = "default"` and `entity_id = "user-123"`
- WHEN read completes
- THEN `id` SHALL equal `"default/user-123"`

#### Scenario: space_id defaults to "default"

- GIVEN `space_id` is not configured
- WHEN read runs
- THEN the provider SHALL treat `space_id` as `"default"` in all API calls and in the computed `id`

### Requirement: Schema — resolution_group_json (REQ-ESG-002)

The data source SHALL expose `resolution_group_json` as a computed `jsontypes.NormalizedType{}` attribute. On read, the provider SHALL populate this attribute with the normalized JSON body returned by `GET /api/security/entity_store/resolution/group?entity_id=<entity_id>`.

#### Scenario: resolution_group_json contains the full group response

- GIVEN an entity that belongs to a resolution group with multiple members
- WHEN the data source read runs
- THEN `resolution_group_json` SHALL be a non-empty, valid JSON string containing the API response

### Requirement: Schema — kibana_connection (REQ-ESG-003)

The data source SHALL expose `kibana_connection` as an optional single nested block using `schema.GetKbFWConnectionBlock()`, following the existing Plugin Framework convention for Kibana data sources.

#### Scenario: Data source uses provider-level Kibana connection by default

- GIVEN no `kibana_connection` block is configured on the data source
- WHEN the provider resolves the Kibana client
- THEN the provider SHALL use the provider-level Kibana connection defaults for the read API call

#### Scenario: Data source uses entity-local Kibana connection when configured

- GIVEN a `kibana_connection` block is configured with a URL and credentials
- WHEN the provider resolves the Kibana client
- THEN the provider SHALL use the entity-local connection for the resolution group read API call

### Requirement: Read (REQ-ESG-004)

On read, the provider SHALL:
1. Enforce `EnforceMinVersion("9.1.0")` — return an error diagnostic if the connected Elastic Stack is below 9.1.0.
2. Call `GET /api/security/entity_store/resolution/group?entity_id=<entity_id>`, applying `kibanautil.SpaceAwarePathRequestEditor(spaceID)`.
3. On a 404, return an error diagnostic indicating that no resolution group was found for the given entity ID (data sources return errors, not graceful removal, on not-found).
4. On a non-2xx non-404 response, return an error diagnostic with the HTTP status and response body.
5. On success, store the normalized JSON body in `resolution_group_json` and set `id = "<space_id>/<entity_id>"`.

#### Scenario: Read populates resolution_group_json

- GIVEN `entity_id = "user-123"` and the entity belongs to a resolution group
- WHEN read runs against a 9.1.0+ stack with an enterprise license
- THEN `resolution_group_json` SHALL be a non-empty, valid JSON string

#### Scenario: Read errors on 404

- GIVEN `entity_id = "user-999"` and no resolution group exists for this entity
- WHEN read runs
- THEN the provider SHALL return an error diagnostic (not silently produce an empty state)

#### Scenario: Read fails on unsupported version

- GIVEN the connected Elastic Stack is version 8.17.0
- WHEN read runs
- THEN the provider SHALL return a version-gate error diagnostic before calling the API

### Requirement: Space routing (REQ-ESG-005)

The `GET /api/security/entity_store/resolution/group` call SHALL apply `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a `RequestEditorFn`. For `space_id = "default"`, the path SHALL remain unchanged (no `/s/default/` prefix).

#### Scenario: Non-default space routes correctly

- GIVEN `space_id = "security-team"`
- WHEN read is called
- THEN the HTTP request path SHALL be `/s/security-team/api/security/entity_store/resolution/group`

#### Scenario: Default space omits space prefix

- GIVEN `space_id = "default"` (or absent)
- WHEN read is called
- THEN the HTTP request path SHALL be `/api/security/entity_store/resolution/group`
