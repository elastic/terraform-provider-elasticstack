## Context

The index resource currently reads Elasticsearch mappings in `populateFromAPI` and writes the full `_mappings` response into `model.Mappings`. That response includes mappings contributed by matching index templates. For an index with no configured mappings, current acceptance coverage shows the provider already avoids plan drift. For an index with configured mappings, current acceptance coverage shows Terraform detects an inconsistent result immediately after apply because refreshed state includes both the user-owned mapping and template-owned extras.

The existing `mappingsPlanModifier` already contains useful intent-preserving logic for one class of mapping differences: it walks prior-state `properties`, detects incompatible type changes as replacement, and copies fields Elasticsearch will retain when users remove them from config. That logic is currently narrow and plan-modifier-local.

## Goals / Non-Goals

**Goals:**
- Reproduce and document both relevant scenarios before changing production code.
- Preserve user intent for `mappings` while treating template-injected mappings as non-drift.
- Keep a single implementation of mapping-difference semantics shared by semantic equality and replacement detection.
- Remove test-level `ignore_changes` workarounds after the provider behavior is fixed.
- Keep the change scoped to index mappings; do not alter settings or aliases drift behavior.

**Non-Goals:**
- Refactoring the broader index Read flow.
- Migrating away from `Optional+Computed` for `mappings`.
- Changing settings, aliases, index template resources, or Elasticsearch API behavior.
- Preserving compatibility with the current in-branch workaround once the fix is implemented.

## Decisions

### 1. Keep Phase 1 tests as executable documentation

Add two acceptance probes:

- `TestAccResourceIndexTemplateNoMappingDrift`: index has no configured `mappings`; template injects `dynamic_templates` and `properties`; second step asserts `plancheck.ExpectEmptyPlan()`. This currently passes.
- `TestAccResourceIndexTemplateUserMappingNoDrift`: index config owns `user_field`; template injects `dynamic_templates` and `template_field`; second step will assert `plancheck.ExpectEmptyPlan()` after the fix. This currently fails during first apply and is skipped until the fix lands.

Why:
- The passing test prevents accidentally regressing the no-config case.
- The skipped failing test records the real current failure and gives the implementation a precise acceptance target.

### 2. Extract the mapping walker before changing behavior

Move the recursive mapping comparison from `mappingsPlanModifier.modifyMappings` into a shared helper in a new file such as `mappings_walker.go`.

The helper should:
- Walk nested `properties` recursively.
- Preserve existing replacement detection for changed or missing field `type`.
- Preserve the warning-oriented behavior for fields Elasticsearch retains after removal.
- Generalize top-level comparison so API-only keys such as `dynamic_templates`, `_meta`, `runtime`, or template-injected `properties` can be recognized as template-owned extras instead of user-owned drift.
- Return enough structure for both semantic equality and plan replacement detection without forcing either caller to mutate arbitrary JSON maps.

Why:
- It prevents the Read path, semantic equality, and plan modifier from defining subtly different drift rules.
- It allows fast unit coverage for mapping semantics without relying only on acceptance tests.

### 3. Prefer custom semantic equality for the user-owned mapping case

For the reproduced case where both config-owned and template-owned mappings exist, introduce a custom string-backed mappings type:

- `mappingsType` should extend or wrap `jsontypes.NormalizedType`.
- `mappingsValue` should extend or wrap `jsontypes.Normalized`.
- `StringSemanticEquals` should unmarshal both strings, call the shared walker, and return true when the refreshed/API value is a non-drifting superset of the prior user-intent value.
- The `mappings` attribute in `schema.go` should switch `CustomType` from `jsontypes.NormalizedType{}` to `mappingsType{}` while preserving the same serialized state shape.

Why:
- Terraform Plugin Framework semantic equality is the right primitive for "these two serialized strings represent the same logical mapping intent".
- It avoids Read-side state surgery and lets Plugin Framework preserve prior state when the API returns a semantically equal superset.
- It handles the reproduced inconsistent-result-on-apply failure earlier than a plan modifier can.

### 4. Shrink the plan modifier to replacement decisions

After semantic equality owns non-drift equivalence, `mappingsPlanModifier` should stop copying state-only fields into the plan. It should retain replacement detection for incompatible user-owned mapping changes and use the shared helper to make that decision.

Why:
- Value mutation in the plan modifier is no longer the core drift strategy.
- Replacement behavior remains explicit and focused.

### 5. Remove the existing workaround after behavior is fixed

Remove `lifecycle { ignore_changes = [mappings] }` from `TestAccResourceIndexWithTemplate` and update the expected state assertion according to the semantic-equality behavior.

Why:
- Acceptance tests should validate provider behavior directly.
- Keeping `ignore_changes` would continue to mask the bug class this change fixes.

## Risks / Trade-offs

- [Semantic equality preserves too much] -> Mitigation: unit-test conflicting user-owned type changes and missing owned fields so replacement/drift behavior remains intact.
- [Semantic equality preserves too little] -> Mitigation: acceptance-test template-injected `properties` and top-level `dynamic_templates` together.
- [Custom type changes persisted state format] -> Mitigation: keep the type string-backed and normalized, and verify acceptance state behavior.
- [Skipped failing test is forgotten] -> Mitigation: make unskipping it an explicit task before final verification.

## Migration Plan

1. Keep the Phase 1 probes in place, with only the currently failing user-mapping test skipped.
2. Extract and unit-test the shared mapping walker.
3. Add the custom semantic-equality type and wire it into `schema.go`.
4. Simplify `mappingsPlanModifier` to replacement detection.
5. Unskip `TestAccResourceIndexTemplateUserMappingNoDrift`, remove the `ignore_changes` workaround, and run targeted acceptance verification.

## Open Questions

- None for the current scope. Phase 1 already shows the no-config case passes and the user-owned mapping case is the behavior that needs the semantic-equality fix.
