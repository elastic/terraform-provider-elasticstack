## Why

`elasticstack_elasticsearch_index` can report spurious mapping drift when an `elasticstack_elasticsearch_index_template` contributes mappings to an index that also has user-owned mappings. Terraform then sees the provider return a refreshed `mappings` value that is a superset of the configured value and treats the difference as provider inconsistency or replacement drift.

Phase 1 has already been built in this worktree to characterize the current behavior:

- `TestAccResourceIndexTemplateNoMappingDrift` passes: an index with no configured `mappings` and a template that injects mappings does not currently drift.
- `TestAccResourceIndexTemplateUserMappingNoDrift` reproduces the real broken case: when the index has configured `mappings` and the template injects additional mappings, apply fails with `Provider produced inconsistent result after apply`.
- The failing probe is intentionally skipped for now so it documents the bug without blocking CI.

The exact reproduced mismatch is:

- Planned/configured index mapping: `{"properties":{"user_field":{"type":"keyword"}}}`
- Refreshed API-backed state: `{"dynamic_templates":[...],"properties":{"template_field":{"type":"keyword"},"user_field":{"type":"keyword"}}}`

## What Changes

- Keep the Phase 1 acceptance coverage: one passing no-config-mappings test and one skipped user-mapping reproduction test that can be enabled when the fix lands.
- Extract the existing mapping-difference walker from `mappingsPlanModifier` into a shared helper and generalize it beyond `properties` so it can reason about top-level template-injected keys such as `dynamic_templates`.
- Introduce a `mappings` custom semantic-equality type if the user-owned mapping scenario requires it, so refreshed API mappings that are a non-drifting superset of user intent compare equal without mutating Read state.
- Reduce `mappingsPlanModifier` to replacement detection only, using the shared helper as the single definition of non-drifting mapping differences.
- Remove the existing `ignore_changes = [mappings]` workaround in the index-with-template acceptance fixture once the semantic behavior is fixed.
- Add a changelog entry referencing GitHub issue #563 and update any user-facing guidance that mentions the workaround.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `elasticsearch-index`: `mappings` comparison preserves user intent while tolerating template-injected mapping supersets.

## Impact

- Production code under `internal/elasticsearch/index/index/`, especially `mapping_modifier.go`, `schema.go`, and a new shared mapping helper/custom type file.
- Acceptance tests and fixtures under `internal/elasticsearch/index/index/acc_test.go` and `internal/elasticsearch/index/index/testdata/`.
- Existing workaround removal in `TestAccResourceIndexWithTemplate`.
- Changelog/docs updates for the fixed behavior.
