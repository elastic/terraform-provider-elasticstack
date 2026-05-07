## ADDED Requirements

### Requirement: Index lifecycle resource uses the entitycore envelope
The `elasticstack_elasticsearch_index_lifecycle` resource SHALL embed `*entitycore.ElasticsearchResource[tfModel]`. The resource SHALL satisfy `resource.Resource`, `resource.ResourceWithConfigure`, `resource.ResourceWithImportState`, and `resource.ResourceWithUpgradeState`.

#### Scenario: Resource is registered as an envelope resource
- **WHEN** the provider initializes
- **THEN** `elasticstack_elasticsearch_index_lifecycle` SHALL be constructed via `entitycore.NewElasticsearchResource[tfModel]` with real create, update, read, and delete callbacks
- **AND** the concrete type SHALL declare `UpgradeState` and `ImportState`

#### Scenario: Schema includes injected connection block
- **WHEN** the schema factory returns a schema without `elasticsearch_connection`
- **THEN** the envelope's `Schema` method SHALL inject the connection block
- **AND** all phase blocks and attributes SHALL match the pre-migration schema exactly

### Requirement: Model satisfies ElasticsearchResourceModel
The resource's `tfModel` struct SHALL implement `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`.

#### Scenario: Model getters return correct fields
- **WHEN** a `tfModel` value is created with `ID`, `Name`, and `ElasticsearchConnection` populated
- **THEN** `GetID()` SHALL return the `ID` field
- **AND** `GetResourceID()` SHALL return the `Name` field
- **AND** `GetElasticsearchConnection()` SHALL return the `ElasticsearchConnection` field

### Requirement: Read callback fetches and maps ILM policies
The envelope's read callback SHALL call the Get Lifecycle API, return `(model, true, nil)` when the policy exists, and return `(_, false, nil)` when the policy is not found (404 or missing from response).

#### Scenario: Successful read returns populated model
- **GIVEN** an ILM policy exists in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return a populated `tfModel` with all configured phases, actions, and computed `modified_date`

#### Scenario: Missing policy removes from state
- **GIVEN** the ILM policy does not exist in Elasticsearch
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`
- **AND** the envelope SHALL remove the resource from state

### Requirement: Create and update callbacks use Put Lifecycle API
The `createFunc` and `updateFunc` callbacks SHALL expand the planned model into a `models.Policy`, validate version-gated settings against the server version, call the Put Lifecycle API, compute the composite `id`, and return the written model. The envelope SHALL invoke `readFunc` after a successful callback and persist the read result.

#### Scenario: Create puts policy and refreshes state
- **GIVEN** a valid planned `tfModel`
- **WHEN** the create callback runs
- **THEN** it SHALL call Put Lifecycle API
- **AND** the envelope SHALL read back the policy and set state from the read result

#### Scenario: Update rejects unsupported settings on old clusters
- **GIVEN** a planned model with `rollover.min_docs` set and Elasticsearch < 8.4.0
- **WHEN** the update callback runs
- **THEN** it SHALL return an error diagnostic without calling the Put API

### Requirement: Delete callback removes ILM policies
The envelope's delete callback SHALL call the Delete Lifecycle API with the parsed resource identifier.

#### Scenario: Successful delete
- **GIVEN** an ILM policy exists
- **WHEN** the delete callback runs
- **THEN** it SHALL call Delete Lifecycle API
- **AND** return nil diagnostics on success

### Requirement: State upgrader is preserved
The concrete type SHALL continue to register the V0→V1 state upgrader, unwrapping legacy singleton-list phase and action values into object values.

#### Scenario: Upgrade old SDK-shaped nested values
- **GIVEN** persisted schema version 0 state with a phase stored as `[ { ... } ]`
- **WHEN** Terraform runs the state upgrader
- **THEN** the upgraded state SHALL store that phase as a single object value

### Requirement: Disabled toggle preservation across refresh is maintained
On read, when the API omits `readonly`, `freeze`, or `unfollow` actions but the prior state had declared the block, the read callback SHALL preserve that declaration with `enabled = false`.

#### Scenario: Disabled unfollow remains disabled after refresh
- **GIVEN** prior Terraform state declared `unfollow { enabled = false }`
- **WHEN** refresh reads a phase whose API actions omit `unfollow`
- **THEN** the returned model SHALL contain `unfollow.enabled = false`
