# entitycore-resource-envelope Specification

## Purpose

The entitycore resource envelope centralizes common Terraform Plugin Framework behavior for Elasticsearch-backed resources that share the standard connection block, lifecycle preludes, model contract, composite ID handling, and opt-in import convention.
## Requirements
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

### Requirement: Envelope injects elasticsearch_connection block into schema

The system SHALL inject the `elasticsearch_connection` block into the schema returned by the concrete schema factory before exposing it via the `Schema` method.

#### Scenario: Schema includes injected connection block

- **WHEN** an envelope is constructed with a schema factory that returns a schema lacking an `elasticsearch_connection` block
- **THEN** calling `Schema` on the envelope SHALL return a schema that includes the `elasticsearch_connection` block produced by the canonical provider helper
- **AND** the concrete schema attributes and other blocks SHALL remain unchanged

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

Create and update callbacks SHALL share the type `WriteFunc[T]` and receive a `WriteRequest[T]` containing `Plan`, `Prior`, `Config`, and `WriteID`. `Prior` SHALL be a `*T`: `nil` for create invocations and a non-nil pointer to the decoded prior state model for update invocations. Callbacks that distinguish create from update SHALL inspect `req.Prior == nil`.

Create and update callbacks SHALL return `WriteResult[T]` carrying the written model used for read-after-write identity resolution.

#### Scenario: Create callback receives nil Prior and config

- **WHEN** `Create` runs for a resource whose callback fits the envelope contract
- **THEN** the callback SHALL receive `WriteRequest[T]` with `Prior == nil`
- **AND** the callback SHALL receive the planned model and the raw Terraform config in the request object

#### Scenario: Update callback receives prior state and config

- **WHEN** `Update` runs for a resource whose callback fits the envelope contract
- **THEN** the callback SHALL receive `WriteRequest[T]` with `Prior` pointing at the decoded prior-state model
- **AND** the callback SHALL receive both the planned model and the raw Terraform config in the request object

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

#### Scenario: Single WriteFunc may serve both Create and Update

- **WHEN** a concrete Elasticsearch resource wires the same `WriteFunc[T]` value into both `Create` and `Update` slots of `ElasticsearchResourceOptions[T]`
- **THEN** the envelope SHALL invoke that single function for both Terraform operations
- **AND** the function SHALL distinguish create from update by inspecting `req.Prior == nil`

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

### Requirement: Envelope owns the Delete prelude

The system SHALL implement `Delete` by deserializing the prior state into the generic model `T`, parsing the composite ID with `clients.CompositeIDFromStrFw`, resolving the scoped Elasticsearch client from the model's connection block, and invoking the concrete delete function with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)`.

#### Scenario: Successful delete returns nil diagnostics

- **WHEN** the concrete delete function returns no diagnostics
- **THEN** the response diagnostics SHALL contain no errors and Plugin Framework SHALL remove the resource from state per its standard contract

#### Scenario: Delete function error is appended to response diagnostics

- **WHEN** the concrete delete function returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`

#### Scenario: Composite ID parse failure short-circuits delete

- **WHEN** `CompositeIDFromStrFw` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete delete function SHALL NOT be invoked

#### Scenario: Client resolution failure short-circuits delete

- **WHEN** `GetElasticsearchClient` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete delete function SHALL NOT be invoked

### Requirement: Envelope requires a non-nil delete callback

The system SHALL require concrete resources to supply a non-nil delete callback. Resources whose API has no delete operation SHALL pass a callback that returns `nil` diagnostics. The envelope SHALL NOT special-case a nil callback to mean "no API call".

#### Scenario: Constructor accepts a no-op delete callback

- **WHEN** the delete callback is supplied as a function that returns `nil` diagnostics without performing an API call
- **THEN** the envelope SHALL invoke that callback during `Delete` and proceed normally, removing the resource from state

### Requirement: Envelope keeps ImportState opt-in

The system SHALL NOT implement `ImportState` on the envelope. Import support SHALL remain opt-in for concrete resources, matching the provider-wide convention. Concrete resources that support import SHALL implement `ImportState` themselves, typically as a passthrough on the `id` attribute.

#### Scenario: Envelope resource without ImportState does not satisfy ResourceWithImportState

- **WHEN** an envelope is constructed via `NewElasticsearchResource[T]`
- **THEN** the returned resource SHALL satisfy `resource.Resource`
- **AND** SHALL satisfy `resource.ResourceWithConfigure`
- **AND** SHALL NOT satisfy `resource.ResourceWithImportState`

#### Scenario: Concrete resource adds ImportState passthrough

- **WHEN** a concrete resource that embeds `*ElasticsearchResource[T]` defines its own `ImportState` method using `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`
- **THEN** that resource SHALL satisfy `resource.ResourceWithImportState`
- **AND** the `id` attribute SHALL be set to the supplied import identifier

### Requirement: Model type constraint exposes ID and connection block

The system SHALL define a type constraint `ElasticsearchResourceModel` requiring `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`. Concrete `Data` types SHALL satisfy this constraint by declaring value-receiver getter methods over their existing `ID`, natural write identity, and `ElasticsearchConnection` fields.

#### Scenario: Concrete model satisfies the constraint

- **WHEN** a concrete `Data` struct declares `GetID() types.String` returning `d.ID`, `GetResourceID() types.String` returning its plan-safe write identity field, and `GetElasticsearchConnection() types.List` returning `d.ElasticsearchConnection`
- **THEN** that struct SHALL satisfy `ElasticsearchResourceModel`
- **AND** SHALL be usable as the type parameter to `NewElasticsearchResource[T]`

### Requirement: Envelope preserves resource type names exactly

The system SHALL allow concrete migrations to preserve their existing Terraform type name. Each migrated security resource SHALL initialize the envelope with the literal name suffix that produces its current type name.

#### Scenario: Migrated user resource preserves its type name

- **WHEN** the user resource is migrated to the envelope using `NewElasticsearchResource[Data]("security_user", opts)`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_user`

#### Scenario: Migrated system_user resource preserves its type name

- **WHEN** the system user resource is migrated to the envelope using name suffix `security_system_user`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_system_user`

#### Scenario: Migrated role resource preserves its type name

- **WHEN** the role resource is migrated to the envelope using name suffix `security_role`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_role`

#### Scenario: Migrated role_mapping resource preserves its type name

- **WHEN** the role mapping resource is migrated to the envelope using name suffix `security_role_mapping`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_role_mapping`

### Requirement: Envelope coexists with ResourceBase-only entities

The system SHALL NOT change the behavior of resources that embed `*ResourceBase` directly. Resources that do not migrate to the envelope SHALL continue to operate with their existing Configure, Metadata, and Client wiring.

#### Scenario: ResourceBase-only resource continues to work

- **WHEN** a resource that embeds `*ResourceBase` (without the envelope) is configured and used
- **THEN** its Configure, Metadata, and Client behavior SHALL remain unchanged

