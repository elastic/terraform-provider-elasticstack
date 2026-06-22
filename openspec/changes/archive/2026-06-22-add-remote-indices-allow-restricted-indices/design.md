## Context

Both `elasticstack_elasticsearch_security_role` and `elasticstack_kibana_security_role` model Elasticsearch index privileges via nested `indices` and `remote_indices` blocks. The Elasticsearch security role resource already exposes `allow_restricted_indices` on `indices` (optional, computed, with `UseStateForUnknown`), and maps it through `IndexPermsData` / `estypes.IndicesPrivileges`. The `remote_indices` block omits the attribute from schema and state even though:

- `estypes.RemoteIndicesPrivileges` includes `AllowRestrictedIndices`
- `kibanaoapi.SecurityRoleESRemoteIndex` already has `AllowRestrictedIndices *bool`
- `toAPIModel` for Elasticsearch roles reads `idx.AllowRestrictedIndices` from `indexPermissionsToAPIModel` (always nil) rather than a dedicated remote field

The Kibana security role never maps `allow_restricted_indices` for any index type in expand/flatten, though the API client structs are ready.

## Goals / Non-Goals

**Goals:**

- Expose `allow_restricted_indices` on `remote_indices` for both security role resources and their data sources.
- Mirror `indices` semantics on the Elasticsearch resource: optional + computed on the resource, computed on read, `UseStateForUnknown` plan modifier, same description text.
- Round-trip the value through create, update, and read; cover with acceptance tests.

**Non-Goals:**

- Adding `allow_restricted_indices` to `elasticsearch.indices` on the Kibana security role (out of scope for this change).
- Changing defaulting behavior for `indices.allow_restricted_indices` on the Elasticsearch resource.
- Schema version bumps or state upgrade migrations (additive optional attribute only).

## Decisions

### 1. Reuse `IndexPermsData` shape for Elasticsearch remote indices

Add `AllowRestrictedIndices types.Bool` to `RemoteIndexPermsData` (same as `IndexPermsData`) rather than embedding `IndexPermsData` wholesale (which would pull in unused structure). Wire it in `toAPIModel` / `fromAPIModel` and the data source flattener alongside existing remote index fields.

**Alternative considered:** Promote `AllowRestrictedIndices` into `CommonIndexPermsData` — rejected because only `indices` and `remote_indices` use it, not shared query-only helpers.

### 2. Schema parity with existing `indices` block (Elasticsearch)

Add `attrAllowRestrictedIndices` to the `remote_indices` nested block in `schema.go` and data source schema with identical modifiers and embedded description (`descriptions/allow_restricted_indices.md`). `getRemoteIndexPermsAttrTypes()` derives from the block definition, so attr types update automatically.

### 3. Kibana: attribute on `remote_indices` only via shared expand helper

Extend `expandedEntry` with `AllowRestrictedIndices *bool`, read it in `expandEntryCommon`, and set it on `SecurityRoleESRemoteIndex` in `expandRemoteEntry`. In `flattenRemoteIndicesResource`, map API pointer to `types.Bool` (null when absent). Add the attribute to `remoteIndicesResourceBlock()` and data source schema; update `esRemoteIndexResourceAttrTypes()`.

Use optional (not computed) on the Kibana resource to match other remote index optional fields (`query`, `field_security`). No `UseStateForUnknown` unless acceptance tests show plan churn — Elasticsearch resource keeps computed semantics for consistency with its `indices` block.

**Alternative considered:** Also add to `indices` on Kibana for symmetry — deferred; user request targets `remote_indices` only.

### 4. Tests

Extend existing `remote_indices_create` / `remote_indices_update` acceptance configs and checks for both resources. Add or extend unit tests (`flatten_test.go` for Kibana; model round-trip for Elasticsearch if present).

## Risks / Trade-offs

- **[Risk] Enabling restricted index access is dangerous** → Reuse existing warning description from `indices`; no behavioral change beyond exposing an API field users can already set via API/console.
- **[Risk] Kibana vs Elasticsearch semantic mismatch** (`computed` vs optional) → Documented in design; Kibana follows its existing remote_indices optional-field pattern.
- **[Risk] Drift for roles created outside Terraform with `allow_restricted_indices: true`** → Read path populates computed/optional value from API; next apply may show diff if user omits the field — same as `indices` on Elasticsearch resource.

## Migration Plan

No migration required. Existing configurations without `remote_indices.allow_restricted_indices` continue to work; Terraform omits the field on write when unset and reads API defaults on refresh.

## Open Questions

None — API support is confirmed in both Elasticsearch typed client and `kibanaoapi` structs.
