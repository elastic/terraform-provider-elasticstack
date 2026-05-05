## ADDED Requirements

### Requirement: Envelope owns the Create and Update preludes

The system SHALL implement `Create` and `Update` on `NewElasticsearchResource[T]` by deserializing the planned model into `T`, deriving the write resource ID from the model, resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, and invoking the corresponding concrete callback. The concrete create and update callbacks SHALL be invoked with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)`.

#### Scenario: Successful create sets state from returned model

- **WHEN** the concrete create function returns `(model, nil)`
- **THEN** `resp.State.Set` SHALL be called with the returned model `T`
- **AND** response diagnostics SHALL contain no errors

#### Scenario: Successful update sets state from returned model

- **WHEN** the concrete update function returns `(model, nil)`
- **THEN** `resp.State.Set` SHALL be called with the returned model `T`
- **AND** response diagnostics SHALL contain no errors

#### Scenario: Create function error short-circuits state mutation

- **WHEN** the concrete create function returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: Update function error short-circuits state mutation

- **WHEN** the concrete update function returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: Client resolution failure short-circuits create

- **WHEN** `GetElasticsearchClient` returns error diagnostics during `Create`
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete create function SHALL NOT be invoked
- **AND** state SHALL remain untouched

#### Scenario: Client resolution failure short-circuits update

- **WHEN** `GetElasticsearchClient` returns error diagnostics during `Update`
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the concrete update function SHALL NOT be invoked
- **AND** state SHALL remain untouched

#### Scenario: Create and update may use the same callback

- **WHEN** a concrete Elasticsearch resource has identical create and update API behavior
- **THEN** the resource SHALL be able to pass the same callback as both the create and update callback

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

## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Elasticsearch resource behavior

The system SHALL provide a generic constructor `NewElasticsearchResource[T]()` that returns an envelope owning shared Elasticsearch resource behavior. The envelope SHALL provide Metadata, Schema, Create, Read, Update, Delete, and Configure behavior, and SHALL satisfy `resource.Resource`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewElasticsearchResource[T](component, name, schemaFactory, readFunc, deleteFunc, createFunc, updateFunc)` is called with non-nil callbacks
- **THEN** the returned value SHALL provide the shared Metadata, Schema, Create, Read, Update, Delete, and Configure methods
- **AND** the returned value SHALL satisfy `resource.Resource`
- **AND** concrete resources embedding the returned value SHALL NOT need thin Create or Update wrappers when their behavior fits the callback contract

### Requirement: Model type constraint exposes ID and connection block

The system SHALL define a type constraint `ElasticsearchResourceModel` requiring `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`. Concrete `Data` types SHALL satisfy this constraint by declaring value-receiver getter methods over their existing `ID`, natural write identity, and `ElasticsearchConnection` fields.

#### Scenario: Concrete model satisfies the constraint

- **WHEN** a concrete `Data` struct declares `GetID() types.String` returning `d.ID`, `GetResourceID() types.String` returning its plan-safe write identity field, and `GetElasticsearchConnection() types.List` returning `d.ElasticsearchConnection`
- **THEN** that struct SHALL satisfy `ElasticsearchResourceModel`
- **AND** SHALL be usable as the type parameter to `NewElasticsearchResource[T]`

## REMOVED Requirements

### Requirement: Envelope does not implement Create, Update, or ImportState

**Reason**: The envelope now owns Create and Update for resources that fit the standard Elasticsearch lifecycle. Keeping this requirement would contradict the new complete-resource contract.

**Migration**: ImportState remains opt-in under `Envelope keeps ImportState opt-in`. Concrete resources that previously supplied thin Create and Update wrappers SHALL migrate to create and update callbacks passed to `NewElasticsearchResource`.
