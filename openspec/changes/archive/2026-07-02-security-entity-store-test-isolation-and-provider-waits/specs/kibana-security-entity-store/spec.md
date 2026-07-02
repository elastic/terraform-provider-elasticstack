## ADDED Requirements

### Requirement: Delete waits for uninstall completion (REQ-WAIT-001)

After calling `POST /api/security/entity_store/uninstall`, the provider SHALL poll
`GET /api/security/entity_store/status` at a fixed interval (5 seconds) until the response contains
`status == "not_installed"` or the Delete operation deadline is exceeded. The deadline SHALL be
derived from the resource's `timeouts` block (Delete), defaulting to the entitycore Delete timeout
when unset, and SHALL NOT be a separate hardcoded value.

While polling, transient network errors on the status request SHALL be retried. A deadline expiry
SHALL produce a clear error diagnostic and SHALL NOT silently remove the resource from state.

#### Scenario: Delete waits for not_installed before returning

- GIVEN `terraform destroy` runs on an installed entity store
- WHEN `POST /api/security/entity_store/uninstall` returns 200
- AND the store status is `"uninstalling"` for the first two polls
- THEN the provider SHALL continue polling
- AND SHALL remove the resource from state only after status transitions to `"not_installed"`

#### Scenario: Delete times out if uninstall does not complete

- GIVEN `terraform destroy` runs on an installed entity store
- WHEN the store status never reaches `"not_installed"` within the Delete timeout
- THEN the provider SHALL return an error diagnostic describing the timeout
- AND SHALL NOT remove the resource from state

### Requirement: Read waits for the store to leave installing state before reading engines (REQ-WAIT-002)

The provider SHALL first read `GET /api/security/entity_store/status` synchronously; when the
overall status is `"installing"` it SHALL poll every 3 seconds until the status is `"running"`,
`"stopped"`, `"error"`, or `"not_installed"`, or the Read operation deadline is exceeded. The
deadline SHALL be derived from the resource's `timeouts` block (Read), defaulting to the
entitycore Read timeout when unset, and SHALL NOT be a separate hardcoded value.

When the status transitions to `"not_installed"` during polling, the provider SHALL apply the
existing "remove from state" path (REQ-001 Scenario: Read removes resource when not installed).

When the deadline is exceeded while the status is still `"installing"`, the provider SHALL emit a
warning diagnostic and proceed with the partial data available (degraded-read behavior, not a
hard error), to avoid breaking `terraform refresh` on a slow stack.

#### Scenario: Read polls while store is installing

- GIVEN a resource in state and `GET /api/security/entity_store/status` returns `"installing"`
- WHEN the provider reads the resource
- THEN the provider SHALL retry the status call until status is no longer `"installing"` or until the Read timeout is exceeded
- AND SHALL NOT return a partial `entity_types` list unless the timeout is exceeded

#### Scenario: Read emits warning after timeout and returns partial data

- GIVEN a resource in state and the store remains in `"installing"` beyond the Read timeout
- WHEN the provider reads the resource
- THEN the provider SHALL emit a warning diagnostic
- AND SHALL continue reading with whatever engine data is available
- AND SHALL NOT fail the read with a hard error

### Requirement: Install retries on HTTP 500 within the configured timeout (REQ-WAIT-003)

The provider SHALL retry `POST /api/security/entity_store/install` (issued during Create, and during Update when new entity types are added) when it returns HTTP 500, treating this as a transient initialization error and retrying at a fixed poll interval (5 seconds) until the install succeeds or the operation deadline is exceeded. The deadline SHALL be derived from the resource's `timeouts` block (Create/Update), defaulting to the entitycore timeout when unset, and SHALL NOT be a separate hardcoded wall-clock budget.

HTTP 500 is the only status that triggers retry. All other non-2xx responses SHALL be treated as
fatal and returned immediately as error diagnostics without retrying. This mirrors the entity-link
Create retry behavior (`kibana-security-entity-store-entity-link` REQ-ESL-RETRY-001) and is
especially important because the test-isolation cleanup uninstalls the singleton store between
tests, so a subsequent install can race the store's background teardown.

#### Scenario: Install succeeds after retrying a transient 500

- GIVEN the entity store was recently uninstalled and background teardown is still in progress
- AND `POST /api/security/entity_store/install` returns HTTP 500 for the first attempts
- AND returns HTTP 200 on a later attempt
- WHEN the provider creates (or updates) the entity store resource
- THEN the provider SHALL succeed once the install returns 2xx

#### Scenario: Install does not retry on non-500 errors

- GIVEN `POST /api/security/entity_store/install` returns HTTP 400
- WHEN the provider creates the entity store resource
- THEN the provider SHALL immediately return an error diagnostic without retrying

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
