# `elasticsearch-cluster-settings-resource-validation` — Fix false-positive validation for dynamic blocks

Implementation: `internal/elasticsearch/cluster/settings/resource.go`, `export_test.go`,
`helpers_test.go`.

## Purpose

Define requirements for the `ValidateConfig` behavior of `elasticstack_elasticsearch_cluster_settings`
with respect to unknown block values, ensuring dynamic `for_each`-driven `persistent` and
`transient` blocks do not produce false-positive "No cluster settings configured" errors at
validate time.

## CHANGED Requirements

### Requirement: `categoryBlockEmpty` treats unknown as non-empty

`categoryBlockEmpty` SHALL return `false` when the block object is unknown (`block.IsUnknown()`),
rather than treating it as absent. An unknown block value means the block's contents have not yet
been evaluated (e.g., because a `dynamic` block's `for_each` references a local variable that is
not yet resolved at `ValidateResourceConfig` time). An unknown block is not the same as a null
(absent) block.

**Previous behavior (incorrect):** `categoryBlockEmpty` returned `true` for both null and unknown
blocks, causing `validateConfigModel` to emit "No cluster settings configured" even when the block
would evaluate to a non-empty value at plan time.

**New behavior (correct):** `categoryBlockEmpty` returns `true` only for null blocks. It returns
`false` for unknown blocks, indicating that the emptiness check should be deferred.

#### Scenario: Unknown outer block is not treated as empty
- **GIVEN** a `persistent` or `transient` block object where `block.IsUnknown()` is true
- **WHEN** `categoryBlockEmpty` is called with that block
- **THEN** it SHALL return `false`

#### Scenario: Null outer block is still treated as empty
- **GIVEN** a `persistent` or `transient` block object where `block.IsNull()` is true
- **WHEN** `categoryBlockEmpty` is called with that block
- **THEN** it SHALL return `true`

### Requirement: `categoryBlockEmpty` treats unknown inner setting set as non-empty

`categoryBlockEmpty` SHALL return `false` when the inner `setting` set is unknown
(`settingSet.IsUnknown()`). This covers the case where the outer block is present
(known, non-null) but the nested `dynamic "setting"` block's contents are unknown.

**Previous behavior (incorrect):** `categoryBlockEmpty` returned `true` for an unknown setting
set, which could cause false-positive errors when a static outer block contained a
`dynamic`-driven `setting` block.

**New behavior (correct):** `categoryBlockEmpty` returns `false` for an unknown inner
`setting` set.

#### Scenario: Unknown nested setting set is not treated as empty
- **GIVEN** a block object where the outer block is known and non-null, and the `setting` set
  attribute is unknown
- **WHEN** `categoryBlockEmpty` is called with that block
- **THEN** it SHALL return `false`

### Requirement: `validateConfigModel` emits no error when either block is unknown

`validateConfigModel` SHALL NOT emit the "No cluster settings configured" error when either or
both of `persistent` and `transient` are unknown, because it cannot determine that both blocks
will be empty once evaluation completes.

#### Scenario: Both blocks unknown — no validation error
- **GIVEN** a config where both `persistent` and `transient` are unknown
- **WHEN** `validateConfigModel` is called
- **THEN** it SHALL return no diagnostics

#### Scenario: One block null, one block unknown — no validation error
- **GIVEN** a config where one of `persistent` or `transient` is null and the other is unknown
- **WHEN** `validateConfigModel` is called
- **THEN** it SHALL return no diagnostics

#### Scenario: Both blocks null — validation error still fires
- **GIVEN** a config where both `persistent` and `transient` are null
- **WHEN** `validateConfigModel` is called
- **THEN** it SHALL return an error diagnostic with summary "No cluster settings configured"

#### Scenario: Both blocks empty — validation error still fires
- **GIVEN** a config where both `persistent` and `transient` are present but contain no `setting` elements
- **WHEN** `validateConfigModel` is called
- **THEN** it SHALL return an error diagnostic with summary "No cluster settings configured"

#### Scenario: Dynamic for_each block produces no false-positive error
- **GIVEN** a `elasticstack_elasticsearch_cluster_settings` resource where `persistent` or
  `transient` is populated via a `dynamic` block whose `for_each` references a local value
- **WHEN** Terraform calls the `ValidateResourceConfig` RPC before local values are evaluated
- **THEN** the provider SHALL NOT emit an error for that resource

## ADDED Requirements

### Requirement: Unit tests cover unknown-block behavior

The package SHALL include unit tests that verify the corrected unknown-value behavior:

#### Scenario: TestValidateConfigModel_BothUnknown_OK
- **GIVEN** both `persistent` and `transient` are `types.ObjectUnknown`
- **WHEN** `ExportedValidateConfigModel` is called
- **THEN** no error diagnostic is returned

#### Scenario: TestValidateConfigModel_OneUnknown_OK
- **GIVEN** one block is `NullSettingsBlock()` and the other is `types.ObjectUnknown`
- **WHEN** `ExportedValidateConfigModel` is called
- **THEN** no error diagnostic is returned

#### Scenario: TestCategoryBlockEmpty_Unknown_NotEmpty
- **GIVEN** an unknown block object
- **WHEN** `ExportedCategoryBlockEmpty` is called
- **THEN** it returns `false`
