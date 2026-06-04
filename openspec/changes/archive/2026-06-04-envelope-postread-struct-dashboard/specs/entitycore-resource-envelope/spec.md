## MODIFIED Requirements

### Requirement: Envelope supports post-read side effects
The system SHALL allow Elasticsearch envelope users to provide an optional post-read hook. The system SHALL define `ElasticsearchPostReadRequest[T]` as a struct with fields `Client *clients.ElasticsearchScopedClient`, `Prior T`, `State T`, and `Private any`. The `PostReadFunc[T]` type SHALL be `func(ctx context.Context, req ElasticsearchPostReadRequest[T]) (T, diag.Diagnostics)`.

When `PostRead` is configured, the envelope SHALL invoke it after the read callback completes successfully, pass the read-callback result in `req.State`, populate `req.Prior` with the plan model (Create/Update) or the state model decoded before the read (plain Read), and commit the model returned by PostRead to state. If PostRead returns error diagnostics, the envelope SHALL append those diagnostics and SHALL NOT call `resp.State.Set`. The hook SHALL NOT run when the entity is not found or when `readFunc` returns error diagnostics.

#### Scenario: Post-read hook on write path receives plan as Prior and read-callback result as State
- **WHEN** `ElasticsearchResourceOptions.PostRead` is set and the envelope completes a Create or Update operation
- **THEN** PostRead SHALL be invoked with `req.Prior` equal to the plan model and `req.State` equal to the model returned by the read callback
- **AND** the model returned by PostRead SHALL be passed to `resp.State.Set`
- **AND** `resp.State.Set` SHALL NOT be called with the read-callback model directly

#### Scenario: Post-read hook on plain Read path receives prior state as Prior and refreshed model as State
- **WHEN** `ElasticsearchResourceOptions.PostRead` is set and the envelope executes a plain Read operation
- **THEN** PostRead SHALL be invoked with `req.Prior` equal to the state model decoded from Terraform state before the read callback ran and `req.State` equal to the model returned by the read callback
- **AND** the model returned by PostRead SHALL be passed to `resp.State.Set`

#### Scenario: Post-read error diagnostics prevent state set
- **WHEN** PostRead returns diagnostics containing at least one error
- **THEN** the envelope SHALL append those diagnostics and SHALL NOT call `resp.State.Set`

#### Scenario: Post-read hook not set — state set directly from read callback result
- **WHEN** `ElasticsearchResourceOptions.PostRead` is nil
- **THEN** the model returned by the read callback SHALL be passed directly to `resp.State.Set`
- **AND** no PostRead invocation SHALL occur
