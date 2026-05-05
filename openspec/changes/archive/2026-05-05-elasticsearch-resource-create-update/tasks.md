## 1. Entitycore Envelope

- [x] 1.1 Extend `ElasticsearchResourceModel` with a plan-safe resource identity accessor used by Create and Update.
- [x] 1.2 Add required create and update callback types to `NewElasticsearchResource`.
- [x] 1.3 Implement `Create` and `Update` on `ElasticsearchResource[T]` with plan decode, resource ID derivation, scoped Elasticsearch client resolution, callback invocation, diagnostics handling, and returned-model state persistence.
- [x] 1.4 Update entitycore package documentation to describe the complete Elasticsearch resource envelope and callback contract.

## 2. Resource Migrations

- [x] 2.1 Migrate `security_role` to pass create and update callbacks to `NewElasticsearchResource` and remove thin Create/Update wrapper methods.
- [x] 2.2 Migrate `security_role_mapping` to pass create and update callbacks to `NewElasticsearchResource` and remove thin Create/Update wrapper methods.
- [x] 2.3 Migrate `security_system_user` to pass create and update callbacks to `NewElasticsearchResource` and remove thin Create/Update wrapper methods.
- [x] 2.4 Migrate `cluster_script` to the Elasticsearch resource envelope and pass create and update callbacks to remove its thin Create/Update wrapper methods.
- [x] 2.5 Update any remaining `NewElasticsearchResource` call sites and model getter methods required by the new constructor and model constraint.

## 3. Tests and Verification

- [x] 3.1 Update entitycore resource envelope unit tests for Create and Update success paths, callback diagnostics, client resolution failures, and type/interface assertions.
- [x] 3.2 Add or update focused resource tests where existing coverage can verify migrated callback behavior without acceptance-test infrastructure.
- [x] 3.3 Run targeted Go tests for `internal/entitycore` and migrated resource packages.
- [x] 3.4 Run OpenSpec validation for `elasticsearch-resource-create-update`.

**Verification (recorded for audit)**

- **3.3** — `go test -count=1 ./internal/entitycore/...` and migrated packages excluding external acceptance tests (same module path; `-run` filters only apply where `TestAcc*` lives in `*_test` packages):
  - `go test -count=1 ./internal/elasticsearch/security/role -run 'Test(V0ToV1|Data_satisfies|FromAPIModel)'`
  - `go test -count=1 ./internal/elasticsearch/security/rolemapping -run 'Test(RoleTemplatesToJSON|Data_satisfies)'`
  - `go test -count=1 ./internal/elasticsearch/security/systemuser -run 'TestData_satisfies'`
  - `go test -count=1 ./internal/elasticsearch/cluster/script -run 'TestData_satisfies'`
- **3.4** — `node_modules/.bin/openspec validate elasticsearch-resource-create-update --type change --strict` (OpenSpec CLI from `make setup-openspec` / `npm ci`). Repo-wide spec check: `make check-openspec` (`openspec validate --all`).
