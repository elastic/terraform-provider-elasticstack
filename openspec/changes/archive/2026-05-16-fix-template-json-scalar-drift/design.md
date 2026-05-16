## Context

The provider models component-template and index-template `template.mappings` and `template.settings` as shared custom Plugin Framework string types. Those types are responsible for suppressing non-drifting differences between practitioner-authored JSON and the JSON Elasticsearch returns after create/read.

Two regressions surfaced after the 0.15.x Plugin Framework migration:

- `MappingsValue` compares many non-structural leaves with `reflect.DeepEqual`, so practitioner-authored JSON scalars such as `true` are treated as different from Elasticsearch echoes like `"true"`.
- `IndexSettingsValue` already stringifies flattened setting values for comparison, but it canonicalizes Go `nil` as `"<nil>"`, which is not semantically equal to Elasticsearch's `"null"` string echo.

Both regressions appear in `elasticstack_elasticsearch_component_template`, but the custom types are shared with index templates, so the fix must preserve consistent behavior across all consumers. At the same time, the provider must remain conservative: only scalar echo normalization should be relaxed, while object/array structure and existing mapping superset semantics must stay intact.

## Goals / Non-Goals

**Goals:**
- Restore non-drifting behavior when Elasticsearch echoes semantically equivalent scalar values as strings in template mappings and settings.
- Keep mapping comparison strict about structure, field ownership, and true value changes.
- Add narrow tests that lock in the reported regressions without broadening behavior beyond template JSON custom types.
- Document the behavior explicitly in the relevant template specs.

**Non-Goals:**
- Broad normalization changes for unrelated JSON custom types elsewhere in the provider.
- Changes to Elasticsearch API payload generation or to how Terraform configuration is authored.
- Relaxing object-vs-array or other structural mismatches in mappings/settings comparison.

## Decisions

### Decision 1: Relax only scalar leaf comparisons in `MappingsValue`

`MappingsValue` will keep its existing recursive structure-aware comparison and bidirectional non-drifting superset logic. The change will be limited to leaf-value comparison: when two values differ only because one side is a JSON scalar and the other is the equivalent Elasticsearch stringified scalar echo, they will be considered semantically equal.

**Rationale:** The regression is caused by overly strict scalar equality, not by the overall mapping comparison model. Limiting the change to scalar leaves fixes #2987 while preserving current behavior for nested objects, arrays, field ownership, and template-injected extras.

**Alternative considered:** Canonicalize the entire mappings JSON tree up front by converting all scalars to strings. Rejected because it would blur meaningful structural/value distinctions and make the mapping type harder to reason about.

### Decision 2: Canonicalize JSON `null` in `IndexSettingsValue` as `"null"`

`IndexSettingsValue` will continue to flatten settings, normalize keys, and stringify scalar values. The canonicalization helper will be adjusted so JSON `null` normalizes to `"null"` instead of Go's `"<nil>"` string representation.

**Rationale:** This is the narrowest fix for #2988 and preserves the current settings-comparison model, including dotted-vs-nested key normalization and string-based comparison of effective settings.

**Alternative considered:** Add custom post-comparison exceptions just for `nil` vs `"null"`. Rejected because canonicalization is the correct layer for this behavior and keeps subsequent comparisons simple.

### Decision 3: Cover shared custom types with unit tests and reported resource paths with acceptance tests

The change will add unit tests for the shared custom types and focused component-template acceptance tests for the reported boolean-mappings and null-settings regressions.

**Rationale:** Unit tests are the best place to pin the shared semantic rules precisely. Acceptance tests ensure the real resource path no longer produces inconsistent-result errors after apply.

**Alternative considered:** Add equivalent acceptance coverage for every resource using the shared types. Rejected as unnecessary expansion; shared unit tests plus direct component-template regressions provide sufficient confidence for this fix.

## Risks / Trade-offs

- **[Risk]** Scalar normalization could become too permissive and hide real drift.
  - **Mitigation:** Limit equivalence to scalar-vs-stringified-scalar comparisons and add negative unit cases for genuinely different values.

- **[Risk]** Shared-type changes could alter index-template behavior in subtle ways.
  - **Mitigation:** Scope the normalization narrowly, keep existing structural logic intact, and update the shared template specs so the behavior is explicit rather than accidental.

- **[Risk]** Acceptance tests may depend on Elasticsearch echo behavior that differs by version.
  - **Mitigation:** Keep acceptance scenarios close to the reported regressions and use unit tests to lock in the exact custom-type semantics.
