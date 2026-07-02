## Context

The Security Entity Store is a **singleton per Kibana space**: `POST /api/security/entity_store/install`
merges entity types with whatever is already present rather than replacing them. Uninstall is
asynchronous — the API call returns immediately but the store background tasks (`extract_entity_task`,
`risk_score_maintainer`) take non-deterministic time to wind down. On 9.5.0-SNAPSHOT this timing
is materially slower than 9.4, exposing two latent gaps:

1. **Cross-test contamination**: test A installs `["generic"]`, tears down without waiting for
   completion, test B installs `["host"]` — the singleton now reports `["generic","host"]` while
   test B's plan expected exactly `["host"]`. The `entity_types: length changed from 1 to 4` error
   is this accumulation.

2. **Async init races**: entity-link and entity resources call the API while the store is still
   in `installing` status, receiving HTTP 500 from background task initialization.

Both gaps were verified against a live 9.5 stack. The provider's own `Read` is accurate — it
returns what the API returns — so there is no correctness bug in state mapping. The problems are
timing windows and test isolation.

## Goals / Non-Goals

**Goals:**

- Eliminate cross-test contamination by making each acceptance test own its cleanup and wait for
  full `not_installed` state.
- Eliminate transient 500s in entity-link/entity create by retrying via
  `asyncutils.WaitForStateTransition`, bounded by the Create `timeouts` deadline.
- Make provider Delete robust for real users (wait for uninstall completion before returning).
- Make provider Read tolerate the `installing` state (poll until `started`/`running` or timeout).

**Non-Goals:**

- Changing the entity-type merge semantics of the Kibana API.
- Guaranteeing test isolation across packages that run in the same space simultaneously — a per-test
  unique space would address that but is a separate, larger change.
- Fixing flakiness in tests that are unrelated to entity-store timing.

## Decisions

### 0. Timeouts are driven by the schema `timeouts` block and reuse `asyncutils`

All three affected resource models already embed `entitycore.ResourceTimeoutsField`
(`security_entity_store`, `security_entity_store_entity_link`, and `security_entity_store/entity`),
so each resource already exposes a user-configurable `timeouts` block, and the entitycore base
envelope already resolves it and injects the deadline into `ctx` via `context.WithTimeout` **before**
calling the Read/Create/Update/Delete callbacks (`base_envelope.go` for Read/Delete,
`resource_envelope.go` for Create/Update; defaults from `resource_timeouts.go`:
`DefaultResourceReadTimeout = 5m`, `DefaultResourceDeleteTimeout = 20m`, etc.).

Therefore the wait/retry loops MUST NOT introduce their own hardcoded wall-clock budgets. The
callback `ctx` already carries the correct deadline; the loops bound themselves on `ctx.Done()`.
Only the *poll interval* is a local constant.

All waiting reuses the existing `asyncutils.WaitForStateTransition(ctx, resourceType, resourceID,
StateChecker, ...Option)` primitive (`internal/asyncutils/state_waiter.go`) — the same helper used
by `fleet/integration/create.go` (`waitForFleetIntegrationInstalled`), `ml_anomaly_job`,
`customintegration`, `agentpolicy`, and `connector`. It polls a `StateChecker` until it returns
`true`, returns `ctx.Err()` on deadline, and accepts `WithPollInterval` for cadence. No new
`retryutil` package is introduced.

### 1. Provider Delete: wait until `not_installed`

`deleteEntityStore()` currently calls `UninstallSecurityEntityStore` and returns immediately.
The fix wraps the wait in `asyncutils.WaitForStateTransition` with a `StateChecker` that calls the
existing `getEntityStoreStatus` helper and returns `true` once `status == "not_installed"`. The
bound comes from the Delete `ctx` deadline (from the `timeouts` block, default 20 minutes);
`WithPollInterval(5 * time.Second)` sets cadence. A `ctx` deadline expiry produces a clear error
diagnostic and does NOT silently remove the resource from state.

Alternatives considered:

- Return immediately (status quo): rejected — leaves real users with race conditions on
  destroy-then-create.
- Add a fixed sleep: rejected — brittle, slower on fast stacks, and insufficient on slow ones.
- A bespoke polling loop with a hardcoded 5-minute budget: rejected — duplicates `asyncutils` and
  ignores the user-configurable `timeouts` block already wired into the resource.

### 2. Provider Read: wait until the store leaves `installing`

`readEntityStore()` calls `getEntityStoreStatus` once and reads whatever engines are present. On
a freshly installed store that is still `installing`, engines may be partially reported.

The fix wraps a `StateChecker` in `asyncutils.WaitForStateTransition` (bound by the Read `ctx`
deadline from the `timeouts` block, default 5 minutes; `WithPollInterval(3 * time.Second)`) that
returns `true` once the overall status is no longer `"installing"` (i.e. `"running"`, `"stopped"`,
`"error"`) or the store is `"not_installed"`. If the store reaches `not_installed` during the wait,
the existing "remove from state" path applies.

Because `WaitForStateTransition` fires its first check after one poll interval, `Read` performs an
initial synchronous `getEntityStoreStatus`; it only enters the wait loop when that first read
reports `"installing"`. On `ctx` deadline expiry (`ctx.Err()`), Read downgrades to a **warning**
diagnostic and proceeds with whatever engine data is available (degraded read), rather than
failing `terraform refresh` on a slow stack.

### 3. Provider entity-link/entity Create: retry on HTTP 500

`POST /api/security/entity_store/resolution/link` (and entity-store entity create) return HTTP 500
when the entity store has not yet reached its initialized state. The fix reuses
`asyncutils.WaitForStateTransition` (bound by the Create `ctx` deadline from the `timeouts` block;
`WithPollInterval`) with a `StateChecker` that performs the create call and maps the result:
HTTP 2xx → `(true, nil)` (done); HTTP 500 → `(false, nil)` (retry on next tick); any other non-2xx
→ `(false, err)` (fail fast). The retry is therefore bounded by the user-configurable Create
timeout, not a separate hardcoded budget.

Alternatives considered:

- Pre-check store status before every create: rejected — adds round-trips and a TOCTOU window.
- Fail fast and ask users to retry: rejected — unhelpful UX for a transient initialization race.
- A new `retryutil` package with exponential back-off: rejected — `asyncutils.WaitForStateTransition`
  is the established primitive for this repo; if back-off is later required it should be added as a
  `WithBackoff` option to `asyncutils` rather than forking a parallel utility.

### 4. Tests: `t.Cleanup` with full uninstall wait

Each acceptance test that manages the entity store resource SHALL register a `t.Cleanup` function
that calls `POST /api/security/entity_store/uninstall` and polls `GET /api/security/entity_store/status`
until `not_installed`. Because the cleanup runs in test code (no resource `ctx`/`timeouts` block), it
bounds itself with a `context.WithTimeout` (default 5 minutes, 5-second interval) passed to
`asyncutils.WaitForStateTransition`, reusing the same waiting primitive as the provider. This runs
after the test body and after any Terraform `CheckDestroy` supplied by the framework.

The cleanup helper is shared (extracted to a test-internal package or a `_test.go` file within
each package) rather than duplicated in every test.

Why `t.Cleanup` rather than `CheckDestroy`: `CheckDestroy` runs at the end of the test case's
last step but not between steps in the same test, and does not run if the test panics. `t.Cleanup`
runs unconditionally and is the correct Go testing primitive for this use-case.

### 5. Tests: loosen `entity_types` assertions to superset containment

Tests that assert `resource.TestCheckResourceAttr("…", "entity_types.#", "1")` are asserting
exact cardinality. Because merge-on-install is documented Kibana behavior (and unchanged from 9.4),
the assertion should verify that the requested type is **present** rather than that it is the only
type.

Replace exact-count checks with:

- `resource.TestCheckTypeSetElemAttr` confirming each expected type is present.
- Remove or replace `resource.TestCheckResourceAttr("…entity_types.#", "N")` where N is the
  number the test originally installed (rather than what the singleton may contain after merge).

## Risks / Trade-offs

- [Polling adds latency to happy-path Delete] → Mitigation: already fast on a clean stack (< 10s
  measured); the ceiling is the user-configurable Delete `timeouts` value (default 20 minutes),
  consistent with all other entitycore resources.
- [Polling adds latency to Read on slow stacks] → Mitigation: polling only activates when status
  is `installing`; a `running`/`started` store returns on the first synchronous read.
- [Retry on 500 may mask genuine store errors] → Mitigation: retry is bounded by the Create
  `timeouts` deadline and only applies to HTTP 500; non-transient errors (any other non-2xx) fail
  fast without retrying.
- [Looser test assertions miss real regressions] → Mitigation: the assertion still confirms the
  requested type is present; an exact-set test would be fragile under merge semantics anyway.

## Migration Plan

1. Add a `waitForUninstall` helper in `internal/kibana/security_entity_store/helpers.go` that calls
   `asyncutils.WaitForStateTransition` (bounded by the Delete `ctx`, `WithPollInterval(5s)`) with a
   `StateChecker` returning `true` on `not_installed`.
2. Call it from `delete.go` after the uninstall API call.
3. Add a `waitForStarted` helper (same pattern, bounded by the Read `ctx`, `WithPollInterval(3s)`)
   and call it from `read.go` only when the initial synchronous status read reports `installing`;
   downgrade `ctx` deadline expiry to a warning + degraded read.
4. Wrap the create call in `internal/kibana/security_entity_store_entity_link/create.go` (and the
   parallel entity create path) in `asyncutils.WaitForStateTransition`, mapping HTTP 500 to retry
   and other non-2xx to fail-fast, bounded by the Create `ctx`.
5. Add a shared `cleanupEntityStore(t, spaceID)` helper (bounding its own `context.WithTimeout`)
   and register it in each acceptance test.
6. Replace exact `entity_types.#` cardinality assertions with superset-containment checks.

## Open Questions

- None for the initial implementation. Poll intervals (5s Delete, 3s Read) are local constants and
  can be tuned as follow-up without a spec change; the wall-clock bounds are governed by each
  resource's `timeouts` block.
