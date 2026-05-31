## 1. Envelope contract

- [ ] 1.1 Extend `ElasticsearchDataSourceModel` with `GetID()` and `GetResourceID()`; extend `KibanaDataSourceModel` with `GetID()`, `GetResourceID()`, and `GetSpaceID()` in `internal/entitycore/data_source_envelope.go`
- [ ] 1.2 Define `ElasticsearchDataSourceOptions[T]{Schema, Read, PostRead}` and `KibanaDataSourceOptions[T]{Schema, Read, PostRead}` with the new read-callback types `(T, bool, diag.Diagnostics)` (Elasticsearch: `resourceID`; Kibana: `resourceID, spaceID`)
- [ ] 1.3 Change `NewElasticsearchDataSource[T]`/`NewKibanaDataSource[T]` to accept the options struct and store `readFunc`/`postReadFunc`
- [ ] 1.4 Add a data source `PostReadFunc`-shaped hook type mirroring the resource `PostReadFunc`

## 2. Shared identity and read orchestration

- [ ] 2.1 Resolve Elasticsearch read identity in `doDataSourceRead` via `resolveElasticsearchReadResourceID`, erroring with "Invalid resource identifier" when empty
- [ ] 2.2 Resolve Kibana read identity (`resourceID`, `spaceID`) via `resolveKibanaResourceIdentity`, honoring the `KibanaUnscopedSpace` opt-out
- [ ] 2.3 Invoke the concrete read function with the resolved identity and capture `(T, found, diags)`
- [ ] 2.4 Implement the centralized not-found policy: on `found == false`, append a standardized not-found error diagnostic (component, name, identity) and skip state set
- [ ] 2.5 Assign the composite `id` from the scoped client and resolved identity on a found read before setting state
- [ ] 2.6 Invoke `PostRead` (when non-nil) after state is set on a found read

## 3. Envelope tests

- [ ] 3.1 Update `internal/entitycore/data_source_envelope_test.go` for the new constructor/options and read signatures
- [ ] 3.2 Add tests for identity resolution (Elasticsearch and Kibana, including composite ids and unscoped space)
- [ ] 3.3 Add tests for the centralized not-found policy and composite `id` assignment
- [ ] 3.4 Add tests for the optional `PostRead` hook (invoked on found, skipped on not-found/error)

## 4. Migrate Elasticsearch data sources

- [ ] 4.1 Migrate `internal/elasticsearch/security/role` (model identity accessors, new read signature, drop manual `id`/not-found field-nulling)
- [ ] 4.2 Migrate `internal/elasticsearch/security/user`
- [ ] 4.3 Migrate `internal/elasticsearch/security/rolemapping`
- [ ] 4.4 Migrate `internal/elasticsearch/cluster/snapshot_repository_data_source.go` and reconcile its previous warning-based not-found behavior
- [ ] 4.5 Migrate `internal/elasticsearch/cluster/info`
- [ ] 4.6 Migrate `internal/elasticsearch/synonyms`
- [ ] 4.7 Migrate `internal/elasticsearch/queryrulesets`
- [ ] 4.8 Migrate `internal/elasticsearch/index/template`
- [ ] 4.9 Migrate `internal/elasticsearch/index/indices`
- [ ] 4.10 Migrate `internal/elasticsearch/enrich`

## 5. Migrate Kibana data sources

- [ ] 5.1 Migrate `internal/kibana/agentbuilderskill` (drop inline composite/space resolution)
- [ ] 5.2 Migrate `internal/kibana/agentbuilderagent`
- [ ] 5.3 Migrate `internal/kibana/agentbuildertool`
- [ ] 5.4 Migrate `internal/kibana/agentbuilderworkflow`
- [ ] 5.5 Migrate `internal/kibana/security_role`
- [ ] 5.6 Migrate `internal/kibana/spaces`
- [ ] 5.7 Migrate `internal/kibana/connectors`
- [ ] 5.8 Migrate `internal/kibana/exportsavedobjects`

## 6. Migrate Fleet data sources

- [ ] 6.1 Migrate `internal/fleet/outputds`
- [ ] 6.2 Migrate `internal/fleet/integrationds`
- [ ] 6.3 Migrate `internal/fleet/enrollmenttokens`

## 7. Verify

- [ ] 7.1 `make build` passes
- [ ] 7.2 Data source acceptance tests pass for migrated entities
- [ ] 7.3 `openspec validate entitycore-datasource-contract-parity --strict` passes
- [ ] 7.4 Update `openspec/specs/entitycore-datasource-envelope/spec.md` (handled at archive time) and confirm no concrete data source retains manual identity/`id`/not-found boilerplate
