# snapshot-domain-layout Specification

## Purpose
TBD - created by archiving change reorganize-snapshot-domain-packages. Update Purpose after archive.
## Requirements
### Requirement: Snapshot domain entities are colocated under elasticsearch/snapshot/
All Terraform entities that operate on Elasticsearch snapshots SHALL be implemented within `internal/elasticsearch/snapshot/` or one of its subpackages.

#### Scenario: Repository entity location
- **WHEN** a developer navigates to the snapshot repository implementation
- **THEN** the resource, datasource, schema, and model files SHALL be found under `internal/elasticsearch/snapshot/repository/`

#### Scenario: Lifecycle entity location
- **WHEN** a developer navigates to the snapshot lifecycle (SLM) implementation
- **THEN** the resource, schema, and model files SHALL be found under `internal/elasticsearch/snapshot/lifecycle/`

#### Scenario: Action entities location
- **WHEN** a developer navigates to the snapshot create or restore action implementations
- **THEN** the action, schema, and model files SHALL be found under `internal/elasticsearch/snapshot/create/` and `internal/elasticsearch/snapshot/restore/` respectively

#### Scenario: No snapshot code remains in cluster/
- **WHEN** inspecting `internal/elasticsearch/cluster/`
- **THEN** no snapshot-related resource, datasource, action, or shared model files SHALL remain

