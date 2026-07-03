## ADDED Requirements

### Requirement: Create retries on HTTP 500 within the configured timeout (REQ-ESL-RETRY-001)

When `POST /api/security/entity_store/resolution/link` returns HTTP 500, the provider SHALL
treat this as a transient initialization error and retry at a fixed poll interval until the create
succeeds or the Create operation deadline is exceeded. The deadline SHALL be derived from the
resource's `timeouts` block (Create), defaulting to the entitycore Create timeout when unset, and
SHALL NOT be a separate hardcoded wall-clock budget.

HTTP 500 is the only status that triggers retry. All other non-2xx responses (400, 403, 404, 409,
etc.) SHALL be treated as fatal and returned immediately as error diagnostics without retrying.

If the Create deadline is reached with the last attempt still returning HTTP 500, the provider
SHALL return an error diagnostic describing the HTTP 500, indicating that the entity store may
not yet be fully initialized.

#### Scenario: Create succeeds after retrying a transient 500

- GIVEN the entity store is in `"installing"` state
- AND `POST /api/security/entity_store/resolution/link` returns HTTP 500 for the first two attempts
- AND returns HTTP 200 on the third attempt
- WHEN the provider creates the entity link resource
- THEN the provider SHALL succeed and populate final state from the successful attempt

#### Scenario: Create does not retry on non-500 errors

- GIVEN `POST /api/security/entity_store/resolution/link` returns HTTP 400
- WHEN the provider creates the entity link resource
- THEN the provider SHALL immediately return an error diagnostic without retrying

#### Scenario: Create fails after exceeding the Create timeout

- GIVEN `POST /api/security/entity_store/resolution/link` returns HTTP 500 for all attempts
- WHEN the provider creates the entity link resource
- THEN the provider SHALL return an error diagnostic after the Create timeout is exceeded
- AND the diagnostic SHALL describe the final HTTP 500 response

## ADDED Requirements

### Requirement: Acceptance tests for entity-link enforce entity store isolation (REQ-ESL-TEST-ISOLATION-001)

Every acceptance test in the entity-link package that manages the entity store resource SHALL
register a `t.Cleanup` function that uninstalls the entity store and waits for `not_installed`
state (using the same shared helper as the entity store package, as specified in
`kibana-security-entity-store` delta REQ-TEST-ISOLATION-001). This ensures that entity-link
tests do not leave residual entity types in the singleton store that contaminate subsequent tests.

#### Scenario: Entity-link test cleanup leaves store in not_installed state

- GIVEN `TestAccResourceSecurityEntityStoreEntityLink` registers `t.Cleanup(cleanupEntityStore)`
- WHEN the test body completes (success or failure)
- THEN `cleanupEntityStore` SHALL run and poll until the entity store reports `not_installed`
- AND the next test in the suite SHALL find the store in a clean state
