## Context

The Terraform Plugin Framework supports `WriteOnly: true` on `resource/schema.StringAttribute` (PF ≥ 1.11). Write-only attributes are accepted during create/update but are never stored in state and are not returned on read. They are available in the config and plan during `ModifyPlan` and CRUD operations, but absent from the state object after apply.

The provider already has a working pattern for write-only secret attributes:

- `elasticstack_elasticsearch_security_user` exposes `password_wo` alongside `password`, with `stringvalidator.PreferWriteOnlyAttribute` on the plain companion and hard `ConflictsWith` between `password`, `password_hash`, and `password_wo`. Note that this resource relies on `password_wo_version` for rotation; this change instead uses the `writeonlyhash` private-state mechanism and deliberately omits `_wo_version` attributes per @tobio.
- The `action_connection` block (used by provider-defined actions) already marks `password`, `api_key`, `bearer_token`, `es_client_authentication`, and `key_data` directly as `WriteOnly: true` — the action schema is never stored in state, so it can do this cleanly.
- The `writeonlyhash` utility (introduced in the `fleet-cloud-connector` change) provides bcrypt-backed drift detection without requiring user-managed version companions.

The **per-resource** `elasticsearch_connection` and `kibana_connection` blocks are different: they use `provider/schema` types currently, which do not support `WriteOnly`. Additionally, the blocks' plain credential attributes are already in provider state (for resources that use them), so adding `WriteOnly` to the existing attributes would be a breaking schema change. The correct solution is to add `_wo` companion attributes on a new `resource/schema`-typed block function.

## Design decisions

### 1. Companion attributes with `_wo` suffix, not direct write-only on existing attributes

Adding `WriteOnly: true` to the existing `password`, `api_key`, etc. attributes on the provider-schema block would be a breaking change: existing state would have those fields, but the framework would no longer accept them on read. The companion approach (`password_wo`, `api_key_wo`, etc.) preserves backward compatibility.

### 2. Split block functions: `GetEsResourceConnectionBlock` / `GetKbResourceConnectionBlock`

A new function using `resource/schema` types is required. The existing `GetEsFWConnectionBlock` / `GetKbFWConnectionBlock` (using `provider/schema`) are retained for provider-level, data-source, ephemeral-resource, and action usage. The new functions mirror the structure but use `resource/schema.StringAttribute` and add the `_wo` companions.

### 3. New structs: `ElasticsearchResourceConnection` / `KibanaResourceConnection`

The new blocks require new structs with the additional `_wo` fields. The existing `ElasticsearchConnection` / `KibanaConnection` structs in `internal/clients/config/provider.go` are used for provider-level config decoding and remain unchanged.

### 4. Credential preference: `_wo` wins when both are set (defensive path only)

When a practitioner sets both `password` and `password_wo`, the hard `ConflictsWith` validator rejects the configuration before the client factory runs. The factory still implements defensive `_wo`-preference logic: if `_wo` is non-empty after decoding, use it; otherwise use the plain value. This is consistent with the security user resource behavior and protects against accidental relaxation of the conflict validator.

`GetElasticsearchClient` and `GetKibanaClient` are updated in place to decode the incoming `types.List` into the resource-variant structs (`[]ElasticsearchResourceConnection` and `[]KibanaResourceConnection`). A small bridging step copies `_wo` values into the corresponding plain fields when set, producing `[]ElasticsearchConnection` / `[]KibanaConnection`. Those resolved slices are then fed into the existing config builders:

- For Elasticsearch: `newBaseConfigFromFramework` and `newElasticsearchConfigFromFramework` work unchanged.
- For Kibana: `NewFromFrameworkKibanaResource` / `newKibanaOapiConfigFromFramework` work unchanged.

Because the substitution happens before `applyAuthOverride` / `clearConflictingAuth` / `withEnvironmentOverrides`, the existing auth clearing and inheritance logic in `internal/clients/config/auth.go` and `internal/clients/config/kibana_oapi.go` requires no modification. All existing callers continue to use `GetElasticsearchClient` / `GetKibanaClient`; no separate resource-variant client methods are introduced.

### 5. No `_wo_version` attributes

Explicitly excluded by @tobio. Drift detection uses `writeonlyhash` exclusively. The private-state hash approach detects silent credential rotations in config without requiring the practitioner to bump a version counter.

### 6. Null-list helpers for resource variants

ImportState and state upgraders across the codebase call `ElasticsearchConnectionNullList()` / `KibanaConnectionNullList()`. These helpers produce a null list typed to the *provider-schema* block object type. Resources that migrate to the new resource-schema block will need to call the new `ElasticsearchResourceConnectionNullList()` / `KibanaResourceConnectionNullList()` helpers instead. Existing callers of the old helpers may remain on the old helpers as long as they continue to use the provider-schema block (no forced migration).

### 7. Fleet connection is out of scope

`fleet_connection` is provider-level only; there are no managed resources that embed it. No changes needed.

### 8. ModifyPlan placement

`ModifyPlan` is implemented at the envelope level (`resource_envelope.go`, `kibana_resource_envelope.go`) rather than per-resource, so all resources that use the envelopes benefit automatically. The envelope already owns the connection block schema injection, making it the natural owner of connection-block drift detection.

The envelope `ModifyPlan` is a no-op when no `_wo` attribute is configured, so it does not add surprising plan-time behavior to existing envelope resources.

### 9. Per-resource-type `writeonlyhash` salt

Each concrete resource type constructs its own `Hasher` using the fully-qualified Terraform resource type name (`elasticstack_elasticsearch_<name>` / `elasticstack_kibana_<name>`), derived from the envelope name argument in `NewElasticsearchResource` / `NewKibanaResource`. This matches the `writeonlyhash` helper contract and avoids cross-resource-type hash correlation.

### 10. Concrete resources with their own `ModifyPlan` must delegate to the envelope

Several concrete resources already implement `ModifyPlan` (e.g., `elasticstack_elasticsearch_index`, `elasticstack_kibana_space`, `elasticstack_elasticsearch_connector`). Because Go method promotion is shadowed by an outer method, these resources must explicitly call the envelope's `ModifyPlan` from their own `ModifyPlan` to retain `_wo` drift detection.

### 11. Resources that use `ResourceBase` directly are intentionally unaffected

Resources that embed `ResourceBase` directly (not the envelopes) and call `GetElasticsearchClient` / `GetKibanaClient` directly — such as `elasticstack_fleet_agent_policy`, `elasticstack_fleet_integration_policy`, `elasticstack_kibana_import_saved_objects`, and `elasticstack_apm_agent_configuration` — continue to use the provider-schema block and do not gain `_wo` attributes. Converting them is out of scope for this change.

## Open questions

*(All questions from the prior run were answered by @tobio — none remain.)*

## Out of scope

- `fleet_connection` block — provider-level only, no managed resources use it.
- Provider-level `elasticsearch` / `kibana` blocks — not written to state; write-only semantics would be redundant.
- Data source connection blocks — no resource lifecycle; write-only attributes are not meaningful for data sources.
- Ephemeral resource connection blocks — ephemeral resources never persist state, so write-only semantics are redundant.
- Resources that embed `ResourceBase` directly and do not use the resource envelopes.
- `_wo_version` / version companion attributes — explicitly excluded by @tobio.

## Affected files (expected)

| File | Change |
|------|--------|
| `internal/schema/connection.go` | Add `GetEsResourceConnectionBlock()`, `GetKbResourceConnectionBlock()`, `ElasticsearchResourceConnectionNullList()`, `ElasticsearchResourceConnectionObjectType()`, `KibanaResourceConnectionNullList()`, `KibanaResourceConnectionObjectType()`, new fallback maps and object-type functions for resource variants |
| `internal/clients/config/provider.go` | Add `ElasticsearchResourceConnection` and `KibanaResourceConnection` structs with `_wo` fields |
| `internal/clients/provider_client_factory.go` | Update `GetElasticsearchClient` / `GetKibanaClient` to decode into `[]ElasticsearchResourceConnection` / `[]KibanaResourceConnection`, apply `_wo`-over-plain preference, and build the scoped client from the resolved `[]ElasticsearchConnection` / `[]KibanaConnection` using the existing config builders |
| `internal/entitycore/resource_envelope.go` | Switch `GetEsFWConnectionBlock()` → `GetEsResourceConnectionBlock()`; wire `ModifyPlan` for `_wo` drift detection; update null-list call if ImportState uses it |
| `internal/entitycore/kibana_resource_envelope.go` | Switch `GetKbFWConnectionBlock()` → `GetKbResourceConnectionBlock()`; wire `ModifyPlan` for `_wo` drift detection; update null-list call if ImportState uses it |
