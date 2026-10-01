## ADDED Requirements

### Requirement: No ordering contract on the spaces list (REQ-008)

The `spaces` computed list attribute SHALL NOT be assumed to return entries in any particular order, including the built-in `default` space. Neither the Kibana Get All Spaces API nor this data source's mapping (REQ-001, REQ-004–REQ-005) guarantees a stable position for any space, and the position of a given space (including `default`) MAY differ across Kibana versions or between calls.

Acceptance tests for this data source MUST NOT assert a specific space's attributes by a fixed list index (for example `spaces.0.id`). They MUST instead look up the entry by its `id` (for example by scanning `spaces.#` and matching `spaces.<i>.id`, as `testCheckSpaceAttrByID` in `internal/kibana/spaces/data_source_test.go` already does) before asserting on its other attributes.

#### Scenario: Default space is not guaranteed to be first

- **GIVEN** a Kibana instance with the built-in `default` space and one or more additional spaces
- **WHEN** the data source reads all spaces
- **THEN** the `default` space's entry MAY appear at any position within the `spaces` list
- **AND** consumers (including acceptance tests) SHALL locate it by `id` rather than by assuming index `0`

#### Scenario: Acceptance test asserts the default space by id, not by index

- **GIVEN** an acceptance test for `elasticstack_kibana_spaces` that verifies an attribute of the `default` space
- **WHEN** the test is written or reviewed
- **THEN** the assertion SHALL look up the `default` space's entry by `id` (for example via `testCheckSpaceAttrByID`) rather than referencing `spaces.0.*`
- **AND** the test SHALL pass regardless of the position Kibana returns for the `default` space
