## ADDED Requirements

### Requirement: Index alias resource uses the entitycore envelope
The `elasticstack_elasticsearch_index_alias` resource SHALL embed `*entitycore.ElasticsearchResource[Data]`.

#### Scenario: Resource is registered as a Plugin Framework resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered as a PF resource
- **AND** it SHALL be constructed via `entitycore.NewElasticsearchResource[Data]`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's model SHALL implement `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` as value-receiver methods.

#### Scenario: Model getters return correct fields
- **WHEN** a model value is created with all required fields
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the alias write identifier
- **AND** `GetElasticsearchConnection()` SHALL return the connection field

### Requirement: Create callback updates the alias
The `createFunc` callback SHALL PUT the alias definition.

#### Scenario: Create alias
- **GIVEN** a valid planned model
- **WHEN** the create callback runs
- **THEN** it SHALL call the update alias API
- **AND** the envelope SHALL read back and persist state

### Requirement: Update callback replaces the alias
The `updateFunc` callback SHALL delete old aliases and PUT the new definition.

#### Scenario: Update alias
- **GIVEN** an existing alias with changed settings
- **WHEN** the update callback runs
- **THEN** it SHALL reconcile aliases and PUT
- **AND** the envelope SHALL read back and persist state

### Requirement: Read callback fetches aliases
The `readFunc` callback SHALL GET the index aliases and return whether the alias exists.

#### Scenario: Successful read
- **GIVEN** an alias exists on the index
- **WHEN** the read callback runs
- **THEN** it SHALL return a populated model with `found == true`

#### Scenario: Missing alias removes from state
- **GIVEN** the alias does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback removes the alias
The `deleteFunc` callback SHALL DELETE the alias.

#### Scenario: Successful delete
- **GIVEN** an alias exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call the delete alias API

### Requirement: Import behavior is preserved
The concrete resource SHALL implement `ImportState` as passthrough.

#### Scenario: Import alias
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state
