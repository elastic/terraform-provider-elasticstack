## 1. Entitycore Envelope

- [ ] 1.1 Extend `ElasticsearchResourceModel` with a plan-safe resource identity accessor used by Create and Update.
- [ ] 1.2 Add required create and update callback types to `NewElasticsearchResource`.
- [ ] 1.3 Implement `Create` and `Update` on `ElasticsearchResource[T]` with plan decode, resource ID derivation, scoped Elasticsearch client resolution, callback invocation, diagnostics handling, and returned-model state persistence.
- [ ] 1.4 Update entitycore package documentation to describe the complete Elasticsearch resource envelope and callback contract.

## 2. Resource Migrations

- [ ] 2.1 Migrate `security_role` to pass create and update callbacks to `NewElasticsearchResource` and remove thin Create/Update wrapper methods.
- [ ] 2.2 Migrate `security_role_mapping` to pass create and update callbacks to `NewElasticsearchResource` and remove thin Create/Update wrapper methods.
- [ ] 2.3 Migrate `security_system_user` to pass create and update callbacks to `NewElasticsearchResource` and remove thin Create/Update wrapper methods.
- [ ] 2.4 Migrate `cluster_script` to the Elasticsearch resource envelope and pass create and update callbacks to remove its thin Create/Update wrapper methods.
- [ ] 2.5 Update any remaining `NewElasticsearchResource` call sites and model getter methods required by the new constructor and model constraint.

## 3. Tests and Verification

- [ ] 3.1 Update entitycore resource envelope unit tests for Create and Update success paths, callback diagnostics, client resolution failures, and type/interface assertions.
- [ ] 3.2 Add or update focused resource tests where existing coverage can verify migrated callback behavior without acceptance-test infrastructure.
- [ ] 3.3 Run targeted Go tests for `internal/entitycore` and migrated resource packages.
- [ ] 3.4 Run OpenSpec validation for `elasticsearch-resource-create-update`.
