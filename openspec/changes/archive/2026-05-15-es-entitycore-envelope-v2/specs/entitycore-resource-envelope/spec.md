## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Elasticsearch resource behavior

The system SHALL provide a generic constructor `NewElasticsearchResource[T]()` that returns an envelope owning shared Elasticsearch resource behavior. The Elasticsearch envelope constructor SHALL take the Terraform type-name suffix and an options struct, not a component enum and a positional callback list. The envelope SHALL provide Metadata, Schema, Create, Read, Update, Delete, and Configure behavior, and SHALL satisfy `resource.Resource`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewElasticsearchResource[T]("security_user", opts)` is called with non-nil required callbacks in `opts`
- **THEN** the returned value SHALL provide the shared Metadata, Schema, Create, Read, Update, Delete, and Configure methods
- **AND** the returned value SHALL satisfy `resource.Resource`
- **AND** concrete resources embedding the returned value SHALL NOT need thin Create or Update wrappers when their behavior fits the callback contract

### Requirement: Envelope owns Configure and Metadata

The system SHALL implement `Configure` and `Metadata` on the Elasticsearch envelope using the same provider-data conversion and type-name composition rules as `ResourceBase`, with the Elasticsearch namespace segment implied by the envelope rather than passed by the caller. The configured `*clients.ProviderClientFactory` SHALL be reachable from the envelope's Read, Create, Update, and Delete preludes.

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewElasticsearchResource[T]("security_user", opts)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_elasticsearch_security_user`

### Requirement: Envelope owns the Read prelude

The system SHALL implement `Read` by deserializing the prior state into the generic model `T`, resolving a stable read identity from the model and/or composite ID, resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, enforcing any optional version requirements declared by the model, and invoking the concrete read function. The concrete read function SHALL continue to be invoked with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)`.

#### Scenario: Read uses model-declared read identity when available

- **WHEN** the decoded state model implements `WithReadResourceID` and returns a non-empty value
- **THEN** the envelope SHALL use that value as the read identity for `readFunc`

#### Scenario: Read falls back to composite ID resource segment

- **WHEN** the decoded state model does not implement `WithReadResourceID` or returns an empty read identity
- **THEN** the envelope SHALL parse the composite `id` from state and use `compID.ResourceID` as the read identity for `readFunc`

#### Scenario: Version requirements short-circuit read

- **WHEN** the decoded state model implements `WithVersionRequirements` and requirement evaluation returns error diagnostics or an unsupported-version diagnostic
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete read function SHALL NOT be invoked

### Requirement: Envelope owns the Create and Update preludes

The system SHALL implement `Create` and `Update` on `NewElasticsearchResource[T]` by deserializing the relevant framework inputs, deriving the write resource ID from the model, resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, enforcing any optional version requirements declared by the planned model, invoking the corresponding concrete callback with a structured request object, and then invoking `readFunc` with the model returned by the callback. State SHALL be set from the model returned by `readFunc`, not directly from the concrete callback.

Create callbacks SHALL receive `ElasticsearchCreateRequest[T]` containing `Plan`, `Config`, and `WriteID`.

Update callbacks SHALL receive `ElasticsearchUpdateRequest[T]` containing `Plan`, `Prior`, `Config`, and `WriteID`.

#### Scenario: Update callback receives prior state and config

- **WHEN** `Update` runs for a resource whose callback fits the envelope contract
- **THEN** the callback SHALL receive both the planned model and the prior state model in `ElasticsearchUpdateRequest[T]`
- **AND** the callback SHALL receive the raw Terraform config in the request object

#### Scenario: Read-after-write uses model-declared read identity when available

- **WHEN** a successful create or update callback returns a model implementing `WithReadResourceID` with a non-empty value
- **THEN** the envelope SHALL use that read identity for the subsequent `readFunc` call instead of `WriteID`

#### Scenario: Read-after-write falls back to write identity

- **WHEN** a successful create or update callback returns a model that does not implement `WithReadResourceID` or returns an empty value
- **THEN** the envelope SHALL call `readFunc` using `WriteID`

#### Scenario: Version requirements short-circuit create or update

- **WHEN** the planned model implements `WithVersionRequirements` and requirement evaluation returns error diagnostics or an unsupported-version diagnostic
- **THEN** the diagnostics SHALL be appended to the response
- **AND** the concrete create or update callback SHALL NOT be invoked

### Requirement: Envelope supports post-read side effects

The system SHALL allow Elasticsearch envelope users to provide an optional post-read hook. After a successful read flow that sets Terraform state, the envelope SHALL invoke the post-read hook with the scoped client, the model persisted to state, and the framework private-state handle.

The hook SHALL NOT run when the entity is not found, when `readFunc` returns error diagnostics, or when state persistence fails.

#### Scenario: Post-read hook persists private state after read

- **WHEN** `readFunc` returns `(model, true, nil)` and `resp.State.Set` succeeds
- **THEN** the envelope SHALL invoke the configured post-read hook with the persisted model and `resp.Private`

### Requirement: Shared version requirement type is envelope-neutral

The system SHALL define the shared requirement type as `VersionRequirement`, not `DataSourceVersionRequirement`, and the optional `WithVersionRequirements` interface SHALL return `[]VersionRequirement`.

#### Scenario: Resource and data source envelopes use the same requirement type

- **WHEN** a model implements `WithVersionRequirements`
- **THEN** both Kibana and Elasticsearch envelopes SHALL accept the same `VersionRequirement` return type
