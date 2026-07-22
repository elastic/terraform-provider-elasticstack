# Design: Fix entity-store acceptance-test isolation

## Root cause

The Security Entity Store is a **singleton per Kibana space**. All acceptance tests across five packages (`security_entity_store`, `security_entity_store/entities`, `security_entity_store/entity`, `security_entity_store_entity_link`, `security_entity_store_resolution_group`) install the store into the hardcoded space `"default"`. The Go test runner parallelises across packages by default (`ACCTEST_PACKAGE_PARALLELISM=10`), and the CI 2-shard split puts three of these five packages in shard 0, where they run concurrently against the same Kibana instance. Concurrent installs and uninstalls against a singleton produce the observed `entity_types inconsistent result` and HTTP 500 errors.

## Chosen approach: per-test randomised space

Each test generates a random 12-character `spaceID` and creates an `elasticstack_kibana_space` resource in its Terraform fixture. The store, entities, links, and data sources all reference `elasticstack_kibana_space.test.space_id`. Terraform's dependency-driven destroy ordering naturally destroys the entity store (and waits for `not_installed`) before destroying the space.

This is the same pattern already used by `internal/kibana/tag`, `internal/kibana/osquery_pack`, and `internal/kibana/synthetics/*` — no new helpers are needed.

## Alternatives considered

**Shared space with serialisation** — adding a mutex or `t.Parallel()` guards so only one test runs at a time. Rejected: this would serialize tests that could otherwise run concurrently, and the lock scope would span across package boundaries which is not supported by the Go test runner natively.

**Skipping cross-package parallelism in CI** — reducing `ACCTEST_PACKAGE_PARALLELISM` or adding shard-level serialisation for these packages. Rejected: fragile against future package additions that might collide in the same shard, and requires CI configuration changes that don't fix the underlying issue.

**Schema split (`entity_types` desired vs `installed_entity_types` computed)** — the research comment's Approach C. Rejected for this issue: breaking schema change requiring a migration guide; out of scope per issue #3952 closure gates.

## Implementation pattern

### Go test side

```go
const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

func TestAccResourceKibanaSecurityEntityStore_basic(t *testing.T) {
    skipIfUnsupported(t)
    spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
    t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

    resource.Test(t, resource.TestCase{
        PreCheck: func() { acctest.PreCheck(t) },
        Steps: []resource.TestStep{
            {
                ProtoV6ProviderFactories: acctest.Providers,
                ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
                ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("...", "space_id", spaceID),
                    // ...
                ),
            },
            {
                ProtoV6ProviderFactories: acctest.Providers,
                ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
                ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
                PlanOnly:                 true,
            },
        },
    })
}
```

Key points:
- The same `spaceID` is passed to **every step** in a test via `ConfigVariables` — this is essential for multi-step tests that reuse a `ConfigDirectory` across steps (the space must not be recreated mid-test).
- `t.Cleanup` uses the generated `spaceID`, not `"default"`.
- `TestCheckResourceAttr(... "space_id", ...)` assertions are updated to use `spaceID` instead of `"default"`.
- Import tests (`ImportStateIdFunc` or `ImportStateVerify`) continue to work because the state ID is derived from the resource, not hardcoded.

### Terraform fixture side

```hcl
variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}

resource "elasticstack_kibana_security_entity_store" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}
```

All store, entity, entity-link, and data-source resources in the same fixture wire `space_id = elasticstack_kibana_space.test.space_id`.

### Tests that only exercise validation/error paths

Tests such as `entityJsonConflict`, `entityIdMismatch`, `mixedPaginationError`, and `entityIdFilterConflict` never reach the API (they fail at plan time). They can receive the `space_id` variable and include the `elasticstack_kibana_space.test` resource for consistency, though the space will never actually be created in practice.

## Affected files

### Package: `internal/kibana/security_entity_store`

**Go test** (`acc_test.go`): all 9 test functions (8 store resource tests + 1 status data-source test).

**TF fixtures** (11 listed fixture files):
- `TestAccResourceKibanaSecurityEntityStore_basic/basic/`
- `TestAccResourceKibanaSecurityEntityStore_singleType/single_type/`
- `TestAccResourceKibanaSecurityEntityStore_updateLogExtraction/update_log_extraction/`
- `TestAccResourceKibanaSecurityEntityStore_import/create/`
- `TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails/create/`, `shrink/`
- `TestAccResourceKibanaSecurityEntityStore_shrinkWithFlag/shrink_with_flag/`
- `TestAccResourceKibanaSecurityEntityStore_startedFalse/started_false/`
- `TestAccResourceKibanaSecurityEntityStore_historySnapshot/history_snapshot/`
- `TestAccDataSourceKibanaSecurityEntityStoreStatus_basic/default/`, `with_components/`

### Package: `internal/kibana/security_entity_store/entities`

**Go test** (`acc_test.go`): all 8 test functions. The `entities` data source does not install the store itself but reads from one — tests that need a live store create one in the fixture; those that only test validation/pagination errors do not.

**TF fixtures** (8 listed fixtures):
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_basic/list/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_filter/filter/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_pageMode/page_mode/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_spaceId/space_id/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_entityId/entity_id/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_filterQuery/filter_query/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_mixedPaginationError/mixed_pagination/`
- `TestAccDataSourceKibanaSecurityEntityStoreEntities_entityIdFilterConflict/entity_id_filter_conflict/`

### Package: `internal/kibana/security_entity_store/entity`

**Go test** (`acc_test.go`): all 11 test functions (7 live-store/apply-path + 4 plan-time validation).

**TF fixtures** (12 listed fixture files):
- `TestAccResourceKibanaSecurityEntityStoreEntity_generic/create_host/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_updateHost/create_host/`, `update_host/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_import/create_host/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_hostType/create_host/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_userType/create_user/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_serviceType/create_service/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonFallback/create_host_json/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonConflict/entity_json_conflict/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_entityIdMismatch/entity_id_mismatch/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonIdMismatch/entity_json_id_mismatch/`
- `TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonConflict/host_json_conflict/`

### Package: `internal/kibana/security_entity_store_entity_link`

**Go test** (`acc_test.go`): 2 test functions (`TestAccResourceSecurityEntityStoreEntityLink`, `TestAccResourceSecurityEntityStoreEntityLink_SingleElement`). Note: `TestAccResourceSecurityEntityStoreEntityLink_Validation` tests are schema-validation-only and can receive the variable for consistency.

**TF fixtures** (4 directories / 5 files):
- `TestAccResourceSecurityEntityStoreEntityLink/create/`, `update/`
- `TestAccResourceSecurityEntityStoreEntityLink_SingleElement/single_element/`
- `TestAccResourceSecurityEntityStoreEntityLink_Validation/self_link/`, `empty/`

### Package: `internal/kibana/security_entity_store_resolution_group`

**Go test** (`acc_test.go`): 1 test function (`TestAccDataSourceSecurityEntityStoreResolutionGroup`).

**TF fixtures** (1 directory):
- `TestAccDataSourceSecurityEntityStoreResolutionGroup/read/`

### CI

**`.github/workflows/provider.yml`**: add `"9.4.2"` to the version matrix list, between `"9.4.0"` and `"9.5.0-SNAPSHOT"`.

## Open questions

- Does the 9.4.2 install API always inject `generic` alongside every typed engine, or only when `entity_types` is omitted from the install body? (From the research comment — answered by the human investigation: not reachable in isolated, single-space runs, so this is not a blocker.)
- Are the HTTP 500 errors for `userType` and `resolutionGroup` reproducible on a clean 9.4.2 stack, or timing/flake related? (From the research comment — answered: caused by concurrent singleton access, eliminated by per-space isolation.)
- Does the same extra-engine behavior appear on 9.4.0? (From the research comment — moot: the issue is the race, not the Kibana version.)
- Does installing `service` on 9.4.2 actually install all four engine types, or does the status endpoint misreport them? (From the research comment — moot: not reproducible in isolation.)
