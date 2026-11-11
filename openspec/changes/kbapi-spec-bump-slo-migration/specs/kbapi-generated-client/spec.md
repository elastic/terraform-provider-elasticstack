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

- **GIVEN** the target spec's inline `Cases_attachment_type_filter` array branch has the same `maxItems`
  bound as `Cases_string_array`
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

A transform that sets a value at a hardcoded component schema path (for example `fixSloFloatFormats` and
`fixSloResponseArtifacts`) SHALL be covered by a regeneration-time guard that fails
`make -C generated/kbapi all` when the downloaded spec does not already contain the node the transform
targets, rather than allowing the transform to silently create an unused path with no build or test
failure.

#### Scenario: Guard fails on a renamed target path

- **GIVEN** a schema-name-dependent transform targets a component path that does not exist in the
  just-downloaded spec (for example because the upstream schema renamed or restructured that component)
- **WHEN** `make -C generated/kbapi all` runs
- **THEN** the `make` invocation SHALL fail with a diagnostic identifying the missing path
- **AND** the transform SHALL NOT silently succeed while having no effect

#### Scenario: Guard passes when the target path exists

- **GIVEN** every schema-name-dependent transform's target path exists in the downloaded spec
- **WHEN** `make -C generated/kbapi all` runs
- **THEN** the guard SHALL NOT fail the build
