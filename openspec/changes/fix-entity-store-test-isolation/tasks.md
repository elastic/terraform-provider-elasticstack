# Tasks: Fix entity-store acceptance-test isolation

All tasks in this change are test-infrastructure only. No provider Go source, schema, API client, or documentation files should be touched.

---

## Task 1 — Package `security_entity_store`: update Go acc_test.go

**Status:** complete

**File**: `internal/kibana/security_entity_store/acc_test.go`

1. [x] Add import `"github.com/hashicorp/terraform-plugin-testing/config"` and `sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"` (check existing imports; `sdkacctest` may already be imported under a different alias in related packages — use whatever is consistent).
2. [x] Add constant:
   ```go
   const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"
   ```
3. [x] For **each** of the 9 test functions listed below (8 resource + 1 data source), apply the following pattern:
   - At the top of the test (after `skipIfUnsupported(t)`), generate: `spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)`
   - Replace `t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })` with `t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })`
   - Add `ConfigVariables: config.Variables{"space_id": config.StringVariable(spaceID)}` to **every** `TestStep` in the test (both applying and `PlanOnly` steps — the same variable map).
   - Update any `resource.TestCheckResourceAttr(... "space_id", "default")` checks to use `spaceID`.

   **Test functions** (9: 8 resource + 1 data source):
   - `TestAccResourceKibanaSecurityEntityStore_basic` — ConfigDirectory `"basic"` (2 steps)
   - `TestAccResourceKibanaSecurityEntityStore_singleType` — ConfigDirectory `"single_type"` (2 steps)
   - `TestAccResourceKibanaSecurityEntityStore_updateLogExtraction` — ConfigDirectory `"update_log_extraction"` (2 steps)
   - `TestAccResourceKibanaSecurityEntityStore_import` — ConfigDirectory `"create"` (2 steps: apply + ImportState)
   - `TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails` — ConfigDirectories `"create"` and `"shrink"` (2 steps)
   - `TestAccResourceKibanaSecurityEntityStore_shrinkWithFlag` — ConfigDirectory `"shrink_with_flag"` (2 steps)
   - `TestAccResourceKibanaSecurityEntityStore_startedFalse` — ConfigDirectory `"started_false"` (2 steps)
   - `TestAccResourceKibanaSecurityEntityStore_historySnapshot` — ConfigDirectory `"history_snapshot"` (2 steps)
   - `TestAccDataSourceKibanaSecurityEntityStoreStatus_basic` — ConfigDirectories `"default"` and `"with_components"` (2 steps; both assertions `"space_id", "default"` become `"space_id", spaceID`)

---

## Task 2 — Package `security_entity_store`: update TF fixtures

**Status:** complete

Apply to every `.tf` file in the directories below. Each file needs:

```hcl
variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}
```

And `space_id = elasticstack_kibana_space.test.space_id` added to the `elasticstack_kibana_security_entity_store` resource block (and to any data source block that accepts `space_id`).

**Fixtures** (11 listed `main.tf` files):
- `testdata/TestAccResourceKibanaSecurityEntityStore_basic/basic/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_singleType/single_type/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_updateLogExtraction/update_log_extraction/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_import/create/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails/create/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails/shrink/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_shrinkWithFlag/shrink_with_flag/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_startedFalse/started_false/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStore_historySnapshot/history_snapshot/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreStatus_basic/default/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreStatus_basic/with_components/main.tf`

For the `TestAccDataSourceKibanaSecurityEntityStoreStatus_basic` fixtures, the data source `elasticstack_kibana_security_entity_store_status` already accepts `space_id`; set it to `elasticstack_kibana_space.test.space_id` as well.

---

## Task 3 — Package `security_entity_store/entities`: update Go acc_test.go

**Status:** complete

**File**: `internal/kibana/security_entity_store/entities/acc_test.go`

1. Add same imports (`config`, `sdkacctest`) and `accTestKibanaSpaceIDCharset` constant.
2. For each of the 8 test functions:
   - Generate `spaceID` per test.
   - Add `CleanupEntityStore` calls only where the test actually installs the entity store (functions that create store + entity fixtures). Tests that only trigger validation errors (`mixedPaginationError`, `entityIdFilterConflict`) still receive the variable but may not need cleanup.
   - Add `ConfigVariables` with `space_id` to every step.
   - Update any `"space_id", "default"` assertions.

   **Test functions** (8):
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_basic`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_mixedPaginationError`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_entityIdFilterConflict`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_filter`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_pageMode`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_spaceId`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_entityId`
   - `TestAccDataSourceKibanaSecurityEntityStoreEntities_filterQuery`

---

## Task 4 — Package `security_entity_store/entities`: update TF fixtures

**Status:** complete

Add `variable "space_id"`, `elasticstack_kibana_space.test`, and wire `space_id` on all entity-store, entity resource (where present), and data-source blocks.

**Directories** (8 files):
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_basic/list/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_mixedPaginationError/mixed_pagination/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_entityIdFilterConflict/entity_id_filter_conflict/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_filter/filter/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_pageMode/page_mode/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_spaceId/space_id/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_entityId/entity_id/main.tf`
- `testdata/TestAccDataSourceKibanaSecurityEntityStoreEntities_filterQuery/filter_query/main.tf`

For `mixedPaginationError` and `entityIdFilterConflict`: these tests error at plan time. Include the variable and space resource for consistency even though the space won't be created.

---

## Task 5 — Package `security_entity_store/entity`: update Go acc_test.go

**Status:** complete

**File**: `internal/kibana/security_entity_store/entity/acc_test.go`

1. [x] Add same imports and `accTestKibanaSpaceIDCharset` constant.
2. [x] For each of the 11 test functions (7 live-store/apply-path + 4 plan-time validation):
   - Generate `spaceID` per test.
   - Replace `"default"` in `CleanupEntityStore` calls with `spaceID` (apply-path tests only; omit cleanup for plan-time validation tests).
   - Add `ConfigVariables` with `space_id` to every step.

   **Test functions** (11):
   - `TestAccResourceKibanaSecurityEntityStoreEntity_generic`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_updateHost`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_import`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonConflict`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_entityIdMismatch`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonIdMismatch`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_hostType`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_userType`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_serviceType`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonFallback`
   - `TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonConflict`

   Note: `entityJsonConflict`, `entityIdMismatch`, `entityJsonIdMismatch`, `hostJsonConflict` fail at plan time; include `space_id` variable and space resource in fixtures but no store installation occurs.

---

## Task 6 — Package `security_entity_store/entity`: update TF fixtures

**Status:** complete

Add `variable "space_id"`, `elasticstack_kibana_space.test`, and wire `space_id` on all entity-store and entity resource blocks.

**Directories** (11 files):
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_generic/create_host/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_updateHost/create_host/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_updateHost/update_host/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_import/create_host/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_hostType/create_host/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_userType/create_user/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_serviceType/create_service/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonFallback/create_host_json/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonConflict/entity_json_conflict/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_entityIdMismatch/entity_id_mismatch/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonIdMismatch/entity_json_id_mismatch/main.tf`
- `testdata/TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonConflict/host_json_conflict/main.tf`

The entity resource block also accepts `space_id`; wire it on every entity resource that is a sibling of the store in the same fixture.

---

## Task 7 — Package `security_entity_store_entity_link`: update Go acc_test.go

**File**: `internal/kibana/security_entity_store_entity_link/acc_test.go`

1. Add same imports and `accTestKibanaSpaceIDCharset` constant.
2. For each of the 3 test functions:
   - Generate `spaceID` per test.
   - Replace `"default"` in `CleanupEntityStore` calls with `spaceID`.
   - Add `ConfigVariables` with `space_id` to every step.
   - Update `"space_id", "default"` and `"id", "default/..."` assertions to use `spaceID`.

   **Test functions** (3):
   - `TestAccResourceSecurityEntityStoreEntityLink` — 3 steps (create, update, import)
   - `TestAccResourceSecurityEntityStoreEntityLink_SingleElement`
   - `TestAccResourceSecurityEntityStoreEntityLink_Validation` (schema validation; include variable for consistency)

---

## Task 8 — Package `security_entity_store_entity_link`: update TF fixtures

Add `variable "space_id"`, `elasticstack_kibana_space.test`, and wire `space_id` on all store, entity, and entity-link resource blocks.

**Directories** (5 files):
- `testdata/TestAccResourceSecurityEntityStoreEntityLink/create/main.tf`
- `testdata/TestAccResourceSecurityEntityStoreEntityLink/update/main.tf`
- `testdata/TestAccResourceSecurityEntityStoreEntityLink_SingleElement/single_element/main.tf`
- `testdata/TestAccResourceSecurityEntityStoreEntityLink_Validation/self_link/main.tf`
- `testdata/TestAccResourceSecurityEntityStoreEntityLink_Validation/empty/main.tf`

The entity-link resource (`elasticstack_kibana_security_entity_store_entity_link`) accepts `space_id`; wire it on every entity-link block. Note: `Validation` fixtures produce plan-time errors; include the variable and space for consistency.

---

## Task 9 — Package `security_entity_store_resolution_group`: update Go acc_test.go

**File**: `internal/kibana/security_entity_store_resolution_group/acc_test.go`

1. Add same imports and `accTestKibanaSpaceIDCharset` constant.
2. For `TestAccDataSourceSecurityEntityStoreResolutionGroup`:
   - Generate `spaceID`.
   - Replace `"default"` in `CleanupEntityStore` with `spaceID`.
   - Add `ConfigVariables` to the single step.
   - Update `"space_id", "default"` and `"id", "default/..."` assertions to use `spaceID`.

---

## Task 10 — Package `security_entity_store_resolution_group`: update TF fixture

**File**: `testdata/TestAccDataSourceSecurityEntityStoreResolutionGroup/read/main.tf`

Add `variable "space_id"`, `elasticstack_kibana_space.test`, and wire `space_id` on the store, entity, and data source blocks. The data source `elasticstack_kibana_security_entity_store_resolution_group` accepts `space_id`.

---

## Task 11 — CI matrix: add 9.4.2

**File**: `.github/workflows/provider.yml`

In the `strategy.matrix.version` list (currently ending with `"9.4.0"` then `"9.5.0-SNAPSHOT"`), insert `"9.4.2"` between those two entries:

```yaml
          - "9.4.0"
          - "9.4.2"
          - "9.5.0-SNAPSHOT"
```

No other CI file changes are needed.

---

## Verification checklist (after implementation)

- [ ] `go build ./...` succeeds (no compile errors in test files).
- [ ] Each of the five packages runs successfully in isolation against a live 9.4.2 stack.
- [ ] All five packages run concurrently (`go test ./internal/kibana/security_entity_store/... ./internal/kibana/security_entity_store_resolution_group/... ./internal/kibana/security_entity_store_entity_link/...`) against 9.4.2 without the `entity_types inconsistent result` or HTTP 500 errors.
- [ ] `9.4.2` appears in the CI matrix in `provider.yml`.
