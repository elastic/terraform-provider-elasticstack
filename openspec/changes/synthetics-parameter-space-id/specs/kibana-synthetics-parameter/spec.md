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

The resource SHALL expose an **optional, computed** `space_id` string attribute defined via the canonical `kbschema.ResourceSpaceIDAttribute()` helper: `Default` of `"default"` (`clients.DefaultSpaceID`), with `UseStateForUnknown` and `RequiresReplace` plan modifiers. When `space_id` is not configured, the schema default SHALL materialize `"default"` before create or update. The model SHALL NOT implement `KibanaUnscopedSpace`; the envelope's normal non-empty `space_id` validation SHALL apply and SHALL be satisfied by the schema default.

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

The resource `id` SHALL be stored as a composite `<space_id>/<parameter_uuid>` string (via `clients.CompositeID`) after a successful create or update. This encoding enables `resolveKibanaResourceIdentity` to recover both the UUID and the space from state without requiring a separate read of the `space_id` field. On read-after-write, `modelFromOAPI` SHALL accept the `spaceID` from the write context and assemble the composite `id`. A legacy bare-UUID `id` (no `/`) in prior state SHALL continue to resolve correctly: it SHALL fall back to the bare UUID as the resource id and the default space, and SHALL be rewritten to composite form on the next create, update, or refresh. No schema-version bump or `StateUpgraders` migration SHALL be introduced.

#### Scenario: `id` set to composite after create

- GIVEN a successful parameter create in space `"my-space"`
- WHEN Kibana returns the new parameter UUID `"abc-123"`
- THEN the provider SHALL store `id = "my-space/abc-123"` in state

#### Scenario: `id` set to composite in default space

- GIVEN a successful parameter create with no explicit `space_id`
- WHEN Kibana returns the new parameter UUID `"abc-123"`
- THEN the provider SHALL store `id = "default/abc-123"` in state

### Requirement: Backward compatibility for legacy bare-UUID state (REQ-014)

The resource SHALL remain compatible with existing state that stores a bare-UUID `id` without any schema-version bump or `StateUpgraders` migration. A legacy default-space parameter SHALL continue to be readable, updatable, and deletable: `resolveKibanaResourceIdentity` SHALL parse the bare UUID (no `/`) as the resource id with the default space, and CRUD SHALL route to the unscoped path. The bare-UUID `id` SHALL be rewritten to the composite `"default/<uuid>"` form on the next successful create, update, or refresh, without destroying or recreating the resource.

#### Scenario: Legacy bare-UUID state resolves to default space

- GIVEN existing state containing a `elasticstack_kibana_synthetics_parameter` with a bare UUID `id` and no `space_id`
- WHEN read, update, or delete runs
- THEN the provider SHALL treat it as a default-space parameter, route to the unscoped Synthetics Parameters path, and (on write/refresh) rewrite `id` to `"default/<uuid>"` with no destroy/recreate

### Requirement: Non-default space at existing API baseline (REQ-015)

Space-prefixed Synthetics Parameters API paths for non-default Kibana spaces SHALL NOT require a Kibana version floor above the resource's existing **8.12.0** API baseline. The provider SHALL NOT introduce a `GetVersionRequirements` check solely to gate non-default `space_id` routing (contrast Synthetics private location, which requires a higher stack version for non-default space).

#### Scenario: Non-default space without extra version gate

- GIVEN a parameter configured with a non-default `space_id`
- WHEN create, read, update, or delete runs against a Kibana version that already satisfies the resource's existing minimum
- THEN the provider SHALL route via the space-prefixed Synthetics Parameters API path and SHALL NOT fail with a version diagnostic introduced solely for non-default space routing
