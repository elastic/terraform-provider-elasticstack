## 1. Implementation

- [x] 1.1 In `internal/elasticsearch/security/apikey/resource/state_upgrade.go`,
  refactor the **v0 → v1** upgrader to use the raw-state pattern:

  ```go
  0: {
      StateUpgrader: func(ctx context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
          stateMap := stateutil.UnmarshalStateMap(req, resp)
          if resp.Diagnostics.HasError() {
              return
          }
          stateutil.NullifyEmptyString(stateMap, "expiration", "metadata", "role_descriptors")
          stateutil.MarshalStateMap(stateMap, resp)
      },
  },
  ```

  Remove `PriorSchema: &schema0` from this upgrader entry (the raw-state pattern
  does not require a prior schema).

- [x] 1.2 In the same file, refactor the **v1 → v2** upgrader to use the raw-state
  pattern:

  ```go
  1: {
      StateUpgrader: func(ctx context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
          stateMap := stateutil.UnmarshalStateMap(req, resp)
          if resp.Diagnostics.HasError() {
              return
          }
          stateutil.NullifyEmptyString(stateMap, "metadata", "role_descriptors")
          if v, ok := stateMap["type"]; !ok || v == nil || v == "" {
              stateMap["type"] = apikey.DefaultAPIKeyType
          }
          stateutil.MarshalStateMap(stateMap, resp)
      },
  },
  ```

  Remove `PriorSchema: &schema1` from this upgrader entry.

- [x] 1.3 Remove the `schemaWithConnection` helper function and the `schema0`/`schema1`
  variables from `UpgradeState` if they are no longer referenced elsewhere in the
  file. Also remove any imports that are no longer needed after this refactor
  (e.g. `providerschema`, `typeutils`, `timeouts`, `basetypes`).

- [x] 1.4 Add `"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"`
  to the import block of `state_upgrade.go` if it is not already present.

## 2. Testing

- [x] 2.1 Create `internal/elasticsearch/security/apikey/resource/state_upgrade_test.go`
  with unit tests for both upgraders. Cover the following cases for each upgrader:

  **v0 → v1:**
  - `metadata` is `""` → upgraded state has `metadata` = `null`.
  - `role_descriptors` is `""` → upgraded state has `role_descriptors` = `null`.
  - Both `metadata` and `role_descriptors` are `""` simultaneously → both are `null`.
  - `metadata` is a valid JSON string (e.g. `"{}"`) → preserved unchanged.
  - `metadata` is `null` or absent → remains `null`.
  - `expiration` is `""` → upgraded state has `expiration` = `null` (existing behavior preserved).
  - `expiration` is a non-empty string → preserved unchanged.

  **v1 → v2:**
  - `metadata` is `""` → upgraded state has `metadata` = `null`.
  - `role_descriptors` is `""` → upgraded state has `role_descriptors` = `null`.
  - Both `metadata` and `role_descriptors` are `""` simultaneously → both are `null`.
  - `metadata` is a valid JSON string → preserved unchanged.
  - `type` is absent or `""` → upgraded state has `type` = `"rest"` (existing behavior preserved).
  - `type` is `"cross_cluster"` → preserved unchanged.

## 3. Spec

- [x] 3.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate api-key-state-upgrade-json-nullify --type change`
  and resolve any reported problems.
- [ ] 3.2 When implementation is complete, sync the delta spec into
  `openspec/specs/elasticsearch-security-api-key/spec.md` or archive the change
  per the project workflow.
