## ADDED Requirements

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
