# Spec-Drift Review — fix-index-mapping-update-unidirectional-check

**Lane:** spec-drift (Spec Memory)
**Iteration:** 2
**Status:** implementation complete — all prior findings resolved, no new drift detected

---

## Summary

Reviewed the committed implementation against the accumulated `elasticsearch-index`
spec corpus (21 main requirements + delta spec at
`openspec/changes/fix-index-mapping-update-unidirectional-check/specs/elasticsearch-index/spec.md`).

All five findings from the iteration-1 design-level review (C1, C2, W1, W2, S1) are
resolved by the committed implementation. No new spec drift was found.

---

## Prior findings — resolution status

### C1: `updateMappings` used bidirectional predicate — **RESOLVED**

`internal/elasticsearch/index/index/update.go:197` now calls
`planMappings.RequiresMappingsUpdate(ctx, stateMappings)` and skips the Put Mapping
call when `requiresUpdate` is false. The bidirectional `StringSemanticEquals` gate is
gone from the update path.

### C2: `RequiresMappingsUpdate` method did not exist — **RESOLVED**

`internal/elasticsearch/index/mappings_value.go:167-185` adds the method with correct
unidirectional semantics:

```go
return !MappingsSemanticallyEqual(planMap, stateMap), diags
```

`planMap` is the receiver (plan intent); `stateMap` is the argument (prior API state).
`MappingsSemanticallyEqual(planMap, stateMap)` asks "does state already cover everything
plan wants?" Negating gives "plan has something state lacks → PUT needed." The argument
order is load-bearing; the inline comment names the issue reference (elastic/protections-cloud#19769)
so future editors cannot silently swap it.

### W1: Adoption path had identical drift — **RESOLVED**

`internal/elasticsearch/index/index/create.go:183` delegates to the same `updateMappings`
helper, so the fix propagates automatically. Verified: `adoptExistingIndexOnCreate` calls
`r.updateMappings(ctx, client, concreteName, plan.Mappings, synthetic.Mappings)` where
`synthetic.Mappings` comes from the live existing index.

### W2: No regression tests for contractual scenarios — **RESOLVED**

Three spec scenarios are now covered:

| Scenario | Coverage |
|---|---|
| Adding a mapping field calls Put Mapping API | `TestAccResourceIndexMappingsUpdateRegression` + `checkIndexHasField` direct ES read |
| State superset of plan → skip Put Mapping | Unit: `TestIndexMappingsValue_RequiresMappingsUpdate/"state is superset of plan"` (9 cases total) |
| Adoption adds plan field, retains live-only field | `TestAccResourceIndexUseExistingAdoptMappings` + two `checkIndexHasField` assertions |

The regression test (`TestAccResourceIndexMappingsUpdateRegression`) verifies via direct ES
API read (`getElasticsearchIndexStateByName`) that `field2` actually lands in the live cluster
after the update step — this correctly catches the case where Terraform state reflects the
planned value even when the PUT was silently skipped.

The template-no-mapping-drift acceptance tests (`TestAccResourceIndexTemplateNoMappingDrift`
at `acc_test.go:366`, `TestAccResourceIndexUseExistingTemplateNoMappingDrift` at
`acc_test.go:787`) remain in place and continue to cover the "template-injected extras do not
cause mapping update" scenario.

### S1: Main spec retained pre-delta language — **RESOLVED**

`openspec show elasticsearch-index --type spec --json --requirements` now returns the
unidirectional update-flow language with the three scenarios. The accumulated corpus is
internally consistent.

---

## Accumulated spec compliance — no new drift

All 21 accumulated requirements for `elasticsearch-index` were reviewed against the touched
files. The change is narrowly scoped to the update-decision logic; no other requirement is
affected.

Key compliance points for the touched surface:

1. **Update flow (requirement 9)** — unidirectional gate correctly implemented:
   `RequiresMappingsUpdate` returns `true` only when plan has content absent from state.
   `StringSemanticEquals` (bidirectional) remains untouched and continues to serve
   plan-time drift suppression and replacement decisions only, as the spec requires.

2. **Adoption flow (requirement 15)** — uses the same `updateMappings` helper, satisfying
   "same unidirectional Put Mapping decision." Fields present only in the existing index are
   retained because `RequiresMappingsUpdate` checks from plan's perspective, not state's.

3. **Null/unknown guards** — `RequiresMappingsUpdate` returns `false` for null/unknown plan
   (nothing planned), `true` for null/unknown state with a known plan (state has nothing,
   plan adds everything). These edge cases are tested in the unit suite.

---

## Verdict

**`approve`**

No spec drift detected. All prior findings are resolved, the implementation faithfully
satisfies the accumulated `elasticsearch-index` requirements, and the new regression tests
provide direct API-level evidence that the fix is behaviorally correct.
