# `ci-schema-coverage-rotation-memory` — Script-driven schema-coverage memory behavior

Workflow implementation: authored source under `.github/workflows-src/schema-coverage-rotation/`, compiled to `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml`.
Script implementation: `scripts/schema-coverage-rotation/`.

## Purpose

Define requirements for moving schema-coverage memory bootstrap, entity reconciliation, selection, and timestamp persistence out of workflow prompt prose and into repository-local Go commands that the workflow invokes after repo-memory hooks complete.

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

### Requirement: Memory commands use a caller-supplied working file path
The schema-coverage memory preparation, selection, and update commands SHALL accept the live working memory file path as an explicit input parameter and SHALL operate on that supplied path rather than hardcoding a runtime-specific repo-memory location.

#### Scenario: Workflow passes the live repo-memory path
- **WHEN** the workflow invokes a schema-coverage memory command against the current repo-memory workspace
- **THEN** it provides the working memory file path as a command input and the command operates on that supplied file

### Requirement: Memory preparation bootstraps and reconciles the canonical entity inventory
The schema-coverage memory preparation command SHALL use the caller-supplied working memory file path, SHALL initialize that file from `.github/aw/memory/schema-coverage.json` when the supplied path does not exist, SHALL derive the canonical entity inventory by importing the provider registrations exposed through `provider/plugin_framework.go` and `provider/provider.go`, SHALL preserve entity type as `resource` or `data source`, SHALL ensure newly discovered entities are present in memory with either their existing timestamp or `null`, SHALL remove memory entries for entities that are no longer registered by either provider implementation, and SHALL write reconciled results via an atomic replace-on-success rename.

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
The schema-coverage selection command SHALL choose exactly the requested number of entities across both resources and data sources using oldest-first ordering, where `null` timestamps sort before populated timestamps, ties on equal timestamp state and value are broken by lexicographic `type` then `name`, and SHALL emit a JSON array on stdout whose entries use stable `type` and `name` fields for each selected entity.

#### Scenario: Mixed analyzed and never-analyzed entities exist
- **WHEN** the selection command ranks available entities for a run
- **THEN** entities with `null` timestamps are selected before entities with populated timestamps, regardless of whether they are resources or data sources

#### Scenario: Selection output is consumed by the agent
- **WHEN** the agent requests the next entities to analyze
- **THEN** the command returns a JSON array whose entries include both stable `type` and `name` fields for each selected entity

#### Scenario: Multiple entities share the same timestamp
- **WHEN** two or more candidate entities have the same timestamp state and value during selection
- **THEN** the command orders those tied entities by lexicographic `type` and then lexicographic `name`

### Requirement: Timestamp persistence is handled by a script command
The schema-coverage memory update command SHALL accept the caller-supplied working memory file path and one or more analyzed entities, SHALL persist all analyzed entities' timestamps as the current UTC time in a single atomic write after all selected entities have been analyzed, regardless of whether any individual analysis produced an actionable issue, and SHALL update the memory file via an atomic replace-on-success rename.

#### Scenario: All analyzed entities are recorded in one write
- **WHEN** the agent completes analysis of all selected entities in a run
- **THEN** the memory update command records the current UTC analysis timestamp for every analyzed entity in a single atomic write

#### Scenario: Entities without actionable gaps are still recorded
- **WHEN** an analyzed entity has no actionable testing gaps
- **THEN** the memory update command still includes that entity's current UTC analysis timestamp in the batch write
