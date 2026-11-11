## ADDED Requirements

### Requirement: Pinned OpenAPI spec ref builds cleanly

`generated/kbapi/Makefile`'s `github_ref` SHALL point at one explicit, immutable Kibana `main` commit
SHA. Running `make -C generated/kbapi all` against that ref SHALL produce a `generated/kbapi/kibana.gen.go`
that compiles on its own, with no undefined-type or undefined-constant errors arising from inline
`oneOf`/`anyOf` query parameters that `oapi-codegen` cannot name.

#### Scenario: Cases attachment type filter parameter compiles

- **WHEN** `make -C generated/kbapi all` runs against the pinned ref
- **THEN** the `Cases_attachment_type_filter` query parameter SHALL be represented as named `$ref`
  component types rather than emitting unnamed union accessor methods that reference undefined `N0`/`N1`
  types
- **AND** `go build ./generated/kbapi/...` SHALL succeed

### Requirement: Cases attachment type filter transform preserves the parameter's value contract

The `fixCasesAttachmentTypeFilterParameter` transform SHALL replace `Cases_attachment_type_filter`'s
inline `oneOf` of a string and a string array with `$ref`s to existing `Cases_string` and
`Cases_string_array` component schemas only when those components' shape matches the target spec's
original inline array schema. When the shapes differ, the transform SHALL use a dedicated array
component instead of reusing `Cases_string_array` unmodified.

#### Scenario: Reused array component preserves maxItems

- **GIVEN** the target spec's inline `Cases_attachment_type_filter` array branch has the same item schema
  and array constraints (including `minItems`, `maxItems`, and `uniqueItems`) as `Cases_string_array`
- **WHEN** the transform runs
- **THEN** the parameter SHALL reference `Cases_string_array` and SHALL NOT alter its `maxItems` value

#### Scenario: Divergent array shape does not silently change on reuse
- **GIVEN** the target spec's inline array branch has no `maxItems` or a different bound than
  `Cases_string_array`
- **WHEN** the transform runs
- **THEN** the parameter SHALL NOT reference `Cases_string_array` unmodified
- **AND** the transform SHALL use a dedicated component schema that preserves the target spec's own
  bound

### Requirement: Schema-name-dependent transforms fail loudly when their source node is missing

A regeneration-time guard SHALL verify the existing source schema/property nodes used by
`fixCasesAttachmentTypeFilterParameter`, `fixSloFloatFormats`, and `fixSloResponseArtifacts`. It SHALL fail
`make -C generated/kbapi all` with a diagnostic identifying any missing source node.
For numeric overrides, the source node is the existing numeric property schema, not the `format` key
being added. For artifacts injection, the source nodes are the existing response `properties` map and
referenced artifacts component, not the `properties.artifacts` property being added.

#### Scenario: Guard fails on a renamed source path

- **GIVEN** a required source schema/property path does not exist in the just-downloaded spec
- **WHEN** `make -C generated/kbapi all` runs
- **THEN** the `make` invocation SHALL fail with a diagnostic identifying the missing source path
- **AND** the transform SHALL NOT silently succeed while having no effect

#### Scenario: Guard passes when source nodes exist but output leaves are absent

- **GIVEN** all required source schema/property nodes exist in the downloaded spec
- **AND** the `format` keys or response `artifacts` property to be added are absent
- **WHEN** `make -C generated/kbapi all` runs
- **THEN** the guard SHALL NOT fail the build
