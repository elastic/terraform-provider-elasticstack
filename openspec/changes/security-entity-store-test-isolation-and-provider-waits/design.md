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
- Eliminate transient 500s in entity-link/entity create by adding bounded retry with back-off.
- Make provider Delete robust for real users (wait for uninstall completion before returning).
- Make provider Read tolerate the `installing` state (poll until `started`/`running` or timeout).

**Non-Goals:**
- Changing the entity-type merge semantics of the Kibana API.
- Guaranteeing test isolation across packages that run in the same space simultaneously — a per-test
  unique space would address that but is a separate, larger change.
- Fixing flakiness in tests that are unrelated to entity-store timing.

## Decisions

### 1. Provider Delete: poll until `not_installed`

`deleteEntityStore()` currently calls `UninstallSecurityEntityStore` and returns immediately.
The fix adds a polling loop (using the same `getEntityStoreStatus` helper already present in
`helpers.go`) that retries `GET /api/security/entity_store/status` until `status == "not_installed"`
or a bounded timeout (default 5 minutes, 5-second poll interval) is exceeded.

Why 5 minutes / 5 seconds: the Kibana background tasks (`extract_entity_task`, `risk_score_maintainer`)
need to complete their current run cycle. Observed times on the live stack were under 30 seconds;
5 minutes is a generous upper bound consistent with other provider resource timeouts.

Alternatives considered:
- Return immediately (status quo): rejected — leaves real users with race conditions on
  destroy-then-create.
- Add a fixed sleep: rejected — brittle, slower on fast stacks, and insufficient on slow ones.

### 2. Provider Read: poll until `started`/`running`

`readEntityStore()` calls `getEntityStoreStatus` once and reads whatever engines are present. On
a freshly installed store that is still `installing`, engines may be partially reported.

The fix adds a short polling loop (default 2-minute timeout, 3-second poll interval) inside `Read`
that retries until at least one engine has `status == "started"` or the overall status is `"running"`.
If the store reaches `not_installed` during polling, the existing "remove from state" path applies.

Why these values: the store normally initializes in seconds; 2 minutes allows for slow stacks.

### 3. Provider entity-link/entity Create: retry on HTTP 500

`POST /api/security/entity_store/resolution/link` (and entity-store entity create) return HTTP 500
when the entity store has not yet reached its initialized state. The fix wraps the create call in
a bounded retry loop with exponential back-off (start 2s, max 30s, up to 10 attempts or 2-minute
wall-clock budget), treating HTTP 500 as retryable and all other non-2xx as fatal.

Alternatives considered:
- Pre-check store status before every create: rejected — adds round-trips and a TOCTOU window.
- Fail fast and ask users to retry: rejected — unhelpful UX for a transient initialization race.

### 4. Tests: `t.Cleanup` with full uninstall wait

Each acceptance test that manages the entity store resource SHALL register a `t.Cleanup` function
that calls `POST /api/security/entity_store/uninstall` and polls `GET /api/security/entity_store/status`
until `not_installed` (same timeout/interval as the provider Delete fix). This runs after the test
body and after any Terraform `CheckDestroy` supplied by the framework.

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
  measured); 5-minute ceiling is consistent with standard Terraform default timeouts.
- [Polling adds latency to Read on slow stacks] → Mitigation: polling only activates when status
  is `installing`; a `running`/`started` store returns on the first poll.
- [Retry on 500 may mask genuine store errors] → Mitigation: retry is bounded (10 attempts / 2m)
  and only applies to HTTP 500; non-transient errors manifest after the retry budget is exhausted.
- [Looser test assertions miss real regressions] → Mitigation: the assertion still confirms the
  requested type is present; an exact-set test would be fragile under merge semantics anyway.

## Migration Plan

1. Add `waitForUninstall` helper in `internal/kibana/security_entity_store/helpers.go`.
2. Call it from `delete.go` after the uninstall API call.
3. Add `waitForStarted` helper and call it from `read.go` when status is `installing`.
4. Add retry-on-500 in `internal/kibana/security_entity_store_entity_link/create.go` (and the
   parallel entity create path).
5. Add shared `cleanupEntityStore(t, spaceID)` helper to the test packages and register it in
   each acceptance test.
6. Replace exact `entity_types.#` cardinality assertions with superset-containment checks.

## Open Questions

- None for the initial implementation. Timing constants (poll interval, timeout) should be
  reviewed after observing behavior on CI; they can be tuned as follow-up without a spec change.
