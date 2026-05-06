# entitycore-kibana-resource-envelope Specification

## Purpose
TBD - created by archiving change kibana-resource-envelope. Update Purpose after archive.
## Requirements
### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL provide a generic constructor `NewKibanaResource[T]()` that returns an envelope owning shared Kibana resource behavior. The envelope SHALL provide Metadata, Schema, Configure, Create, Read, Update, and Delete behavior, and SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewKibanaResource[T](component, name, schemaFactory, readFunc, deleteFunc, createFunc, updateFunc)` is called with non-nil callbacks
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the returned value SHALL NOT satisfy `resource.ResourceWithImportState`

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewKibanaResource[T](ComponentKibana, "maintenance_window", …)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_kibana_maintenance_window`

---

### Requirement: KibanaResourceModel interface defines the model contract

The system SHALL define a `KibanaResourceModel` interface with four value-receiver methods: `GetID() types.String`, `GetResourceID() types.String`, `GetSpaceID() types.String`, and `GetKibanaConnection() types.List`. Concrete resource models SHALL satisfy this interface to be used with `NewKibanaResource[T]`.

- `GetID` SHALL return the verbatim Terraform state identifier (may be a composite `<space>/<resourceID>` string or a plain ID, depending on the resource).
- `GetResourceID` SHALL return the primary API key used in write and read operations — either a user-specified name (user-ID resources) or the API-assigned UUID (UUID resources). For UUID resources, `GetResourceID` MAY return the same value as `GetID`.
- `GetSpaceID` SHALL return the Kibana space identifier for API calls.
- `GetKibanaConnection` SHALL return the decoded `kibana_connection` block list.

#### Scenario: User-ID resource implements the interface

- **WHEN** a model struct provides `GetID()` returning the composite state ID, `GetResourceID()` returning the user-specified name, `GetSpaceID()` returning the space, and `GetKibanaConnection()` returning the connection block
- **THEN** it SHALL satisfy `KibanaResourceModel` and be usable as the type parameter `T` for `NewKibanaResource[T]`

#### Scenario: API-UUID resource implements the interface

- **WHEN** a model struct provides `GetID()` and `GetResourceID()` both returning the API-assigned UUID (same value), `GetSpaceID()` returning the space, and `GetKibanaConnection()` returning the connection block
- **THEN** it SHALL satisfy `KibanaResourceModel` and be usable as the type parameter `T` for `NewKibanaResource[T]`

---

### Requirement: Envelope optionally enforces server version requirements

The system SHALL define a `WithVersionRequirements` optional interface that Kibana entity models may implement to declare minimum Kibana server version requirements. When a decoded model satisfies this interface, the generic Kibana resource envelope SHALL evaluate the requirements after scoped client resolution and before invoking the concrete lifecycle callback. The envelope SHALL enforce version requirements in Create, Read, and Update.

The `WithVersionRequirements` interface SHALL reuse `DataSourceVersionRequirement` for the requirement type.

#### Scenario: Model with version requirements short-circuits Create

- **WHEN** a model implements `WithVersionRequirements` and `GetVersionRequirements()` returns error diagnostics
- **THEN** the envelope's Create method SHALL append those diagnostics and SHALL NOT invoke the create callback

#### Scenario: Model with version requirements short-circuits Read

- **WHEN** a model implements `WithVersionRequirements` and `GetVersionRequirements()` returns error diagnostics
- **THEN** the envelope's Read method SHALL append those diagnostics and SHALL NOT invoke the read callback

#### Scenario: Model with version requirements short-circuits Update

- **WHEN** a model implements `WithVersionRequirements` and `GetVersionRequirements()` returns error diagnostics
- **THEN** the envelope's Update method SHALL append those diagnostics and SHALL NOT invoke the update callback

### Requirement: Envelope injects kibana_connection block into schema

The system SHALL inject the `kibana_connection` block into the schema returned by the concrete schema factory before exposing it via the `Schema` method.

#### Scenario: Schema includes injected connection block

- **WHEN** an envelope is constructed with a schema factory that returns a schema lacking a `kibana_connection` block
- **THEN** calling `Schema` on the envelope SHALL return a schema that includes the `kibana_connection` block produced by the canonical provider helper
- **AND** the concrete schema attributes and other blocks SHALL remain unchanged

#### Scenario: Schema injection does not mutate the factory's return value

- **WHEN** the schema factory is called multiple times via repeated `Schema` calls
- **THEN** each call SHALL produce an independent schema with the `kibana_connection` block
- **AND** the original schema value returned by the factory SHALL NOT be mutated

---

### Requirement: Envelope owns the Create prelude

The system SHALL implement `Create` by deserializing the plan into the generic model `T`, resolving `spaceID` from `model.GetSpaceID()`, validating that `spaceID` is non-empty, resolving the scoped Kibana client via `GetKibanaClient`, and invoking the create callback. The create callback SHALL be invoked with `(context, *clients.KibanaScopedClient, spaceID string, plan T)`. The envelope SHALL NOT validate `GetResourceID()` during Create.

#### Scenario: Successful create persists returned model

- **WHEN** the create callback returns `(model, nil)` with no errors
- **THEN** `resp.State.Set` SHALL be called with the returned model

#### Scenario: Empty spaceID short-circuits create

- **WHEN** `model.GetSpaceID()` returns an empty string
- **THEN** an error diagnostic SHALL be appended
- **AND** the create callback SHALL NOT be invoked

#### Scenario: Client resolution failure short-circuits create

- **WHEN** `GetKibanaClient` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the create callback SHALL NOT be invoked

#### Scenario: Create callback error short-circuits state persistence

- **WHEN** the create callback returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `resp.State.Set` SHALL NOT be called

---

### Requirement: Envelope owns the Read prelude with composite-ID-or-fallback identity resolution

The system SHALL implement `Read` by deserializing the prior state into the generic model `T`, resolving `resourceID` and `spaceID` using the composite-ID-or-fallback rule, resolving the scoped Kibana client, validating that `resourceID` is non-empty, and invoking the read callback. The read callback SHALL be invoked with `(context, *clients.KibanaScopedClient, resourceID string, spaceID string, T)`.

The composite-ID-or-fallback rule SHALL be: attempt to parse `model.GetID()` as a composite ID. The parse result (nil on failure) SHALL be inspected, but any diagnostics returned by the parse function SHALL be discarded — parse failure is treated as a non-error "not composite" signal, not as an error condition. On successful parse, use `compID.ResourceID` as `resourceID` and `compID.ClusterID` as `spaceID`. On parse failure, use `model.GetResourceID().ValueString()` as `resourceID` and `model.GetSpaceID().ValueString()` as `spaceID`.

#### Scenario: Successful read sets state from returned model

- **WHEN** the read callback returns `(model, true, nil)` (entity found, no errors)
- **THEN** `resp.State.Set` SHALL be called with the returned model

#### Scenario: Not-found read removes resource from state

- **WHEN** the read callback returns `(_, false, nil)` (entity missing, no errors)
- **THEN** `resp.State.RemoveResource` SHALL be called
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: Composite ID parse succeeds — parsed parts used

- **WHEN** `model.GetID()` returns a valid composite ID (e.g. `"default/my-stream"`)
- **THEN** `resourceID` SHALL be `"my-stream"` and `spaceID` SHALL be `"default"`
- **AND** `model.GetResourceID()` and `model.GetSpaceID()` SHALL NOT be used for identity resolution

#### Scenario: Composite ID parse fails — fallback to model methods

- **WHEN** `model.GetID()` returns a plain value that is not a valid composite ID (e.g. `"abc-uuid"`)
- **THEN** `resourceID` SHALL be `model.GetResourceID().ValueString()` and `spaceID` SHALL be `model.GetSpaceID().ValueString()`

#### Scenario: Empty resourceID after resolution short-circuits read

- **WHEN** composite ID parse fails and `model.GetResourceID()` returns an empty string
- **THEN** an error diagnostic SHALL be appended
- **AND** the read callback SHALL NOT be invoked

#### Scenario: Client resolution failure short-circuits read

- **WHEN** `GetKibanaClient` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the read callback SHALL NOT be invoked

---

### Requirement: Envelope owns the Update prelude and passes both plan and prior state

The system SHALL implement `Update` using the same composite-ID-or-fallback identity resolution as Read (applied to the plan model), resolving the scoped Kibana client, and invoking the update callback. The update callback SHALL be invoked with `(context, *clients.KibanaScopedClient, resourceID string, spaceID string, plan T, prior T)` where `plan` is decoded from `req.Plan` and `prior` is decoded from `req.State`.

#### Scenario: Successful update persists returned model

- **WHEN** the update callback returns `(model, nil)` with no errors
- **THEN** `resp.State.Set` SHALL be called with the returned model

#### Scenario: Empty resourceID after resolution short-circuits update

- **WHEN** composite ID parse fails and `plan.GetResourceID()` returns an empty string
- **THEN** an error diagnostic SHALL be appended
- **AND** the update callback SHALL NOT be invoked

#### Scenario: Update callback receives prior state

- **WHEN** the update callback is invoked
- **THEN** the fifth argument SHALL be the decoded plan model
- **AND** the sixth argument SHALL be the decoded prior state model

#### Scenario: Client resolution failure short-circuits update

- **WHEN** `GetKibanaClient` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the update callback SHALL NOT be invoked

---

### Requirement: Envelope owns the Delete prelude

The system SHALL implement `Delete` using the same composite-ID-or-fallback identity resolution as Read (applied to the state model), resolving the scoped Kibana client, validating that `resourceID` is non-empty, and invoking the delete callback. The delete callback SHALL be invoked with `(context, *clients.KibanaScopedClient, resourceID string, spaceID string, T)`.

#### Scenario: Delete callback is invoked with resolved identity

- **WHEN** the state model resolves to a non-empty `resourceID` and `spaceID`
- **THEN** the delete callback SHALL be invoked once with those values

#### Scenario: Empty resourceID after resolution short-circuits delete

- **WHEN** composite ID parse fails and `model.GetResourceID()` returns an empty string
- **THEN** an error diagnostic SHALL be appended
- **AND** the delete callback SHALL NOT be invoked

#### Scenario: Client resolution failure short-circuits delete

- **WHEN** `GetKibanaClient` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** the delete callback SHALL NOT be invoked

---

### Requirement: Nil create or update callback surfaces a configuration error diagnostic

The system SHALL check that the `createFunc` and `updateFunc` callbacks are non-nil before invoking them. If either is nil at invocation time, the envelope SHALL append a configuration error diagnostic and SHALL NOT invoke the nil callback. This nil check SHALL take precedence over all other pre-invocation checks (such as spaceID validation and client resolution).

#### Scenario: Nil create callback produces configuration error before other checks

- **WHEN** `NewKibanaResource` is called with a nil `createFunc` and `Create` is invoked
- **THEN** an error diagnostic describing the nil callback configuration SHALL be appended
- **AND** no other Create prelude checks (spaceID, client resolution) SHALL produce diagnostics

#### Scenario: Nil update callback produces configuration error before other checks

- **WHEN** `NewKibanaResource` is called with a nil `updateFunc` and `Update` is invoked
- **THEN** an error diagnostic describing the nil callback configuration SHALL be appended
- **AND** no other Update prelude checks (resourceID, client resolution) SHALL produce diagnostics

---

### Requirement: PlaceholderKibanaWriteCallbacks returns callbacks that error when invoked

The system SHALL provide `PlaceholderKibanaWriteCallbacks[T]()` that returns a `(KibanaCreateFunc[T], KibanaUpdateFunc[T])` pair. Each placeholder SHALL add a predictable error diagnostic if invoked. This function exists for concrete resources that override Create and/or Update on the envelope struct and therefore never intend the envelope's callbacks to be called.

#### Scenario: Placeholder create callback returns error diagnostic

- **WHEN** the placeholder `KibanaCreateFunc[T]` is invoked
- **THEN** it SHALL return a non-empty `diag.Diagnostics` with an error-level diagnostic describing the misconfiguration

#### Scenario: Placeholder update callback returns error diagnostic

- **WHEN** the placeholder `KibanaUpdateFunc[T]` is invoked
- **THEN** it SHALL return a non-empty `diag.Diagnostics` with an error-level diagnostic describing the misconfiguration

---

### Requirement: Streams resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/kibana/streams` so that its `Resource` struct embeds `*entitycore.KibanaResource[streamModel]` returned by `NewKibanaResource`. The `streamModel` SHALL implement `KibanaResourceModel` with `GetResourceID()` returning `m.Name` and `GetSpaceID()` returning `m.SpaceID`. The resource's schema, CRUD method bodies, and Terraform schema attributes SHALL remain functionally equivalent after migration.

#### Scenario: Streams resource satisfies envelope interface after migration

- **WHEN** `newResource()` is called for the streams resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope

---

### Requirement: Maintenance window resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/kibana/maintenance_window` so that its `Resource` struct embeds `*entitycore.KibanaResource[Model]` returned by `NewKibanaResource`. The `Model` SHALL implement `KibanaResourceModel` with `GetResourceID()` returning `m.ID` (the API-assigned UUID) and `GetSpaceID()` returning `m.SpaceID`. The resource's schema, CRUD method bodies, and Terraform schema attributes SHALL remain functionally equivalent after migration.

#### Scenario: Maintenance window resource satisfies envelope interface after migration

- **WHEN** `newResource()` is called for the maintenance window resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope

