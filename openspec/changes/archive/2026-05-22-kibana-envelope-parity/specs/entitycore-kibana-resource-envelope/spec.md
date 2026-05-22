## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL provide a generic constructor `NewKibanaResource[T]()` that accepts a `KibanaResourceOptions[T]` options struct (not a positional callback list) and returns an envelope owning shared Kibana resource behavior. The envelope SHALL provide Metadata, Schema, Configure, Create, Read, Update, and Delete behavior, and SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewKibanaResource[T](component, name, opts)` is called with a `KibanaResourceOptions[T]` containing non-nil required callbacks
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the returned value SHALL NOT satisfy `resource.ResourceWithImportState`

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewKibanaResource[T](ComponentKibana, "maintenance_window", opts)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_kibana_maintenance_window`

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

## NEW Requirements

### Requirement: Write callbacks use unified KibanaWriteFunc type

The system SHALL define a single `KibanaWriteFunc[T KibanaResourceModel]` type for both Create and Update callbacks. Write callbacks SHALL receive a `KibanaWriteRequest[T]` struct and return a `KibanaWriteResult[T]` and diagnostics. Separate `KibanaCreateFunc[T]` and `KibanaUpdateFunc[T]` types SHALL be removed.

`KibanaWriteRequest[T]` SHALL carry:
- `Plan T`: the decoded Terraform plan
- `Prior *T`: nil for Create invocations; a pointer to the decoded prior state model for Update invocations
- `Config T`: the Terraform configuration decoded into T by the envelope
- `WriteID string`: the string value returned by `plan.GetResourceID().ValueString()`
- `SpaceID string`: the string value returned by `plan.GetSpaceID().ValueString()`

`KibanaWriteResult[T]` SHALL carry:
- `Model T`: the model returned from the write operation, used for read-after-write identity resolution

#### Scenario: Create callback receives nil Prior

- **WHEN** `Create` runs
- **THEN** the callback SHALL receive `KibanaWriteRequest[T]` with `Prior == nil`
- **AND** `Plan` SHALL contain the decoded plan model
- **AND** `Config` SHALL contain the Terraform configuration decoded into T
- **AND** `WriteID` SHALL be `plan.GetResourceID().ValueString()`
- **AND** `SpaceID` SHALL be `plan.GetSpaceID().ValueString()`

#### Scenario: Update callback receives non-nil Prior

- **WHEN** `Update` runs
- **THEN** the callback SHALL receive `KibanaWriteRequest[T]` with `Prior` pointing at the decoded prior state model
- **AND** `Plan` SHALL contain the decoded plan model
- **AND** `Config` SHALL contain the Terraform configuration decoded into T

#### Scenario: Single KibanaWriteFunc serves both Create and Update

- **WHEN** the same `KibanaWriteFunc[T]` is wired into both `KibanaResourceOptions.Create` and `KibanaResourceOptions.Update`
- **THEN** invocations from `Create` SHALL have `req.Prior == nil`
- **AND** invocations from `Update` SHALL have `req.Prior != nil`
- **AND** the function MAY distinguish the two paths by inspecting `req.Prior`

---

### Requirement: Envelope owns Create and Update with enforced read-after-write

The system SHALL implement `Create` and `Update` by deserializing the plan and config, validating spaceID, resolving the scoped Kibana client, enforcing version requirements, invoking the write callback with `KibanaWriteRequest[T]`, then calling `readFunc` with the identity resolved from the written model, and persisting the *read* model — not the write result — into state.

After a successful write, read identity SHALL be resolved as:
- `readResourceID = writtenModel.GetResourceID().ValueString()`
- `readSpaceID = writtenModel.GetSpaceID().ValueString()`

If `readFunc` returns `found == false`, the envelope SHALL append an error diagnostic with summary "Resource not found" and SHALL NOT persist state. If `readFunc` returns error diagnostics, those SHALL be appended and state SHALL NOT be mutated.

The optional `PostRead` hook SHALL be invoked after a successful read-after-write that persists state, in addition to after standalone `Read` operations.

#### Scenario: Create read-after-write persists read model not write model

- **WHEN** the create callback returns a model and `readFunc` returns a different model with `found == true`
- **THEN** `resp.State.Set` SHALL be called with the model returned by `readFunc`
- **AND** the model returned by the create callback SHALL NOT be persisted directly

#### Scenario: Update read-after-write persists read model

- **WHEN** the update callback returns a model and `readFunc` returns a different model with `found == true`
- **THEN** `resp.State.Set` SHALL be called with the model returned by `readFunc`
- **AND** the model returned by the update callback SHALL NOT be persisted directly

#### Scenario: Not found after Create returns error

- **WHEN** the create callback returns successfully but `readFunc` returns `found == false`
- **THEN** an error diagnostic with summary "Resource not found" SHALL be appended
- **AND** state SHALL NOT be mutated

#### Scenario: Not found after Update returns error

- **WHEN** the update callback returns successfully but `readFunc` returns `found == false`
- **THEN** an error diagnostic with summary "Resource not found" SHALL be appended
- **AND** state SHALL NOT be mutated

#### Scenario: readFunc error after write short-circuits state persistence

- **WHEN** the write callback returns successfully but `readFunc` returns error diagnostics
- **THEN** those diagnostics SHALL be appended
- **AND** state SHALL NOT be mutated

#### Scenario: PostRead invoked after Create read-after-write

- **WHEN** a `PostRead` hook is configured and Create completes a successful read-after-write
- **THEN** the hook SHALL be invoked with the read model and the framework private-state handle

#### Scenario: PostRead invoked after Update read-after-write

- **WHEN** a `PostRead` hook is configured and Update completes a successful read-after-write
- **THEN** the hook SHALL be invoked with the read model and the framework private-state handle

---

### Requirement: Envelope supports post-read side effects

The system SHALL allow Kibana envelope users to provide an optional `PostRead KibanaPostReadFunc[T]` hook via `KibanaResourceOptions[T]`. After a successful read flow that sets Terraform state (including read-after-write), the envelope SHALL invoke the hook with the scoped client, the model persisted to state, and the framework private-state handle.

The hook SHALL NOT run when the entity is not found, when `readFunc` returns error diagnostics, or when state persistence fails.

#### Scenario: PostRead hook invoked after successful standalone Read

- **WHEN** `readFunc` returns `(model, true, nil)` and `resp.State.Set` succeeds
- **THEN** the envelope SHALL invoke the configured PostRead hook with the persisted model and `resp.Private`

#### Scenario: PostRead skipped when not found

- **WHEN** `readFunc` returns `found == false`
- **THEN** the PostRead hook SHALL NOT be invoked

#### Scenario: PostRead skipped when readFunc returns errors

- **WHEN** `readFunc` returns error diagnostics
- **THEN** the PostRead hook SHALL NOT be invoked

#### Scenario: PostRead skipped when state set fails

- **WHEN** `resp.State.Set` returns error diagnostics
- **THEN** the PostRead hook SHALL NOT be invoked

#### Scenario: PostRead receives framework private handle

- **WHEN** the PostRead hook is invoked
- **THEN** the `privateState` argument SHALL be the same object as `resp.Private` from the framework response

---

### Requirement: PlaceholderKibanaWriteCallback returns a single callback that errors when invoked

The system SHALL provide `PlaceholderKibanaWriteCallback[T]()` that returns a single `KibanaWriteFunc[T]`. The placeholder SHALL add a predictable error diagnostic if invoked. This function exists for concrete resources that override Create and/or Update on the envelope struct and therefore never intend the envelope's callbacks to be called. The previous two-return-value form `PlaceholderKibanaWriteCallbacks[T]()` SHALL be removed.

#### Scenario: Placeholder callback returns error diagnostic

- **WHEN** the placeholder `KibanaWriteFunc[T]` is invoked (for any Create or Update path)
- **THEN** it SHALL return a non-empty `diag.Diagnostics` with an error-level diagnostic describing the misconfiguration

---

## REMOVED Requirements

### Requirement: Envelope owns the Create prelude

**Reason**: Superseded by the updated "Envelope owns Create and Update with enforced read-after-write" requirement. The old Create prelude description referenced `KibanaCreateFunc[T]` positional arguments and no read-after-write, both of which are replaced.
**Migration**: Use `KibanaWriteFunc[T]` via `KibanaResourceOptions.Create`. Access `spaceID` via `req.SpaceID`, plan via `req.Plan`.

### Requirement: Envelope owns the Update prelude and passes both plan and prior state

**Reason**: Superseded by the updated "Envelope owns Create and Update with enforced read-after-write" requirement. The old Update prelude described `KibanaUpdateFunc[T]` with positional `(resourceID, spaceID, plan, prior)` arguments and no read-after-write.
**Migration**: Use `KibanaWriteFunc[T]` via `KibanaResourceOptions.Update`. Access `resourceID` via `req.WriteID`, `spaceID` via `req.SpaceID`, plan via `req.Plan`, prior via `req.Prior`.

### Requirement: PlaceholderKibanaWriteCallbacks returns callbacks that error when invoked

**Reason**: Replaced by the singular `PlaceholderKibanaWriteCallback[T]()` returning a single `KibanaWriteFunc[T]`.
**Migration**: Replace `createFn, updateFn := entitycore.PlaceholderKibanaWriteCallbacks[T]()` with `placeholder := entitycore.PlaceholderKibanaWriteCallback[T]()` and supply `placeholder` to both `KibanaResourceOptions.Create` and `KibanaResourceOptions.Update`.
