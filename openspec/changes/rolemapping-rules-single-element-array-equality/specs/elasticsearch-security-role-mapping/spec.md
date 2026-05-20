# Delta spec: `elasticsearch-security-role-mapping` — rules single-element array semantic equality

Extends [`openspec/specs/elasticsearch-security-role-mapping/spec.md`](../../../../specs/elasticsearch-security-role-mapping/spec.md).

## MODIFIED Requirements

### Requirement: JSON state mapping for rules (REQ-019, amended)

The existing requirement specifies that on read the resource SHALL serialize `rules` into a normalized JSON string and store it in state. This amendment removes the constraint that single-element arrays inside `field` objects are collapsed to strings before storage. State SHALL now store the raw JSON produced by the typed API client (which returns array form for single-element field values).

> **Rationale**: Collapsing arrays to strings in the read path caused perpetual diffs when the user's config encoded the same value as an array. Removing the collapsing simplifies the read path; semantic equality (REQ-023) handles the mismatch transparently.

#### Scenario: Rules stored as raw typed client output

- GIVEN ES returns a role mapping with `{"groups":"project1"}` (string form)
- AND the typed client deserializes it to `{"groups":["project1"]}` (array form)
- WHEN the read path stores the value in state
- THEN state SHALL contain `{"groups":["project1"]}` (array form, no single-element collapsing)

## ADDED Requirements

### Requirement: Rules semantic equality for single-element arrays (REQ-023)

The `rules` attribute on the resource SHALL use a custom Plugin Framework type (`NormalizedRulesValue`) that overrides `StringSemanticEquals`. The custom equality function SHALL normalize both the plan value and the state value by collapsing single-element arrays inside `field` objects to plain string values before comparing. The comparison SHALL be based on the normalized form; the stored plan and state values SHALL NOT be mutated.

The normalization algorithm:
1. Parse the JSON string into a tree.
2. Walk every node. For any map node that contains a `"field"` key whose value is itself a map, inspect each entry in that inner map: if the entry value is a JSON array of exactly one element, replace it with that single element (unwrap the array).
3. Recurse into all child nodes.
4. Marshal the resulting tree back to a JSON string.

On null or unknown values, or on JSON parse errors, the function SHALL fall back to the standard `jsontypes.Normalized` semantic equality.

#### Scenario: Single-element array equals string form

- GIVEN `rules` config encodes `{"field":{"groups":["project1"]}}` (array form)
- AND state holds `{"field":{"groups":"project1"}}` (string form from prior state)
- WHEN Terraform evaluates plan equality
- THEN no diff SHALL be detected

#### Scenario: Single-element array in config equals array in state

- GIVEN `rules` config encodes `{"field":{"groups":["project1"]}}` (array form)
- AND state holds `{"field":{"groups":["project1"]}}` (array form from typed client read)
- WHEN Terraform evaluates plan equality
- THEN no diff SHALL be detected

#### Scenario: Multi-element arrays compared correctly

- GIVEN `rules` config encodes `{"field":{"groups":["a","b"]}}` (multi-element array)
- AND state holds `{"field":{"groups":["a"]}}` (different value)
- WHEN Terraform evaluates plan equality
- THEN a diff SHALL be detected

#### Scenario: Null values are equal

- GIVEN both plan value and state value are null
- WHEN Terraform evaluates plan equality
- THEN no diff SHALL be detected

### Requirement: Rules attribute custom type (REQ-024)

The `rules` attribute in `schema.go` SHALL use `NormalizedRulesType{}` as its `CustomType`. The `Data.Rules` field in `models.go` SHALL be typed as `NormalizedRulesValue`. All locations in the package that construct the `rules` value (read path) SHALL use `NewNormalizedRulesValue(v string)` or `NewNormalizedRulesNull()`.

#### Scenario: Custom type used in schema

- GIVEN the resource schema definition
- WHEN the `rules` attribute is inspected
- THEN its `CustomType` SHALL be `NormalizedRulesType{}`

### Requirement: Read path stores typed client output for rules (REQ-025)

On read, the resource SHALL marshal `roleMapping.Rules` directly to JSON and store it in state as a `NormalizedRulesValue` without applying the single-element-array collapsing normalizer. The `normalizeRuleNode` walk function SHALL NOT be applied in the read path.

#### Scenario: Array form preserved in state on read

- GIVEN a role mapping with a single-element field rule value in ES
- WHEN the provider reads the role mapping
- THEN state SHALL store the array form returned by the typed client, not a collapsed string
