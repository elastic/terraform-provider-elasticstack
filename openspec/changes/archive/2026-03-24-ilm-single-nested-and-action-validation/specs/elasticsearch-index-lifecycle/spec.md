## ADDED Requirements

### Requirement: Single nested blocks for phases and actions (REQ-020)

The resource SHALL model each of the phase blocks `hot`, `warm`, `cold`, `frozen`, and `delete` as a **Plugin Framework `SingleNestedBlock`** (at most one block per phase in configuration; state stores a single nested object or null when absent), not as a list nested block with a maximum length of one.

Each ILM action block allowed under a phase (for example `set_priority`, `rollover`, `forcemerge`, `searchable_snapshot`, `wait_for_snapshot`, `delete`, and other actions defined by the provider schema) SHALL likewise be modeled as a **`SingleNestedBlock`**.

The **`elasticsearch_connection`** block SHALL remain a **list nested block** as provided by the shared provider connection schema.

#### Scenario: Phase block cardinality

- GIVEN a Terraform configuration for this resource
- WHEN the user declares a phase (for example `hot { ... }`)
- THEN the schema SHALL allow at most one such block for that phase and SHALL persist that phase as an object-shaped value in state, not as a single-element list

#### Scenario: Action block cardinality

- GIVEN a phase that supports an ILM action block
- WHEN the user declares that action (for example `forcemerge { ... }`)
- THEN the schema SHALL allow at most one such block and SHALL persist it as an object-shaped value in state, not as a single-element list

### Requirement: State schema version and upgrade (REQ-021)

The resource SHALL use a **non-zero** `schema.Schema.Version` for this resource type after this change.

The resource SHALL implement **`ResourceWithUpgradeState`** and SHALL migrate stored Terraform state from the **prior version** (list-shaped nested values for phases and ILM actions) to the **new version** (object-shaped nested values) for the same logical configuration.

The migration SHALL unwrap list-encoded values **only** for known ILM phase keys and known ILM action keys under those phases (including the delete-phase ILM action block named `delete`). The migration SHALL **not** alter the encoding of **`elasticsearch_connection`**.

#### Scenario: Upgrade from list-shaped phase state

- GIVEN persisted state where a phase is stored as a JSON array containing one object
- WHEN Terraform loads state and runs the state upgrader
- THEN the upgraded state SHALL store that phase as a single object (or equivalent null) consistent with `SingleNestedBlock` semantics

#### Scenario: Connection block unchanged by upgrade

- GIVEN persisted state that includes `elasticsearch_connection` as a list
- WHEN the state upgrader runs
- THEN the `elasticsearch_connection` value SHALL remain list-shaped as defined by the connection schema

### Requirement: Action fields optional with object-level AlsoRequires (REQ-022)

For the ILM action blocks **`forcemerge`**, **`searchable_snapshot`**, **`set_priority`**, **`wait_for_snapshot`**, and **`downsample`**, each attribute that is **required for API correctness when the action is declared** SHALL be **optional** at the Terraform attribute level (so an entirely omitted action block does not force those attributes to appear).

When the user **declares** one of these action blocks in configuration, validation SHALL require that all of the following previously required attributes are set (non-null), using object-level validation equivalent to **`objectvalidator.AlsoRequires`**:

- **`forcemerge`**: `max_num_segments`
- **`searchable_snapshot`**: `snapshot_repository`
- **`set_priority`**: `priority`
- **`wait_for_snapshot`**: `policy`
- **`downsample`**: `fixed_interval`

Existing attribute-level validators (for example minimum values) SHALL remain on those attributes where applicable.

#### Scenario: Omitted action block is valid

- GIVEN a phase without a particular action block (for example no `forcemerge` block)
- WHEN Terraform validates configuration
- THEN validation SHALL NOT fail solely because `max_num_segments` is unset

#### Scenario: Empty action block is invalid

- GIVEN the user declares `forcemerge { }` with no attributes
- WHEN Terraform validates configuration
- THEN validation SHALL fail with a diagnostic indicating the required fields when the block is present

#### Scenario: Searchable snapshot requires repository when present

- GIVEN the user declares `searchable_snapshot { force_merge_index = true }` without `snapshot_repository`
- WHEN Terraform validates configuration
- THEN validation SHALL fail with a diagnostic
