# typed-client-bootstrap Specification

## Purpose
TBD - created by archiving change typed-client-bootstrap. Update Purpose after archive.
## Requirements
### Requirement: ElasticsearchScopedClient exposes a typed client accessor
`ElasticsearchScopedClient` SHALL provide a method `GetESTypedClient()` that returns `*elasticsearch.TypedClient`.

#### Scenario: Typed client is accessible after initialization
- **WHEN** an `ElasticsearchScopedClient` has been constructed with a valid `*elasticsearch.Client`
- **THEN** calling `GetESTypedClient()` returns a non-nil `*elasticsearch.TypedClient`

### Requirement: Typed client is lazily initialized
`GetESTypedClient()` SHALL convert the underlying `*elasticsearch.Client` to `*elasticsearch.TypedClient` using `ToTyped()` only on the first invocation.

#### Scenario: First call creates the typed client
- **WHEN** `GetESTypedClient()` is called for the first time on an `ElasticsearchScopedClient` instance
- **THEN** `ToTyped()` is invoked exactly once on the underlying client

#### Scenario: Subsequent calls reuse the cached typed client
- **GIVEN** `GetESTypedClient()` has already been called once
- **WHEN** `GetESTypedClient()` is called again on the same instance
- **THEN** `ToTyped()` is NOT invoked again and the same `*elasticsearch.TypedClient` pointer is returned

### Requirement: Typed client initialization is thread-safe
The lazy initialization of the typed client SHALL be safe for concurrent use by multiple goroutines.

#### Scenario: Concurrent access during first call
- **WHEN** multiple goroutines call `GetESTypedClient()` concurrently before any prior invocation
- **THEN** `ToTyped()` is invoked exactly once and all goroutines receive the same `*elasticsearch.TypedClient` pointer

### Requirement: Existing untyped client accessor remains unchanged
`GetESClient()` on `ElasticsearchScopedClient` SHALL continue to behave exactly as before this change.

#### Scenario: Untyped accessor is unaffected
- **WHEN** code calls `GetESClient()` on an `ElasticsearchScopedClient`
- **THEN** it returns the same `*elasticsearch.Client` and exhibits the same error behavior as prior to this change

