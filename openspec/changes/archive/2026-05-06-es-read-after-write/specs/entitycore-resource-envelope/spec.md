## MODIFIED Requirements

### Requirement: Envelope owns the Create and Update preludes

The system SHALL implement `Create` and `Update` on `NewElasticsearchResource[T]` by deserializing the planned model into `T`, deriving the write resource ID from the model, resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, invoking the corresponding concrete callback, and then invoking `readFunc` with the model returned by the callback. The concrete create and update callbacks SHALL be invoked with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)`. After a successful callback invocation, `readFunc` SHALL be invoked with `(context, *clients.ElasticsearchScopedClient, resourceID string, writtenModel)` where `writtenModel` is the model returned by the concrete callback. State SHALL be set from the model returned by `readFunc`, not directly from the concrete callback.

#### Scenario: Successful create sets state from read result

- **WHEN** the concrete create function returns `(writtenModel, nil)` and `readFunc` returns `(stateModel, true, nil)`
- **THEN** `resp.State.Set` SHALL be called with `stateModel` (the model returned by `readFunc`)
- **AND** response diagnostics SHALL contain no errors

#### Scenario: Successful update sets state from read result

- **WHEN** the concrete update function returns `(writtenModel, nil)` and `readFunc` returns `(stateModel, true, nil)`
- **THEN** `resp.State.Set` SHALL be called with `stateModel` (the model returned by `readFunc`)
- **AND** response diagnostics SHALL contain no errors

#### Scenario: Resource not found after create produces error

- **WHEN** the concrete create function returns `(writtenModel, nil)` and `readFunc` returns `(_, false, nil)`
- **THEN** an error diagnostic SHALL be appended to `resp.Diagnostics` identifying the resource type using the envelope's component and resource name
- **AND** state SHALL remain untouched (neither `resp.State.Set` nor `resp.State.RemoveResource` SHALL be called)

#### Scenario: Resource not found after update produces error

- **WHEN** the concrete update function returns `(writtenModel, nil)` and `readFunc` returns `(_, false, nil)`
- **THEN** an error diagnostic SHALL be appended to `resp.Diagnostics` identifying the resource type using the envelope's component and resource name
- **AND** state SHALL remain untouched (neither `resp.State.Set` nor `resp.State.RemoveResource` SHALL be called)

#### Scenario: readFunc error after create short-circuits state mutation

- **WHEN** the concrete create function returns `(writtenModel, nil)` and `readFunc` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: readFunc error after update short-circuits state mutation

- **WHEN** the concrete update function returns `(writtenModel, nil)` and `readFunc` returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: Create function error short-circuits state mutation

- **WHEN** the concrete create function returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `readFunc` SHALL NOT be invoked
- **AND** `resp.State.Set` SHALL NOT be called

#### Scenario: Update function error short-circuits state mutation

- **WHEN** the concrete update function returns error diagnostics
- **THEN** the diagnostics SHALL be appended to `resp.Diagnostics`
- **AND** `readFunc` SHALL NOT be invoked
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
