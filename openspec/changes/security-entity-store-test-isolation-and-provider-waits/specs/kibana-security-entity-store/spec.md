## MODIFIED Requirements

### Requirement: Delete waits for uninstall completion (REQ-WAIT-001)

After calling `POST /api/security/entity_store/uninstall`, the provider SHALL poll
`GET /api/security/entity_store/status` until the response contains `status == "not_installed"` or
a configurable timeout (default 5 minutes, 5-second poll interval) is exceeded.

While polling, transient network errors on the status request SHALL be retried. A timeout expiry
SHALL produce a clear error diagnostic and SHALL NOT silently remove the resource from state.

#### Scenario: Delete waits for not_installed before returning

- GIVEN `terraform destroy` runs on an installed entity store
- WHEN `POST /api/security/entity_store/uninstall` returns 200
- AND the store status is `"uninstalling"` for the first two polls
- THEN the provider SHALL continue polling
- AND SHALL remove the resource from state only after status transitions to `"not_installed"`

#### Scenario: Delete times out if uninstall does not complete

- GIVEN `terraform destroy` runs on an installed entity store
- WHEN the store status never reaches `"not_installed"` within the 5-minute timeout
- THEN the provider SHALL return an error diagnostic describing the timeout
- AND SHALL NOT remove the resource from state

### Requirement: Read waits for the store to leave installing state before reading engines (REQ-WAIT-002)

The provider SHALL poll `GET /api/security/entity_store/status` when the overall status is
`"installing"`, retrying every 3 seconds up to a 2-minute timeout until the status is `"running"`,
`"stopped"`, `"error"`, or `"not_installed"`.

When the status transitions to `"not_installed"` during polling, the provider SHALL apply the
existing "remove from state" path (REQ-001 Scenario: Read removes resource when not installed).

When the timeout is exceeded while the status is still `"installing"`, the provider SHALL emit a
warning diagnostic and proceed with the partial data available (degraded-read behavior, not a
hard error), to avoid breaking `terraform refresh` on a slow stack.

#### Scenario: Read polls while store is installing

- GIVEN a resource in state and `GET /api/security/entity_store/status` returns `"installing"`
- WHEN the provider reads the resource
- THEN the provider SHALL retry the status call until status is no longer `"installing"` or until the 2-minute timeout
- AND SHALL NOT return a partial `entity_types` list unless the timeout is exceeded

#### Scenario: Read emits warning after timeout and returns partial data

- GIVEN a resource in state and the store remains in `"installing"` beyond the 2-minute timeout
- WHEN the provider reads the resource
- THEN the provider SHALL emit a warning diagnostic
- AND SHALL continue reading with whatever engine data is available
- AND SHALL NOT fail the read with a hard error

## ADDED Requirements

### Requirement: Acceptance tests enforce isolation between entity store tests (REQ-TEST-ISOLATION-001)

Every acceptance test that manages the entity store resource SHALL register a `t.Cleanup`
function that uninstalls the entity store and waits for `not_installed` (same timeout/interval
as the provider Delete wait) before the test run concludes. This cleanup MUST be idempotent:
calling it when the store is already `not_installed` SHALL succeed without error.

#### Scenario: Cleanup runs after test body regardless of test outcome

- GIVEN an acceptance test registers `t.Cleanup(cleanupEntityStore)`
- WHEN the test body panics or fails
- THEN `cleanupEntityStore` SHALL still run and leave the space in `not_installed` state

#### Scenario: Cleanup on already-uninstalled store is a no-op

- GIVEN the entity store is already in `not_installed` state when the cleanup runs
- WHEN `cleanupEntityStore` is called
- THEN the function SHALL return without error

### Requirement: Acceptance tests use superset assertions for entity_types (REQ-TEST-SUPERSET-001)

Acceptance tests SHALL NOT assert exact cardinality of `entity_types` when the test installs a
strict subset of the possible entity types. Assertions SHALL verify that each expected type is
**present** in the returned set (superset containment), not that the set is exactly equal to the
installed subset. This tolerates the Kibana `install` merge-on-install behavior without masking
actual type-presence regressions.

#### Scenario: Superset assertion passes when extra types are present

- GIVEN a test installs `entity_types = ["host"]`
- AND a prior test left `"generic"` installed in the singleton store
- WHEN the provider reads and returns `entity_types = ["generic", "host"]`
- THEN a `TestCheckTypeSetElemAttr` assertion on `"host"` SHALL pass
- AND an exact-count assertion on `entity_types.# == 1` SHALL be absent from the test
