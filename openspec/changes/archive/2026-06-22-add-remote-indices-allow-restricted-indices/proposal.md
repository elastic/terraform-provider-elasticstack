## Why

Elasticsearch and Kibana security roles support an `allow_restricted_indices` flag on both local and remote index privilege entries in their APIs. The provider already exposes this attribute on `indices` blocks for `elasticstack_elasticsearch_security_role`, but not on `remote_indices` for either security role resource. Users configuring cross-cluster roles cannot express restricted-index access for remote indices through Terraform, causing drift or manual workarounds.

## What Changes

- Add optional `allow_restricted_indices` to the `remote_indices` nested block on `elasticstack_elasticsearch_security_role` (resource and data source), matching the existing `indices` attribute semantics (optional, computed, `UseStateForUnknown` on the resource).
- Add optional `allow_restricted_indices` to `elasticsearch.remote_indices` on `elasticstack_kibana_security_role` (resource and data source), wired through expand/flatten to the Kibana Role Management API.
- Map the attribute in create/update/read paths and round-trip it in acceptance tests for both resources.
- Regenerate provider documentation for the new schema fields.

## Capabilities

### New Capabilities

_None — this extends existing security role capabilities._

### Modified Capabilities

- `elasticsearch-security-role`: Add `remote_indices.allow_restricted_indices` to schema, API mapping, plan preservation, and data source read mapping.
- `kibana-security-role`: Add `elasticsearch.remote_indices.allow_restricted_indices` to schema, API mapping, and read state flattening.

## Impact

- **Code**: `internal/elasticsearch/security/role/` (schema, models, data source), `internal/kibana/security_role/` (schema, expand, flatten, attr types).
- **APIs**: Elasticsearch Put/Get role and Kibana Role Management PUT/GET — field already exists on API types (`RemoteIndicesPrivileges`, `SecurityRoleESRemoteIndex`).
- **Docs**: `docs/resources/elasticsearch_security_role.md`, `docs/data-sources/elasticsearch_security_role.md`, `docs/resources/kibana_security_role.md`, `docs/data-sources/kibana_security_role.md`.
- **Tests**: Acceptance and unit tests for remote indices create/update on both resources.
- **Breaking changes**: None — additive schema only.
