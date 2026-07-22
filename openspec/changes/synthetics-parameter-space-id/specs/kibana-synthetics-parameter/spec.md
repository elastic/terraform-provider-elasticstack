## CHANGED Requirements

### Requirement: Space-scoped Synthetics Parameters API (REQ-001)

_Replaces existing REQ-001._

The resource SHALL manage Synthetics parameters through Kibana's Synthetics Parameters API using the space-aware path pattern: create via `POST /s/{space_id}/api/synthetics/params`, read via `GET /s/{space_id}/api/synthetics/params/{id}`, update via `PUT /s/{space_id}/api/synthetics/params/{id}`, and delete via `DELETE /s/{space_id}/api/synthetics/params` (or `DELETE /s/{space_id}/api/synthetics/params/{id}` for Kibana ≥ 8.17.0). When `space_id` is `"default"` or empty, the path SHALL remain unchanged (no `/s/{space_id}` prefix injected). All operations SHALL use `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to rewrite the request URL path before the call is sent.

#### Scenario: CRUD uses space-aware Synthetics Parameters APIs

- GIVEN a managed Synthetics parameter with `space_id = "my-space"`
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the corresponding Kibana Synthetics Parameters API with the `/s/my-space/` path prefix

#### Scenario: Default space uses unscoped path

- GIVEN a managed Synthetics parameter with `space_id = "default"` (or `space_id` unset)
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the Kibana Synthetics Parameters API without a space path prefix

### Requirement: Import by composite or bare id (REQ-004)

_Replaces existing REQ-004._

The resource SHALL support Terraform import using either a composite `<space_id>/<parameter_uuid>` or a bare `<parameter_uuid>` as the import identifier. When a bare UUID is provided (no `/`), the provider SHALL treat it as belonging to the default space and SHALL populate `space_id = "default"` in state. When a composite identifier is provided, the provider SHALL extract `<space_id>` and `<parameter_uuid>` and populate both `id` and `space_id` in state. On import, a composite `id` where the resource-UUID segment is empty SHALL return an error diagnostic.

#### Scenario: Import with composite id

- GIVEN an import id in the format `<space_id>/<parameter_uuid>`
- WHEN import runs and read is performed
- THEN the provider SHALL set `space_id` to `<space_id>` and `id` to `<space_id>/<parameter_uuid>` in state, and SHALL call the Synthetics Parameters API under the correct space path

#### Scenario: Import with bare UUID

- GIVEN an import id that contains no `/`
- WHEN import runs
- THEN the provider SHALL treat the id as a default-space parameter, set `space_id = "default"`, and set `id = "default/<parameter_uuid>"` in state

#### Scenario: Import with empty resource segment

- GIVEN an import id of the form `<space_id>/` (empty UUID segment)
- WHEN import runs
- THEN the provider SHALL return an error diagnostic

## ADDED Requirements

### Requirement: `space_id` attribute (REQ-012)

The resource SHALL expose an **optional, computed** `space_id` string attribute with `UseStateForUnknown` and `RequiresReplace` plan modifiers. When `space_id` is not configured, the attribute SHALL be computed and set to `"default"` after the first successful create. The provider SHALL validate that `space_id` is a known, non-empty value before invoking create or update.

#### Scenario: `space_id` defaults to "default"

- GIVEN a parameter configured without `space_id`
- WHEN create runs
- THEN `space_id` SHALL be set to `"default"` in state

#### Scenario: `space_id` routes to named space

- GIVEN a parameter configured with `space_id = "ops-team"`
- WHEN create runs
- THEN the provider SHALL POST to `/s/ops-team/api/synthetics/params` and store `space_id = "ops-team"` in state

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed parameter with `space_id = "default"`
- WHEN configuration changes `space_id` to `"ops-team"`
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Composite `id` encoding (REQ-013)

The resource `id` SHALL be stored as a composite `<space_id>/<parameter_uuid>` string after a successful create. This encoding enables `resolveKibanaResourceIdentity` to recover both the UUID and the space from state without requiring a separate read of the `space_id` field. On read-after-write, `modelFromOAPI` SHALL accept the `spaceID` from the write context and assemble the composite `id`.

#### Scenario: `id` set to composite after create

- GIVEN a successful parameter create in space `"my-space"`
- WHEN Kibana returns the new parameter UUID `"abc-123"`
- THEN the provider SHALL store `id = "my-space/abc-123"` in state

#### Scenario: `id` set to composite in default space

- GIVEN a successful parameter create with no explicit `space_id`
- WHEN Kibana returns the new parameter UUID `"abc-123"`
- THEN the provider SHALL store `id = "default/abc-123"` in state

### Requirement: Schema version bump and state migration v0→v1 (REQ-014)

The resource schema version SHALL be bumped from **0** to **1**. A `StateUpgraders` entry for version **0** SHALL rewrite the `id` attribute from a bare UUID to `"default/<uuid>"` and SHALL add `"space_id": "default"` to the migrated state. This migration SHALL be non-destructive: no resource is destroyed or recreated as part of the state upgrade.

#### Scenario: State migration rewrites bare UUID

- GIVEN a Terraform state file containing a `elasticstack_kibana_synthetics_parameter` resource with a bare UUID `id` (schema version 0)
- WHEN the provider runs for the first time at schema version 1
- THEN the provider SHALL upgrade the state: `id` becomes `"default/<uuid>"` and `space_id` is set to `"default"`, with no plan difference caused by the migration alone
