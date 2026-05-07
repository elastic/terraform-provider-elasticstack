## ADDED Requirements

### Requirement: Model satisfies entitycore resource envelope constraint

The `tfModel` type SHALL declare value-receiver methods `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List` so that it satisfies `entitycore.ElasticsearchResourceModel`.

#### Scenario: Model usable as envelope type parameter

- **WHEN** `entitycore.NewElasticsearchResource[tfModel]` is constructed
- **THEN** compilation SHALL succeed because `tfModel` implements the required interface

### Requirement: Read callback is package-level and shared with override

The resource SHALL provide a package-level `readAPIKey` function with signature `func(context.Context, *clients.ElasticsearchScopedClient, string, tfModel) (tfModel, bool, diag.Diagnostics)`. This function SHALL be passed as the read callback to `NewElasticsearchResource` and SHALL also be callable from the concrete type's `Read` override.

#### Scenario: Read override delegates API interaction to the callback

- **WHEN** the concrete `Read` method runs after a successful state decode and client resolution
- **THEN** it SHALL invoke `readAPIKey` to perform the actual Elasticsearch API interaction
- **AND** it SHALL not duplicate the API call logic

### Requirement: Delete callback is package-level

The resource SHALL provide a package-level `deleteAPIKey` function with signature `func(context.Context, *clients.ElasticsearchScopedClient, string, tfModel) diag.Diagnostics`. This function SHALL be passed as the delete callback to `NewElasticsearchResource`.

#### Scenario: Envelope delete invokes the callback

- **WHEN** the envelope's `Delete` runs for an api_key resource
- **THEN** it SHALL invoke `deleteAPIKey` after resolving the scoped client

### Requirement: Create and Update use placeholder callbacks with concrete overrides

The resource SHALL pass `PlaceholderElasticsearchWriteCallbacks[tfModel]()` for create and update to `NewElasticsearchResource`. The concrete `Resource` type SHALL define its own `Create` and `Update` methods that shadow the envelope's, preserving the existing private-state, version-gating, and cross-cluster API key flows.

#### Scenario: Create is handled by the concrete type

- **WHEN** Terraform calls Create on the api_key resource
- **THEN** the concrete `Resource.Create` SHALL run, not the envelope's
- **AND** the private-state write and version-checking logic SHALL remain unchanged

#### Scenario: Update is handled by the concrete type

- **WHEN** Terraform calls Update on the api_key resource
- **THEN** the concrete `Resource.Update` SHALL run, not the envelope's
- **AND** the private-state read and version-gating logic SHALL remain unchanged

### Requirement: Read is overridden to write private cluster version

The concrete `Resource` type SHALL define its own `Read` method. After a successful read that finds the API key, the override SHALL call `saveClusterVersion` to persist the current server version in private state. If the API key is not found, the override SHALL remove the resource from state and SHALL NOT write private state.

#### Scenario: Successful read saves cluster version

- **WHEN** Read finds the API key in Elasticsearch
- **THEN** state SHALL be persisted via `resp.State.Set`
- **AND** the server version SHALL be saved to private state via `saveClusterVersion`

#### Scenario: Not-found read removes resource without saving version

- **WHEN** Read does not find the API key
- **THEN** `resp.State.RemoveResource` SHALL be called
- **AND** private state SHALL NOT be written

### Requirement: Schema factory omits connection block

The schema factory SHALL return a `schema.Schema` whose `Blocks` map does not contain `elasticsearch_connection`. The envelope SHALL inject the connection block before exposing the schema.

#### Scenario: Schema exposed through envelope includes connection block

- **WHEN** the envelope's `Schema` method is called
- **THEN** the returned schema SHALL include the `elasticsearch_connection` block
- **AND** the concrete factory SHALL not declare it

### Requirement: Version-checking plan modifiers remain functional

The `requiresReplaceIfUpdateNotSupported` plan modifier SHALL continue to read the cached cluster version from private state and require replacement when the cached version is lower than `8.4.0`. The `saveClusterVersion` helper SHALL remain callable from the concrete `Read` override.

#### Scenario: Plan modifier reads cached version after migration

- **GIVEN** a successful Read has saved a cluster version to private state
- **WHEN** a plan is computed that changes `metadata` or `role_descriptors`
- **THEN** the plan modifier SHALL read the cached version from private state
- **AND** it SHALL require replacement if the cached version is below `8.4.0`

### Requirement: State upgrades remain on the concrete type

The concrete `Resource` type SHALL continue to implement `ResourceWithUpgradeState`, providing state upgraders from schema version `0` to `1` and from `1` to `2`.

#### Scenario: State upgrade registration

- **WHEN** the concrete resource type is asserted as `resource.ResourceWithUpgradeState`
- **THEN** the assertion SHALL succeed
- **AND** the upgrade map SHALL contain the two defined upgraders

## MODIFIED Requirements

None — externally-observable behavior is preserved verbatim.

## REMOVED Requirements

None.
