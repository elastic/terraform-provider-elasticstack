# entitycore-resource-envelope Specification

## Purpose
TBD - created by archiving change elasticsearch-resource-envelope. Update Purpose after archive.
## Requirements
### Requirement: Envelope constructor produces a valid Resource

The system SHALL provide a generic constructor `NewElasticsearchResource[T]()` that returns a value satisfying `resource.Resource`.

#### Scenario: Constructor returns valid resource

- **WHEN** `NewElasticsearchResource[T](component, name, schemaFactory, readFunc, deleteFunc)` is called with non-nil callbacks
- **THEN** the returned value SHALL satisfy `resource.Resource`
- **AND** the returned value SHALL satisfy `resource.ResourceWithConfigure`
- **AND** the returned value SHALL NOT satisfy `resource.ResourceWithImportState` (import is opt-in for concrete resources)

### Requirement: Envelope owns Configure and Metadata

The system SHALL implement `Configure` and `Metadata` on the envelope using the same provider-data conversion and type-name composition rules as `ResourceBase`. The configured `*clients.ProviderClientFactory` SHALL be reachable from the envelope's Read and Delete preludes.

#### Scenario: Configure stores the provider client factory

- **WHEN** `Configure` receives provider data that converts successfully to `*clients.ProviderClientFactory`
- **THEN** the envelope SHALL retain that factory for use by subsequent Read and Delete calls

#### Scenario: Configure does not store a client after diagnostic failure

- **WHEN** `Configure` has appended the conversion diagnostics and the response has error-level diagnostics
- **THEN** the envelope SHALL not assign a factory from that conversion, and SHALL leave unchanged any factory previously stored by an earlier successful `Configure` call

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewElasticsearchResource[T](ComponentElasticsearch, "security_user", …)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_elasticsearch_security_user`

### Requirement: Envelope injects elasticsearch_connection block into schema

The system SHALL inject the `elasticsearch_connection` block into the schema returned by the concrete schema factory before exposing it via the `Schema` method.

#### Scenario: Schema includes injected connection block

- **WHEN** an envelope is constructed with a schema factory that returns a schema lacking an `elasticsearch_connection` block
- **THEN** calling `Schema` on the envelope SHALL return a schema that includes the `elasticsearch_connection` block produced by the canonical provider helper
- **AND** the concrete schema attributes and other blocks SHALL remain unchanged

### Requirement: Envelope owns the Read prelude

The system SHALL implement `Read` by deserializing the prior state into the generic model `T`, parsing the composite ID with `clients.CompositeIDFromStrFw`, resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, and invoking the concrete read function. The concrete read function SHALL be invoked with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)` where `resourceID` is `compID.ResourceID`.

#### Scenario: Successful read sets state from returned model

- **WHEN** the concrete read function returns `(model, true, nil)` (entity found, no errors)
- **THEN** `resp.State.Set` SHALL be called with the returned model `T`

#### Scenario: Not-found read removes resource from state

- **WHEN** the concrete read function returns `(_, false, nil)` (entity missing, no errors)
- **THEN** `resp.State.RemoveResource` SHALL be called
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: Read function error short-circuits state mutation

- **WHEN** the concrete read function returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** neither `resp.State.Set` nor `resp.State.RemoveResource` SHALL be called

#### Scenario: Composite ID parse failure short-circuits read

- **WHEN** `CompositeIDFromStrFw` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete read function SHALL NOT be invoked
- **AND** state SHALL remain untouched

#### Scenario: Client resolution failure short-circuits read

- **WHEN** `GetElasticsearchClient` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete read function SHALL NOT be invoked
- **AND** state SHALL remain untouched

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

### Requirement: Envelope does not implement ImportState

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

The system SHALL define a type constraint `ElasticsearchResourceModel` requiring `GetID() types.String` and `GetElasticsearchConnection() types.List`. Concrete `Data` types SHALL satisfy this constraint by declaring value-receiver getter methods over their existing `ID` and `ElasticsearchConnection` fields.

#### Scenario: Concrete model satisfies the constraint

- **WHEN** a concrete `Data` struct declares `GetID() types.String` returning `d.ID` and `GetElasticsearchConnection() types.List` returning `d.ElasticsearchConnection`
- **THEN** that struct SHALL satisfy `ElasticsearchResourceModel`
- **AND** SHALL be usable as the type parameter to `NewElasticsearchResource[T]`

### Requirement: Envelope preserves resource type names exactly

The system SHALL allow concrete migrations to preserve their existing Terraform type name. Each migrated security resource SHALL initialize the envelope with the component namespace and literal name suffix that produce its current type name.

#### Scenario: Migrated user resource preserves its type name

- **WHEN** the user resource is migrated to the envelope using component `elasticsearch` and name `security_user`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_user`

#### Scenario: Migrated system_user resource preserves its type name

- **WHEN** the system user resource is migrated to the envelope using component `elasticsearch` and name `security_system_user`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_system_user`

#### Scenario: Migrated role resource preserves its type name

- **WHEN** the role resource is migrated to the envelope using component `elasticsearch` and name `security_role`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_role`

#### Scenario: Migrated role_mapping resource preserves its type name

- **WHEN** the role mapping resource is migrated to the envelope using component `elasticsearch` and name `security_role_mapping`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_role_mapping`

### Requirement: Envelope coexists with ResourceBase-only entities

The system SHALL NOT change the behavior of resources that embed `*ResourceBase` directly. Resources that do not migrate to the envelope SHALL continue to operate with their existing Configure, Metadata, and Client wiring.

#### Scenario: ResourceBase-only resource continues to work

- **WHEN** a resource that embeds `*ResourceBase` (without the envelope) is configured and used
- **THEN** its Configure, Metadata, and Client behavior SHALL remain unchanged

