## Why

Security Entity Store acceptance tests are flaky on the 9.5.0-SNAPSHOT matrix due to two related
root causes: (1) cross-test contamination of the per-space singleton entity store when a prior test
leaves types installed (merge-on-install accumulates them for the next test), and (2) transient
HTTP 500s when tests proceed before the store has reached the `started`/initialized state.

Investigation confirmed that Kibana does **not** auto-expand entity types — `install` merges,
which is the same behavior on 9.4. The 9.5-specific factor is slower async init/cleanup timing,
making these latent design gaps visible.

Both gaps affect real users on 9.5, not just tests: a user running `terraform destroy` followed
immediately by `terraform apply` may observe stale state if uninstall has not completed; a user
reading or creating entity-link resources while the store is still initializing will see spurious
500 errors.

## What Changes

- **Provider — Delete** waits for the entity store to reach `not_installed` (polling
  `GET /api/security/entity_store/status` via `asyncutils.WaitForStateTransition`) before
  returning, so Terraform state is only removed when the API agrees the store is gone. The wait is
  bounded by the resource's `timeouts` block (default 20m), not a hardcoded value.
- **Provider — Read** waits for the store to leave `installing` (polling status via
  `asyncutils.WaitForStateTransition`, bounded by the `timeouts` block, default 5m) before reading
  entity types, preventing the plan from seeing a transient half-initialized state. On deadline
  expiry it degrades to a warning + partial read rather than failing.
- **Provider — entity-link/entity Create** retries on HTTP 500 via
  `asyncutils.WaitForStateTransition` (bounded by the Create `timeouts`), so transient
  store-initialization errors do not fail the apply. Non-500 responses fail fast.
- **Tests** add a `t.Cleanup` function in every acceptance test that uninstalls the entity store
  and waits for `not_installed` before the next test can proceed, breaking cross-test contamination
  of the singleton.
- **Tests** loosen `entity_types` assertions from exact-set equality to superset containment
  ("contains the requested types"), tolerating accumulated types left by the Kibana merge-on-install
  semantic during parallel or sequential test runs.

## Capabilities

### Modified Capabilities

- `kibana-security-entity-store`: Delete waits for uninstall completion; Read waits for started-state.
- `kibana-security-entity-store-entity-link`: Create retries on HTTP 500 within the Create timeout.
- `kibana-security-entity-store-entities-datasource`: (test isolation fix)

### New Capabilities
<!-- None. -->

## Impact

- Changes to provider Go files under `internal/kibana/security_entity_store/` (delete, read)
  and `internal/kibana/security_entity_store_entity_link/` (create).
- Changes to acceptance test files (`acc_test.go`) in those packages and related entity/entity-link
  packages to add cleanup helpers and loosen assertions.
- No schema changes; no new resource types; no documentation changes beyond the spec delta.
