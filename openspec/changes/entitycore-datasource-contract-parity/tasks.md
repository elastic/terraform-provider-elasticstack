## 1. Envelope contract

- [x] 1.1 Extend `ElasticsearchDataSourceModel` with `GetID()` and `GetResourceID()`; extend `KibanaDataSourceModel` with `GetID()`, `GetResourceID()`, and `GetSpaceID()` in `internal/entitycore/data_source_envelope.go`
- [x] 1.2 Define `ElasticsearchDataSourceOptions[T]{Schema, Read, PostRead}` and `KibanaDataSourceOptions[T]{Schema, Read, PostRead}` with the new read-callback types `(T, bool, diag.Diagnostics)` (Elasticsearch: `resourceID`; Kibana: `resourceID, spaceID`)
- [x] 1.3 Change `NewElasticsearchDataSource[T]`/`NewKibanaDataSource[T]` to accept the options struct and store `readFunc`/`postReadFunc`
- [x] 1.4 Add data source PostRead hook types `func(ctx, *clients.ElasticsearchScopedClient, T) diag.Diagnostics` and `func(ctx, *clients.KibanaScopedClient, T) diag.Diagnostics` (mirroring the resource `PostReadFunc` ordering but omitting the `privateState any` argument, since data sources have no private state)

## 2. Shared identity and read orchestration

- [x] 2.1 Resolve Elasticsearch read identity in `doDataSourceRead` via `resolveElasticsearchReadResourceID`, erroring with "Invalid resource identifier" when empty
- [x] 2.2 Resolve Kibana read identity (`resourceID`, `spaceID`) via `resolveKibanaResourceIdentity`, honoring the `KibanaUnscopedSpace` opt-out
- [x] 2.3 Invoke the concrete read function with the resolved identity and capture `(T, found, diags)`
- [x] 2.4 Implement the centralized not-found policy: on `found == false`, append a standardized not-found error diagnostic (component, name, identity) and skip state set
- [x] 2.5 Keep composite `id` assignment in each concrete read function (the envelope does not mutate `id`); standard entities set `id` via `client.ID(...)`, non-standard entities (`cluster/info` cluster UUID, `index/indices` target pattern) set their own `id` before returning `found == true`
- [x] 2.6 Invoke `PostRead` (when non-nil) after state is set on a found read

- [x] 2.8 For Kibana data sources, an explicitly configured `space_id` takes precedence over the space embedded in a composite id
- [x] 2.7 Parse a Kibana model's `GetResourceID()` as a composite `<space>/<resource>` key only when the model opts in via `KibanaCompositeResourceID`; Agent Builder agent and skill data sources opt in

## 3. Envelope tests

- [x] 3.1 Update `internal/entitycore/data_source_envelope_test.go` for the new constructor/options and read signatures
- [x] 3.2 Add tests for identity resolution (Elasticsearch and Kibana, including composite ids and unscoped space)
- [x] 3.3 Add tests for the centralized not-found policy and composite `id` assignment
- [x] 3.4 Add tests for the optional `PostRead` hook (invoked on found, skipped on not-found/error)

## 4. Migrate Elasticsearch data sources

- [x] 4.1 Migrate `internal/elasticsearch/security/role` (model identity accessors, new read signature, drop manual not-found field-nulling; the read function retains its `id` assignment)
- [x] 4.2 Migrate `internal/elasticsearch/security/user`
- [x] 4.3 Migrate `internal/elasticsearch/security/rolemapping`
- [x] 4.4 Migrate `internal/elasticsearch/cluster/snapshot_repository_data_source.go` and reconcile its previous warning-based not-found behavior
- [x] 4.5 Migrate `internal/elasticsearch/cluster/info`
- [x] 4.6 Migrate `internal/elasticsearch/synonyms`
- [x] 4.7 Migrate `internal/elasticsearch/queryrulesets`
- [x] 4.8 Migrate `internal/elasticsearch/index/template`
- [x] 4.9 Migrate `internal/elasticsearch/index/indices`
- [x] 4.10 Migrate `internal/elasticsearch/enrich`

- [x] 4.11 Migrate `internal/elasticsearch/ml/trainedmodel` (data source added after the original proposal)
- [x] 4.12 Migrate `internal/elasticsearch/connector/data_source` (data source added after the original proposal)

## 5. Migrate Kibana data sources

- [x] 5.1 Migrate `internal/kibana/agentbuilderskill` (drop inline composite/space resolution)
- [x] 5.2 Migrate `internal/kibana/agentbuilderagent`
- [x] 5.3 Migrate `internal/kibana/agentbuildertool`
- [x] 5.4 Migrate `internal/kibana/agentbuilderworkflow`
- [x] 5.5 Migrate `internal/kibana/security_role`
- [x] 5.6 Migrate `internal/kibana/spaces`
- [x] 5.7 Migrate `internal/kibana/connectors`
- [x] 5.8 Migrate `internal/kibana/exportsavedobjects`

- [x] 5.9 Migrate `internal/kibana/security_entity_store` (status; singleton, fixed resource id)
- [x] 5.10 Migrate `internal/kibana/security_entity_store/entities` (list-style; fixed resource id)
- [x] 5.11 Migrate `internal/kibana/security_entity_store_resolution_group`
- [x] 5.12 Migrate `internal/kibana/osquery_pack`
- [x] 5.13 Migrate `internal/kibana/osquery_saved_query`
- [x] 5.14 Migrate `internal/kibana/tag` (list-style; fixed resource id)

## 6. Migrate Fleet data sources

- [x] 6.1 Migrate `internal/fleet/outputds`
- [x] 6.2 Migrate `internal/fleet/integrationds` (retains documented soft semantics: absent package yields `found == true` with null `version`)
- [x] 6.3 Migrate `internal/fleet/enrollmenttokens`

## 7. Reconcile affected entity-specific specs

The standardized not-found-is-an-error policy contradicts data source requirements in existing entity specs that document warning-only or partial-empty-state not-found behavior and/or "set `id` regardless of whether found". Update these so the archived requirements do not conflict with the new envelope contract.

- [x] 7.1 Update `openspec/specs/elasticsearch-snapshot-repository/spec.md` (REQ-DS-002 "warning + empty type-block attributes" and REQ-DS-003 "`id` set regardless of whether the repository was found") to the standardized error-on-not-found policy and read-callback-owned `id`
- [x] 7.2 Audit the remaining migrated data source specs for conflicting not-found/`id` requirements and update them (e.g. `elasticsearch-security-role`, `elasticsearch-security-user`, `elasticsearch-security-role-mapping`, `elasticsearch-info`, `elasticsearch-indices`, `elasticsearch-index-template`, `elasticsearch-synonym-sets`, `elasticsearch-query-rulesets`, `elasticsearch-enrich-policy`, the `elasticsearch-ml-trained-model`, the `kibana-agentbuilder-*-datasource`, `kibana-security-role`, `kibana-spaces`, `kibana-action-connector`, `kibana-export-saved-objects`, `fleet-output`, `fleet-integration`, `fleet-enrollment-tokens` specs)

## 8. Verify

- [x] 8.1 `make build` passes
- [x] 8.2 Data source acceptance tests pass for migrated entities
- [x] 8.3 `openspec validate entitycore-datasource-contract-parity --strict` passes
- [x] 8.4 Update `openspec/specs/entitycore-datasource-envelope/spec.md` (handled at archive time) and confirm no concrete data source retains manual identity/not-found boilerplate
