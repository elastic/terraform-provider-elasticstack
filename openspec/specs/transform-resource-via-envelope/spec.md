# transform-resource-via-envelope Specification

## Purpose
TBD - created by archiving change es-entitycore-migration-phase6-transform. Update Purpose after archive.
## Requirements
### Requirement: Transform resource uses the entitycore envelope
The `elasticstack_elasticsearch_transform` resource SHALL be implemented on PF and embed `*entitycore.ElasticsearchResource[tfModel]`.

#### Scenario: Resource is registered as a PF resource
- **WHEN** the provider initializes
- **THEN** the resource SHALL be registered via the envelope

### Requirement: Schema validation enforces pivot/latest mutual exclusivity
The PF schema SHALL enforce that exactly one of `pivot` and `latest` is set using `ExactlyOneOf`.

#### Scenario: Both pivot and latest set
- **GIVEN** a configuration with both `pivot` and `latest`
- **WHEN** validation runs
- **THEN** the provider SHALL return a validation error

#### Scenario: Neither pivot nor latest set
- **GIVEN** a configuration with neither `pivot` nor `latest`
- **WHEN** validation runs
- **THEN** the provider SHALL return a validation error

### Requirement: Model conversion preserves version-gated field behavior
The model-to-API conversion SHALL omit version-gated fields from API requests when the target cluster version is below the field's minimum supported version, preserving the existing warning behavior.

#### Scenario: destination.aliases omitted on old cluster
- **GIVEN** `destination.aliases` is set and cluster version < 8.8.0
- **WHEN** the API request body is built
- **THEN** the request body SHALL omit `destination.aliases`
- **AND** a warning SHALL be logged

#### Scenario: deduce_mappings omitted on old cluster
- **GIVEN** `settings.deduce_mappings = true` and version < 8.1.0
- **WHEN** the API request body is built
- **THEN** the request body SHALL omit `deduce_mappings`
- **AND** a warning SHALL be logged

### Requirement: Create callback puts and optionally starts the transform
The `createFunc` callback SHALL PUT the transform definition and start it when `enabled = true`. The callback SHALL pass `defer_validation` to the Put Transform API without using it to decide whether the transform starts.

#### Scenario: Create and start transform
- **GIVEN** `enabled = true`
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the transform
- **AND** start it
- **AND** the envelope SHALL read back and persist state

#### Scenario: Create without starting transform
- **GIVEN** `enabled = false`
- **WHEN** the create callback runs
- **THEN** it SHALL PUT the transform
- **AND** NOT start it

### Requirement: Update override updates configuration and reconciles enabled state
The concrete resource SHALL override `Update` because transform updates require comparing the prior state with the planned state. The override SHALL call the Update Transform API, omit immutable `pivot` and `latest` fields from the request body, and start or stop the transform only when `enabled` changes.

#### Scenario: Update enables transform
- **GIVEN** prior state has `enabled = false`
- **AND** the planned state has `enabled = true`
- **WHEN** the concrete Update override runs
- **THEN** it SHALL call the Update Transform API
- **AND** start the transform
- **AND** read back and persist refreshed state

#### Scenario: Update disables transform
- **GIVEN** prior state has `enabled = true`
- **AND** the planned state has `enabled = false`
- **WHEN** the concrete Update override runs
- **THEN** it SHALL call the Update Transform API
- **AND** stop the transform
- **AND** read back and persist refreshed state

### Requirement: Read callback fetches transform config
The `readFunc` callback SHALL GET the transform config and populate the model.

#### Scenario: Successful read
- **GIVEN** a transform exists
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == true` with fields populated

#### Scenario: Missing transform removes from state
- **GIVEN** the transform does not exist
- **WHEN** the read callback runs
- **THEN** it SHALL return `found == false`

### Requirement: Delete callback deletes the transform with force
The `deleteFunc` callback SHALL DELETE the transform with `force=true`.

#### Scenario: Delete transform
- **GIVEN** a running transform
- **WHEN** the delete callback runs
- **THEN** it SHALL call Delete Transform with `force=true`

### Requirement: Import behavior is preserved
The concrete resource SHALL implement `ImportState` as passthrough.

#### Scenario: Import transform
- **GIVEN** a composite import id
- **WHEN** import runs
- **THEN** the `id` SHALL be stored in state

