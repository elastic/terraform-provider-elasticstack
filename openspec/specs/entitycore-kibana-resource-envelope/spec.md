# entitycore-kibana-resource-envelope Specification

## Purpose
TBD - created by archiving change kibana-resource-envelope. Update Purpose after archive.
## Requirements
### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL provide a generic constructor `NewKibanaResource[T]()` that accepts a `KibanaResourceOptions[T]` options struct (not a positional callback list) and returns an envelope owning shared Kibana resource behavior. The envelope SHALL provide Metadata, Schema, Configure, Create, Read, Update, and Delete behavior, and SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewKibanaResource[T](component, name, opts)` is called with a `KibanaResourceOptions[T]` containing non-nil required callbacks
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the returned value SHALL NOT satisfy `resource.ResourceWithImportState`

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewKibanaResource[T](ComponentKibana, "maintenance_window", opts)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_kibana_maintenance_window`

#### Scenario: Action connector migration uses the envelope without changing behavior

- **WHEN** `internal/kibana/connectors/` is migrated to embed `*entitycore.KibanaResource[Model]` returned by `NewKibanaResource`
- **THEN** the resource SHALL continue to expose the same schema, CRUD semantics, import behavior, and version-gated update behavior as before migration
- **AND** the resource SHALL remain usable as a Terraform `resource.Resource` and `resource.ResourceWithConfigure` implementation

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

The `WithVersionRequirements` interface SHALL use the shared `VersionRequirement` type (the same type returned by Elasticsearch envelope models that declare version requirements).

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

The system SHALL check that the `Create` and `Update` callbacks in `KibanaResourceOptions[T]` are non-nil before invoking them. If either is nil at invocation time, the envelope SHALL append a configuration error diagnostic and SHALL NOT proceed with other prelude checks (spaceID validation, client resolution). This nil check SHALL take precedence over all other pre-invocation checks.

#### Scenario: Nil create callback produces configuration error before other checks

- **WHEN** `KibanaResourceOptions.Create` is nil and `Create` is invoked
- **THEN** an error diagnostic describing the nil callback configuration SHALL be appended
- **AND** no other Create prelude checks (spaceID, client resolution) SHALL produce diagnostics

#### Scenario: Nil update callback produces configuration error before other checks

- **WHEN** `KibanaResourceOptions.Update` is nil and `Update` is invoked
- **THEN** an error diagnostic describing the nil callback configuration SHALL be appended
- **AND** no other Update prelude checks (resourceID, client resolution) SHALL produce diagnostics

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

### Requirement: KibanaPostReadFunc receives Prior and State models and returns the model to commit to state
The system SHALL define `KibanaPostReadRequest[T]` as a struct with fields `Client *clients.KibanaScopedClient`, `Prior T`, `State T`, and `Private any`. The `KibanaPostReadFunc[T]` type SHALL be `func(ctx context.Context, req KibanaPostReadRequest[T]) (T, diag.Diagnostics)`. When the `PostRead` option is set, the envelope SHALL invoke PostRead after the read callback completes, pass the result of the read callback in `req.State`, and commit the model returned by PostRead to state. If PostRead returns error diagnostics, the envelope SHALL NOT set state.

On the write path (Create/Update), `req.Prior` SHALL be the plan model from the write request. On the plain Read path, `req.Prior` SHALL be the state model that existed before this refresh (the model decoded from the incoming Terraform state before the read callback is invoked).

#### Scenario: PostRead on write path receives plan as Prior and read-callback result as State
- **WHEN** `KibanaResourceOptions.PostRead` is set and the envelope completes a Create or Update operation
- **THEN** PostRead SHALL be invoked with `req.Prior` equal to the plan model and `req.State` equal to the model returned by the read callback
- **AND** the model returned by PostRead SHALL be passed to `resp.State.Set`
- **AND** `resp.State.Set` SHALL NOT be called with the read-callback model directly

#### Scenario: PostRead on plain Read path receives prior state as Prior and refreshed model as State
- **WHEN** `KibanaResourceOptions.PostRead` is set and the envelope executes a plain Read operation
- **THEN** PostRead SHALL be invoked with `req.Prior` equal to the state model decoded from Terraform state before the read callback ran and `req.State` equal to the model returned by the read callback
- **AND** the model returned by PostRead SHALL be passed to `resp.State.Set`

#### Scenario: PostRead error diagnostics prevent state set
- **WHEN** PostRead returns diagnostics containing at least one error
- **THEN** the envelope SHALL append those diagnostics and SHALL NOT call `resp.State.Set`

#### Scenario: PostRead not set — state set directly from read callback result
- **WHEN** `KibanaResourceOptions.PostRead` is nil
- **THEN** the model returned by the read callback SHALL be passed directly to `resp.State.Set`
- **AND** no PostRead invocation SHALL occur

### Requirement: Envelope auto-injects `timeouts` attribute into resource schema
The system SHALL inject a `timeouts` attribute produced by `terraform-plugin-framework-timeouts/resource/timeouts.AttributesAll(ctx)` into every Kibana envelope resource schema. The system SHALL copy the factory-supplied schema attributes first and then assign the envelope's `timeouts` attribute, silently overwriting any factory-supplied attribute with the same key. Concrete resource schema factories MUST NOT include a `timeouts` attribute in their output.

#### Scenario: Schema gains `timeouts` attribute with create/read/update/delete sub-attributes
- **WHEN** a Plugin Framework client requests the schema of any Kibana envelope resource
- **THEN** the response schema SHALL contain an attribute named `timeouts` of the `timeouts.AttributesAll` shape
- **AND** the `timeouts` attribute SHALL expose optional `create`, `read`, `update`, and `delete` string sub-attributes

#### Scenario: Factory-supplied `timeouts` attribute is silently replaced
- **WHEN** a concrete resource schema factory returns a schema whose `Attributes` map already contains a `timeouts` key
- **THEN** the envelope SHALL replace the factory's value with `timeouts.AttributesAll(ctx)`
- **AND** the envelope SHALL NOT panic or surface a diagnostic for the collision

### Requirement: Envelope wraps operation context with the configured timeout
The system SHALL wrap the context for every Create, Read, Update, and Delete operation with `context.WithTimeout` after the model is decoded from `req.Plan`/`req.State`/`req.Config` and **before** scoped client resolution, version-requirement enforcement, space-ID validation, the resource callback, read-after-write, and post-read run. All API-touching steps SHALL execute under the resolved deadline. The system SHALL defer cancellation so the wrapped context is released when the operation returns. The system SHALL append diagnostics returned by `timeouts.Value.Create/Read/Update/Delete` and SHALL NOT continue past the wrap if those diagnostics contain an error.

The system SHALL accept a `Timeouts ResourceTimeouts` field on `KibanaResourceOptions[T]`. For each operation the system SHALL use the corresponding `ResourceTimeouts` field as the default if non-zero, falling back to the package-level constants `DefaultResourceCreateTimeout`, `DefaultResourceReadTimeout`, `DefaultResourceUpdateTimeout`, and `DefaultResourceDeleteTimeout` defined alongside the Elasticsearch envelope.

#### Scenario: Create wraps context using plan `timeouts.create` or per-op default
- **WHEN** the envelope dispatches a Create operation
- **THEN** the envelope SHALL call `planModel.GetTimeouts().Create(ctx, defaultCreate)` where `defaultCreate` is `opts.Timeouts.Create` if non-zero, otherwise `DefaultResourceCreateTimeout`
- **AND** the envelope SHALL wrap the operation context with `context.WithTimeout` using the resolved duration
- **AND** the envelope SHALL defer `cancel()` so the context is released when Create returns

#### Scenario: Read, Update, and Delete apply per-operation timeouts
- **WHEN** the envelope dispatches a Read, Update, or Delete operation
- **THEN** the envelope SHALL apply the same ctx-wrap pattern as Create using the corresponding `timeouts.Value` method and the matching default constant
- **AND** Update SHALL derive the timeout from the plan model
- **AND** Read and Delete SHALL derive the timeout from the state model

#### Scenario: Version-requirement enforcement and space-ID validation run under the operation timeout
- **WHEN** the envelope dispatches any of Create, Read, Update, or Delete
- **THEN** `EnforceVersionRequirements`, space-ID validation, and any other API-issuing envelope plumbing SHALL execute with the wrapped (timeout-bound) context
- **AND** a slow `EnforceVersionRequirements` call SHALL be subject to the same deadline as the resource callback

#### Scenario: Null or unknown stored `timeouts` falls back to the default
- **WHEN** a Read, Update, or Delete operation runs against a state whose `timeouts` value is null, unknown, or has no entry for the current operation (for example after upgrading from a provider version that did not expose `timeouts`)
- **THEN** `model.GetTimeouts().<Op>(ctx, default)` SHALL return the envelope-supplied default with no diagnostics
- **AND** the operation SHALL proceed under that default-bound context

#### Scenario: Configured per-op default overrides the envelope default
- **WHEN** `KibanaResourceOptions.Timeouts.Create` is set to a non-zero duration
- **AND** the practitioner did not specify `timeouts.create` in configuration
- **THEN** the envelope SHALL use the configured per-op default in place of `DefaultResourceCreateTimeout`

#### Scenario: Practitioner-supplied timeout overrides per-op default
- **WHEN** the practitioner specifies `timeouts = { create = "30m" }` in configuration
- **THEN** the envelope SHALL wrap Create with `30m` regardless of `Options.Timeouts.Create` or `DefaultResourceCreateTimeout`

#### Scenario: Timeouts diagnostic prevents callback invocation
- **WHEN** `model.GetTimeouts().Create(ctx, default)` returns diagnostics containing at least one error
- **THEN** the envelope SHALL append those diagnostics
- **AND** the envelope SHALL NOT invoke the Create callback

#### Scenario: Envelope persists `timeouts` to state independent of callback-returned model
- **WHEN** a Read operation completes and the resource `readFunc` returns a model whose `timeouts` field is a zero value (for example because the callback reconstructed the model without copying `ResourceTimeoutsField`)
- **THEN** the envelope SHALL write the prior-state `timeouts` value back into `resp.State` after `resp.State.Set`
- **AND** the operation SHALL succeed without a `timeouts` value-conversion diagnostic
- **WHEN** a Create or Update operation completes and read-after-write returns a model whose `timeouts` field is a zero value
- **THEN** the envelope SHALL write the plan model's `timeouts` value into `outState` after `outState.Set`
- **AND** the operation SHALL succeed without a `timeouts` value-conversion diagnostic

### Requirement: Resource models embed `WithResourceTimeouts`
The system SHALL require `KibanaResourceModel` to embed `WithResourceTimeouts`. Concrete Kibana resource models satisfy the constraint by embedding `ResourceTimeoutsField`. The `WithResourceTimeouts` interface, `ResourceTimeoutsField` struct, and per-op default constants are shared with the Elasticsearch envelope and defined once in the entitycore package.

#### Scenario: Type constraint enforces `WithResourceTimeouts`
- **WHEN** a developer attempts to instantiate `KibanaResource[T]` with a model type that does not satisfy `WithResourceTimeouts`
- **THEN** the Go compiler SHALL reject the instantiation with a type-constraint violation
### Requirement: Fleet integration resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/fleet/integration` so that its `integrationResource` struct embeds `*entitycore.KibanaResource[integrationModel]` returned by `NewKibanaResource`. The `integrationModel` SHALL implement `KibanaResourceModel` with `GetResourceID()` returning the package identifier derived from `name`/`version` and `GetSpaceID()` returning `m.SpaceID`, and SHALL implement `KibanaUnscopedSpace` with `IsUnscopedSpace()` returning `true` when `SpaceID` is null or unknown (so the envelope skips space-ID validation in that case). The resource's Terraform schema attributes, CRUD semantics, state upgrade behaviour, and externally observable behaviour SHALL remain functionally equivalent after migration.

#### Scenario: Fleet integration resource satisfies envelope interface after migration

- **WHEN** `NewResource()` is called for the fleet integration resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope
- **AND** `ResourceWithUpgradeState` SHALL continue to be satisfied by the concrete resource type

#### Scenario: Fleet integration unscoped space skips envelope space validation

- **WHEN** an `integrationModel` value has `SpaceID` null or unknown and the envelope evaluates `KibanaUnscopedSpace`
- **THEN** `IsUnscopedSpace()` SHALL return `true`
- **AND** the envelope SHALL NOT emit an "Invalid space identifier" diagnostic for that write

---

### Requirement: Kibana synthetics monitor resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/kibana/synthetics/monitor` so that its `Resource` struct embeds `*entitycore.KibanaResource[tfModelV0]` returned by `NewKibanaResource`. The `tfModelV0` SHALL implement `KibanaResourceModel` with `GetID()` returning the composite `<space_id>:<monitor_id>` stored in `m.ID`, `GetResourceID()` returning the monitor-UUID portion of that composite identifier, and `GetSpaceID()` returning `m.SpaceID`. The resource's Terraform schema attributes, CRUD semantics, import behaviour, and externally observable behaviour SHALL remain functionally equivalent after migration. The dead `synthetics.ESAPIClient` interface, its compile-time assertion in `monitor/resource.go`, the `synthetics.GetKibanaOAPIClient` helper, and the `GetClient()` method on the concrete resource SHALL be removed as part of the migration.

#### Scenario: Synthetics monitor resource satisfies envelope interface after migration

- **WHEN** `NewResource()` is called for the synthetics monitor resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope

#### Scenario: ESAPIClient interface and helpers are removed

- **WHEN** the synthetics package is compiled after migration
- **THEN** the identifiers `synthetics.ESAPIClient`, `synthetics.GetKibanaOAPIClient`, and the monitor resource's `GetClient` method SHALL NOT exist
- **AND** no compile-time assertion of the form `_ synthetics.ESAPIClient = newResource()` SHALL remain in `monitor/resource.go`

---

### Requirement: Kibana SLO resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/kibana/slo` so that its `Resource` struct embeds `*entitycore.KibanaResource[tfModel]` returned by `NewKibanaResource`. The `tfModel` SHALL implement `KibanaResourceModel` with `GetID()` returning the composite `<space_id>:<slo_id>` stored in `m.ID`, `GetResourceID()` returning the SLO-UUID portion of that composite identifier, and `GetSpaceID()` returning `m.SpaceID`. The SLO Create and Update write callbacks SHALL perform the enabled-reconcile sequence — intermediate read of the server's `enabled` state, conditional Enable/Disable API call when the plan's `enabled` value differs, and a follow-up read — entirely inside the `KibanaWriteFunc[tfModel]`, not in `PostRead`. The resource's Terraform schema attributes, CRUD semantics, config validators, state upgrade behaviour, and externally observable behaviour SHALL remain functionally equivalent after migration.

#### Scenario: SLO resource satisfies envelope interface after migration

- **WHEN** `NewResource()` is called for the SLO resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope
- **AND** `ResourceWithConfigValidators` and `ResourceWithUpgradeState` SHALL continue to be satisfied by the concrete resource type

#### Scenario: Enabled reconciliation runs inside the write callback

- **WHEN** the SLO write callback observes a plan `enabled` value different from the server's `enabled` value after the post-create or post-update intermediate read
- **THEN** the write callback SHALL invoke the Kibana Enable or Disable SLO API as appropriate
- **AND** the write callback SHALL re-read the SLO before returning its `KibanaWriteResult`
- **AND** `PostRead` SHALL NOT be involved in this reconciliation

### Requirement: Write callbacks may opt out of read-after-write via `KibanaWriteResult.SkipReadAfterWrite`

The system SHALL define a boolean field `SkipReadAfterWrite` on `KibanaWriteResult[T]` returned by Kibana envelope Create and Update callbacks. The field SHALL default to `false` when unset. When `SkipReadAfterWrite` is `false`, the envelope SHALL retain the existing write path: invoke the read callback after a successful write callback and commit the read result to state (and invoke PostRead when configured). When `SkipReadAfterWrite` is `true`, the envelope SHALL persist `written.Model` directly to state without invoking the read callback and without invoking PostRead (no read occurred on the write path).

In both paths the envelope SHALL continue to apply `preserveModelTimeouts` using the plan model's `timeouts` value and SHALL set the `timeouts` attribute on state after `Set`, matching existing timeout persistence behavior.

Concrete resources that skip read-after-write SHOULD merge known server-computed values from prior state into the returned model when the plan leaves those attributes Unknown, so direct state write does not persist Unknown computed fields.

#### Scenario: SkipReadAfterWrite false — read-after-write and PostRead run as today

- **WHEN** a Create or Update write callback returns `SkipReadAfterWrite: false` (or the zero value)
- **THEN** the envelope SHALL invoke the read callback after the write callback succeeds
- **AND** when `PostRead` is configured, the envelope SHALL invoke PostRead with the read callback result before committing state

#### Scenario: SkipReadAfterWrite true — no read callback and no PostRead

- **WHEN** an Update write callback returns `SkipReadAfterWrite: true` and a model to persist
- **THEN** the envelope SHALL NOT invoke the read callback
- **AND** the envelope SHALL NOT invoke PostRead
- **AND** the envelope SHALL commit `written.Model` to state

#### Scenario: SkipReadAfterWrite true — plan timeouts still persisted

- **WHEN** an Update write callback returns `SkipReadAfterWrite: true` and a model whose embedded `timeouts` field is a zero value
- **THEN** the envelope SHALL write the plan model's `timeouts` value into state after `Set`
- **AND** the operation SHALL succeed without a `timeouts` value-conversion diagnostic

