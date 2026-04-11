## ADDED Requirements

### Requirement: Agent prompt uses post-hook memory scripts
The `schema-coverage-rotation` workflow SHALL instruct the agent to run repository-local schema-coverage memory script commands after repo-memory hooks complete, and SHALL NOT require the prompt to describe the memory JSON structure in detail.

#### Scenario: Repo memory is available to the agent
- **WHEN** the workflow starts the agent after built-in repo-memory hooks have prepared the workspace
- **THEN** the prompt tells the agent which schema-coverage memory script command or commands to run instead of describing the memory file format

### Requirement: Memory helper is implemented as a Go command
The schema-coverage memory helper SHALL be implemented as a Go command rooted at `scripts/schema-coverage-rotation` so the workflow can invoke it directly from the repository workspace.

#### Scenario: Workflow invokes the helper
- **WHEN** the agent needs to prepare memory, select entities, or record analysis completion
- **THEN** it runs the Go-based helper command from `scripts/schema-coverage-rotation`

### Requirement: Memory preparation bootstraps and reconciles the canonical entity inventory
The schema-coverage memory preparation command SHALL initialize the working memory file from `.github/aw/memory/schema-coverage.json` when the working file does not exist, SHALL derive the canonical entity inventory by importing the provider registrations exposed through `provider/plugin_framework.go` and `provider/provider.go`, SHALL preserve entity type as `resource` or `data source`, SHALL ensure newly discovered entities are present in memory with either their existing timestamp or `null`, and SHALL remove memory entries for entities that are no longer registered by either provider implementation.

#### Scenario: Working memory file is absent
- **WHEN** the memory preparation command runs and the working memory file does not yet exist
- **THEN** it bootstraps the working memory file from `.github/aw/memory/schema-coverage.json` before reconciling the entity inventory

#### Scenario: Provider registration adds a new entity
- **WHEN** the memory preparation command discovers a registered resource or data source that is not yet present in memory
- **THEN** it adds that entity under the correct type with a `null` analysis timestamp

#### Scenario: Entity inventory spans both provider implementations
- **WHEN** the memory preparation command loads the canonical entity inventory
- **THEN** it unions the registered entities from the Plugin Framework provider and the Plugin SDK provider into one de-duplicated resource/data-source inventory

#### Scenario: Provider registration removes an entity
- **WHEN** the memory preparation command finds an entity in memory that is no longer registered by either the Plugin Framework provider or the Plugin SDK provider
- **THEN** it removes that entity from the working memory file during reconciliation

### Requirement: Selection is oldest-first across resources and data sources
The schema-coverage selection command SHALL choose exactly the requested number of entities across both resources and data sources using oldest-first ordering, where `null` timestamps sort before populated timestamps, and SHALL preserve each selected entity's type in its output.

#### Scenario: Mixed analyzed and never-analyzed entities exist
- **WHEN** the selection command ranks available entities for a run
- **THEN** entities with `null` timestamps are selected before entities with populated timestamps, regardless of whether they are resources or data sources

#### Scenario: Selection output is consumed by the agent
- **WHEN** the agent requests the next entities to analyze
- **THEN** the command returns structured results that include both entity name and entity type for each selected entry

### Requirement: Timestamp persistence is handled by a script command
The schema-coverage memory update command SHALL persist the analyzed entity's timestamp as the current UTC time after each analysis, regardless of whether that analysis produced an actionable issue.

#### Scenario: An analyzed entity has actionable gaps
- **WHEN** the agent finishes analyzing an entity and determines that an issue should be created
- **THEN** the memory update command records the entity's current UTC analysis timestamp

#### Scenario: An analyzed entity has no actionable gaps
- **WHEN** the agent finishes analyzing an entity and determines that no issue should be created
- **THEN** the memory update command still records the entity's current UTC analysis timestamp
