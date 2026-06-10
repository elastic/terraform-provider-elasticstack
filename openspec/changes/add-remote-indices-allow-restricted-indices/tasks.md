## 1. Elasticsearch security role

- [x] 1.1 Add `allow_restricted_indices` to `remote_indices` in `internal/elasticsearch/security/role/schema.go` (resource) and `data_source.go` (data source), reusing `allow_restricted_indices.md` description and matching `indices` plan modifiers on the resource
- [x] 1.2 Add `AllowRestrictedIndices` to `RemoteIndexPermsData`; wire `toAPIModel` and `fromAPIModel` in `models.go` to read/write `estypes.RemoteIndicesPrivileges.AllowRestrictedIndices`
- [x] 1.3 Map `allow_restricted_indices` in data source read (`data_source.go` remote indices flattener and `getRemoteIndexPermsDSAttrTypes`)
- [x] 1.4 Extend `remote_indices_create` / `remote_indices_update` test configs with `allow_restricted_indices`; add `TestCheckResourceAttr` assertions in `acc_test.go` and `data_source_test.go`

## 2. Kibana security role

- [x] 2.1 Add `attrAllowRestrictedIndices` constant and description; add attribute to `remoteIndicesResourceBlock()` and data source schema in `internal/kibana/security_role/`
- [x] 2.2 Update `esRemoteIndexResourceAttrTypes()` and data-source attr types for `allow_restricted_indices`
- [x] 2.3 Extend `expandedEntry`, `expandEntryCommon`, and `expandRemoteEntry` to map config → `kibanaoapi.SecurityRoleESRemoteIndex.AllowRestrictedIndices`
- [x] 2.4 Map API → state in `flattenRemoteIndicesResource` (and data source flattener if separate)
- [x] 2.5 Extend `remote_indices_create` / `remote_indices_update` acceptance tests and `flatten_test.go` round-trip coverage

## 3. Documentation and validation

- [x] 3.1 Run `make generate-docs` (or project doc generation target) and verify `docs/resources/elasticsearch_security_role.md`, `docs/data-sources/elasticsearch_security_role.md`, `docs/resources/kibana_security_role.md`, and `docs/data-sources/kibana_security_role.md` list `allow_restricted_indices` under `remote_indices`
- [x] 3.2 Run `make build` and targeted acceptance tests for both security role resources
- [x] 3.3 Run `make check-openspec` after syncing or before archive as appropriate
