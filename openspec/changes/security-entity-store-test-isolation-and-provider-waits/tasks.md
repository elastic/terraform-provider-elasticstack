## 1. Provider: wait for uninstall completion in Delete

- [ ] 1.1 Add `waitForUninstall(ctx, client, spaceID string) diag.Diagnostics` to
      `internal/kibana/security_entity_store/helpers.go`. Implement it with
      `asyncutils.WaitForStateTransition(ctx, "security entity store", spaceID, checker,
      asyncutils.WithPollInterval(5*time.Second))`, where `checker` calls `getEntityStoreStatus`
      and returns `true` once `status == "not_installed"`. Do NOT pass a timeout parameter: the
      Delete `ctx` already carries the deadline from the resource `timeouts` block (default 20m via
      `entitycore.DefaultResourceDeleteTimeout`). Convert `ctx.Err()` (deadline exceeded) into a
      clear error diagnostic.
- [ ] 1.2 Call `waitForUninstall` from `internal/kibana/security_entity_store/delete.go` immediately
      after `kibanaoapi.UninstallSecurityEntityStore` succeeds, before returning.
- [ ] 1.3 Add a unit test in `internal/kibana/security_entity_store/` covering the deadline path
      (a cancelled/expired `ctx` yields an error diagnostic) and the happy path (status
      transitions to `not_installed` on first or second poll). Test the `StateChecker` closure and
      the diagnostic mapping rather than re-testing `asyncutils.WaitForStateTransition` itself.

## 2. Provider: wait for started-state in Read

- [ ] 2.1 Add `waitForStarted(ctx, client, spaceID string) (*entityStoreStatus, []byte, diag.Diagnostics)`
      to `internal/kibana/security_entity_store/helpers.go`. Perform an initial synchronous
      `getEntityStoreStatus`; if the overall status is not `"installing"`, return immediately.
      Otherwise poll via `asyncutils.WaitForStateTransition(ctx, ..., asyncutils.WithPollInterval(
      3*time.Second))` with a `StateChecker` that returns `true` once the status is no longer
      `"installing"` (e.g. `"running"`, `"stopped"`, `"error"`) or `"not_installed"`. Bound by the
      Read `ctx` deadline (default 5m via `entitycore.DefaultResourceReadTimeout`); do NOT pass a
      timeout parameter.
- [ ] 2.2 Replace the single `getEntityStoreStatus` call in `internal/kibana/security_entity_store/read.go`
      with a `waitForStarted` call, so that `Read` does not return a partial engine list while the
      store is still `installing`. On `ctx` deadline expiry, downgrade to a **warning** diagnostic
      and proceed with the last-observed engine data (degraded read), not a hard error.
- [ ] 2.3 Add a unit test covering the `installing`→`running` transition path, the
      `not_installed` early-exit path, and the deadline-expiry → warning + degraded-read path.

## 3. Provider: retry HTTP 500 in entity-link and entity Create

- [ ] 3.1 Wrap the `POST /api/security/entity_store/resolution/link` call in
      `internal/kibana/security_entity_store_entity_link/create.go` with
      `asyncutils.WaitForStateTransition` (bounded by the Create `ctx` deadline from the resource
      `timeouts` block; `asyncutils.WithPollInterval` for cadence). The `StateChecker` performs the
      create call and maps the result: HTTP 2xx → `(true, nil)`; HTTP 500 → `(false, nil)` (retry);
      any other non-2xx → `(false, err)` (fail fast). Do NOT introduce a separate wall-clock budget.
- [ ] 3.2 Apply the same pattern to the entity Create call in
      `internal/kibana/security_entity_store/entity/write.go` (the `POST` that creates an
      entity-store entity, which also hits 500 during store initialization).
- [ ] 3.3 Reuse the existing `internal/asyncutils` package for both callsites — do NOT create a new
      `retryutil` package. If exponential back-off is genuinely required, add it as a
      `WithBackoff` option to `asyncutils` in a separate change rather than forking a utility.
- [ ] 3.4 Add unit tests for the `StateChecker` closures: verify 500 maps to retry
      (`false, nil`), non-500 non-2xx maps to fail-fast (`false, err`), and 2xx maps to done
      (`true, nil`). Verify a deadline-expired `ctx` surfaces `ctx.Err()` as an error diagnostic.

## 4. Tests: add `t.Cleanup` with full uninstall wait

- [ ] 4.1 Add a shared `cleanupEntityStore(t *testing.T, spaceID string)` function in
      `internal/acctest` (non-`_test.go`, so it can be reused across packages). The function MUST:
      - Call `POST /api/security/entity_store/uninstall` (all types) via the Kibana API client
        accessible from the test environment.
      - Wait for `status == "not_installed"` via `asyncutils.WaitForStateTransition`, bounding it
        with a test-local `context.WithTimeout` of 5 minutes (test code has no resource `ctx`),
        `asyncutils.WithPollInterval(5*time.Second)`.
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
