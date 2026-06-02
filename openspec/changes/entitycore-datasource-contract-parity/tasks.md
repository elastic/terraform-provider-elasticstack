## 1. Envelope contract

- [ ] 1.1 Extend `ElasticsearchDataSourceModel` with `GetID()` and `GetResourceID()`; extend `KibanaDataSourceModel` with `GetID()`, `GetResourceID()`, and `GetSpaceID()` in `internal/entitycore/data_source_envelope.go`
- [ ] 1.2 Define `ElasticsearchDataSourceOptions[T]{Schema, Read, PostRead}` and `KibanaDataSourceOptions[T]{Schema, Read, PostRead}` with the new read-callback types `(T, bool, diag.Diagnostics)` (Elasticsearch: `resourceID`; Kibana: `resourceID, spaceID`)
- [ ] 1.3 Change `NewElasticsearchDataSource[T]`/`NewKibanaDataSource[T]` to accept the options struct and store `readFunc`/`postReadFunc`
- [ ] 1.4 Add data source PostRead hook types `func(ctx, *clients.ElasticsearchScopedClient, T) diag.Diagnostics` and `func(ctx, *clients.KibanaScopedClient, T) diag.Diagnostics` (mirroring the resource `PostReadFunc` ordering but omitting the `privateState any` argument, since data sources have no private state)

## 2. Shared identity and read orchestration

- [ ] 2.1 Resolve Elasticsearch read identity in `doDataSourceRead` via `resolveElasticsearchReadResourceID`, erroring with "Invalid resource identifier" when empty
- [ ] 2.2 Resolve Kibana read identity (`resourceID`, `spaceID`) via `resolveKibanaResourceIdentity`, honoring the `KibanaUnscopedSpace` opt-out
- [ ] 2.3 Invoke the concrete read function with the resolved identity and capture `(T, found, diags)`
- [ ] 2.4 Implement the centralized not-found policy: on `found == false`, append a standardized not-found error diagnostic (component, name, identity) and skip state set
- [ ] 2.5 Keep composite `id` assignment in each concrete read function (the envelope does not mutate `id`); standard entities set `id` via `client.ID(...)`, non-standard entities (`cluster/info` cluster UUID, `index/indices` target pattern) set their own `id` before returning `found == true`
- [ ] 2.6 Invoke `PostRead` (when non-nil) after state is set on a found read

## 3. Envelope tests

- [ ] 3.1 Update `internal/entitycore/data_source_envelope_test.go` for the new constructor/options and read signatures
- [ ] 3.2 Add tests for identity resolution (Elasticsearch and Kibana, including composite ids and unscoped space)
- [ ] 3.3 Add tests for the centralized not-found policy and composite `id` assignment
- [ ] 3.4 Add tests for the optional `PostRead` hook (invoked on found, skipped on not-found/error)

## 4. Migrate Elasticsearch data sources

- [ ] 4.1 Migrate `internal/elasticsearch/security/role` (model identity accessors, new read signature, drop manual not-found field-nulling; the read function retains its `id` assignment)
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

## 7. Reconcile affected entity-specific specs

The standardized not-found-is-an-error policy contradicts data source requirements in existing entity specs that document warning-only or partial-empty-state not-found behavior and/or "set `id` regardless of whether found". Update these so the archived requirements do not conflict with the new envelope contract.

- [ ] 7.1 Update `openspec/specs/elasticsearch-snapshot-repository/spec.md` (REQ-DS-002 "warning + empty type-block attributes" and REQ-DS-003 "`id` set regardless of whether the repository was found") to the standardized error-on-not-found policy and read-callback-owned `id`
- [ ] 7.2 Audit the remaining migrated data source specs for conflicting not-found/`id` requirements and update them (e.g. `elasticsearch-security-role`, `elasticsearch-security-user`, `elasticsearch-security-role-mapping`, `elasticsearch-info`, `elasticsearch-indices`, `elasticsearch-index-template`, `elasticsearch-synonym-sets`, `elasticsearch-query-rulesets`, `elasticsearch-enrich-policy`, the `kibana-agentbuilder-*-datasource`, `kibana-security-role`, `kibana-spaces`, `kibana-action-connector`, `kibana-export-saved-objects`, `fleet-output`, `fleet-integration`, `fleet-enrollment-tokens` specs)

## 8. Verify

- [ ] 8.1 `make build` passes
- [ ] 8.2 Data source acceptance tests pass for migrated entities
- [ ] 8.3 `openspec validate entitycore-datasource-contract-parity --strict` passes
- [ ] 8.4 Update `openspec/specs/entitycore-datasource-envelope/spec.md` (handled at archive time) and confirm no concrete data source retains manual identity/not-found boilerplate
