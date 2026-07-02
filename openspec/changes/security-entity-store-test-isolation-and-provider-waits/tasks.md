## 1. Provider: wait for uninstall completion in Delete

- [ ] 1.1 Add `waitForUninstall(ctx, client, spaceID string, timeout, interval time.Duration) diag.Diagnostics`
      to `internal/kibana/security_entity_store/helpers.go`. Poll `getEntityStoreStatus` until
      `status == "not_installed"` or the timeout (default 5 minutes, 5-second interval) is exceeded.
      Return a clear error diagnostic if the timeout is reached.
- [ ] 1.2 Call `waitForUninstall` from `internal/kibana/security_entity_store/delete.go` immediately
      after `kibanaoapi.UninstallSecurityEntityStore` succeeds, before returning.
- [ ] 1.3 Add a unit test in `internal/kibana/security_entity_store/` covering the timeout path
      (mock client returns non-`not_installed` status until timeout) and the happy path (status
      transitions to `not_installed` on first or second poll).

## 2. Provider: wait for started-state in Read

- [ ] 2.1 Add `waitForStarted(ctx, client, spaceID string, timeout, interval time.Duration) (*entityStoreStatus, []byte, diag.Diagnostics)`
      to `internal/kibana/security_entity_store/helpers.go`. Poll `getEntityStoreStatus` until
      the overall status is no longer `"installing"` (e.g. `"running"`, `"stopped"`, `"error"`) or
      `"not_installed"` (no engines), or the timeout (default 2 minutes, 3-second interval) is exceeded.
- [ ] 2.2 Replace the single `getEntityStoreStatus` call in `internal/kibana/security_entity_store/read.go`
      with a `waitForStarted` call, so that `Read` does not return a partial engine list while
      the store is still `installing`.
- [ ] 2.3 Add a unit test covering the `installing`→`running` transition path and the
      `not_installed` early-exit path.

## 3. Provider: retry HTTP 500 in entity-link and entity Create

- [ ] 3.1 Wrap the `POST /api/security/entity_store/resolution/link` call in
      `internal/kibana/security_entity_store_entity_link/create.go` with a bounded retry loop:
      start 2s, double each attempt, cap at 30s, max 10 attempts or 2-minute wall-clock budget.
      Treat HTTP 500 as retryable; treat all other non-2xx as fatal (no retry).
- [ ] 3.2 Apply the same retry pattern to the entity Create call in
      `internal/kibana/security_entity_store/entity/write.go` (the `POST` that creates an
      entity-store entity, which also hits 500 during store initialization).
- [ ] 3.3 Extract the retry helper to a shared internal utility (e.g.
      `internal/kibana/security_entity_store/retryutil.go` or a small `internal/retryutil` package) so both
      callsites use the same implementation.
- [ ] 3.4 Add unit tests for the retry logic: verify that 500 triggers retry, non-500 errors
      do not retry, and that the budget/attempt ceiling causes a final error.

## 4. Tests: add `t.Cleanup` with full uninstall wait

- [ ] 4.1 Add a shared `cleanupEntityStore(t *testing.T, spaceID string)` function in
      `internal/acctest` (non-`_test.go`, so it can be reused across packages). The function MUST:
      - Call `POST /api/security/entity_store/uninstall` (all types) via the Kibana API client
        accessible from the test environment.
      - Poll `GET /api/security/entity_store/status` until `status == "not_installed"` or a
        timeout of 5 minutes (matching the provider Delete timeout).
      - Call `t.Log` with progress messages so CI logs show what the cleanup is doing.
- [ ] 4.2 Register `t.Cleanup(func() { cleanupEntityStore(t, "default") })` at the top of every
      acceptance test in `internal/kibana/security_entity_store/acc_test.go`, including:
      - `TestAccResourceKibanaSecurityEntityStore_basic`
      - `TestAccResourceKibanaSecurityEntityStore_singleType`
      - `TestAccResourceKibanaSecurityEntityStore_import`
      - `TestAccResourceKibanaSecurityEntityStore_startedFalse`
      - `TestAccResourceKibanaSecurityEntityStore_historySnapshot`
      - `TestAccDataSourceKibanaSecurityEntityStoreStatus_basic`
- [ ] 4.3 Register the same cleanup in acceptance tests in related packages:
      - `internal/kibana/security_entity_store/entity/acc_test.go` — all `TestAccResourceKibanaSecurityEntityStoreEntity_*` tests
      - `internal/kibana/security_entity_store_entity_link/acc_test.go` — `TestAccResourceSecurityEntityStoreEntityLink`, `TestAccResourceSecurityEntityStoreEntityLink_SingleElement`
      - `internal/kibana/security_entity_store_resolution_group/acc_test.go` — `TestAccDataSourceSecurityEntityStoreResolutionGroup`
- [ ] 4.4 Ensure the cleanup helper is idempotent (a `not_installed` store does not cause an error).

## 5. Tests: loosen `entity_types` assertions to superset containment

- [ ] 5.1 In `internal/kibana/security_entity_store/acc_test.go`:
      - Replace `TestCheckResourceAttr("…", "entity_types.#", "1")` (cardinality exact match)
        with `TestCheckTypeSetElemAttr` confirming the expected single type is present.
      - Specifically: `TestAccResourceKibanaSecurityEntityStore_singleType` asserts
        `entity_types.# == 1` — loosen this to "contains `host`" without enforcing exact count.
- [ ] 5.2 Apply the same loosening in entity-link and entity-store-entity tests where
      `entity_types` cardinality is asserted against the plan value.

## 6. Validate and integrate

- [ ] 6.1 Run `make build` to confirm no compilation errors are introduced.
- [ ] 6.2 Run the entity-store acceptance test suite (without `TF_ACC` matrix requirements but
      including a sanity build) using `go test -run 'TestAccResourceKibanaSecurityEntityStore|TestAccDataSourceKibana|TestAccResourceSecurityEntityStoreEntityLink|TestAccDataSourceSecurityEntityStoreResolutionGroup'`
      against a running 9.5 stack to confirm the flakiness is resolved.
- [ ] 6.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate security-entity-store-test-isolation-and-provider-waits --type change` and fix any reported problems.
