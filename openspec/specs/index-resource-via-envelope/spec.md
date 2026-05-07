# index-resource-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase5-index. Update Purpose after archive.
## Requirements
### Requirement: Index resource uses the entitycore envelope for Schema, Read, and Delete
The `elasticstack_elasticsearch_index` resource SHALL embed `*entitycore.ElasticsearchResource[tfModel]`. The envelope SHALL own Schema, Read, and Delete.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Model satisfies ElasticsearchResourceModel
The `tfModel` struct SHALL implement `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()`.

#### Scenario: Model getters return correct fields
- **WHEN** a model is created with `ID`, `Name`, and `ElasticsearchConnection`
- **THEN** `GetID()` SHALL return `ID`
- **AND** `GetResourceID()` SHALL return `Name`
- **AND** `GetElasticsearchConnection()` SHALL return the connection list

### Requirement: Create override supports use_existing adoption
The concrete resource SHALL override `Create` to check for an existing index when `use_existing = true`.

#### Scenario: Adopt existing compatible index
- **GIVEN** `use_existing = true` and an index with matching static settings exists
- **WHEN** create runs
- **THEN** it SHALL adopt the existing index
- **AND** reconcile aliases, settings, and mappings
- **AND** add a warning about adoption

#### Scenario: Reject incompatible existing index
- **GIVEN** `use_existing = true` and an index with mismatched static settings
- **WHEN** create runs
- **THEN** it SHALL return an error with the mismatch details

#### Scenario: Ignore use_existing for date-math names
- **GIVEN** `use_existing = true` and a date-math name pattern
- **WHEN** create runs
- **THEN** it SHALL issue a warning and proceed with normal creation

### Requirement: Create override creates new index when use_existing is false
When `use_existing` is false or null, the create override SHALL PUT a new index.

#### Scenario: Create new index
- **GIVEN** `use_existing = false`
- **WHEN** create runs
- **THEN** it SHALL PUT a new index
- **AND** the concrete create override SHALL read back and persist state

### Requirement: Update override derives concrete name from state
The concrete resource SHALL override `Update` to derive the concrete index identity from current state, not from plan.

#### Scenario: Update with date-math resolved name
- **GIVEN** an existing state where the concrete name is `logs-2024.01.15`
- **WHEN** update runs
- **THEN** it SHALL target `logs-2024.01.15` for alias/settings/mappings updates

### Requirement: Update override reconciles aliases, settings, and mappings independently
The update override SHALL compare plan vs state for each aspect and issue partial updates.

#### Scenario: Update aliases only
- **GIVEN** an existing state and a plan with changed aliases only
- **WHEN** update runs
- **THEN** it SHALL only update aliases

#### Scenario: Update mappings only
- **GIVEN** an existing state and a plan with changed mappings only
- **WHEN** update runs
- **THEN** it SHALL only update mappings

### Requirement: Read callback fetches index info
The `readFunc` callback SHALL GET the index and populate the model.

#### Scenario: Successful read
- **GIVEN** an index exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with all fields populated

#### Scenario: Missing index removes from state
- **GIVEN** the index does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Import behavior is preserved
The concrete resource SHALL implement `ImportState` as passthrough on `id`.

#### Scenario: Import index
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state

