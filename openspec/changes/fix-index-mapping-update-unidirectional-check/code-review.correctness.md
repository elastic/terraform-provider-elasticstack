# Correctness Review — fix-index-mapping-update-unidirectional-check

**Lane:** correctness  
**Iteration:** 2  
**Status:** implementation complete; full code review performed

---

## Summary

All 13 tasks are checked off. The implementation is present and correct. Core logic
(`RequiresMappingsUpdate`), wire-up (`update.go`, `create.go`), unit tests, acceptance
tests, and spec delta have all been reviewed against the proposal, design, and spec.

All findings from the previous iteration are resolved. Two minor issues remain
(one documentation gap, one dead-code style nit) but neither affects correctness.

---

## Previous-iteration findings: resolution status

| Finding | Status |
| ------- | ------ |
| C1 — No implementation | **RESOLVED** — all 13 tasks complete |
| W1 — missing `typeutils.IsKnown` guard at `update.go:90` | **RESOLVED** — guard added at `update.go:91–96` |
| W2 — doc comment should note non-`properties` limitation | **OPEN** — see S1 below |
| S1 — tasks.md note about guard asymmetry | **RESOLVED** — guard was added |

---

## Findings

### SUGGESTION — S1: `RequiresMappingsUpdate` doc comment omits the non-`properties` scope limitation

**Location:** `internal/elasticsearch/index/mappings_value.go:155–166`

**Description:** The previous iteration (W2) requested that the doc comment for
`RequiresMappingsUpdate` note that the "template-injected extras don't trigger a
redundant PUT" guarantee applies only at the `properties` level. The current comment
does not include this note.

Concretely: `MappingsSemanticallyEqual` iterates over plan keys and delegates non-`properties`
keys (e.g. `runtime`, `_meta`, `dynamic_templates`) to `scalarSemanticEqual` which uses
`reflect.DeepEqual` for map/slice values. If state has Elasticsearch-injected extras
*within* a `runtime` map that the plan also declares, `scalarSemanticEqual` returns `false`
→ `MappingsSemanticallyEqual` returns `false` → `RequiresMappingsUpdate` returns `true` →
spurious PUT is issued. This is a pre-existing limitation of `MappingsSemanticallyEqual`,
not introduced by this change, but the doc comment should set expectations for future
maintainers.

**Recommended addition** to the doc comment block:

```
// The "template-injected extras don't trigger a redundant PUT" guarantee applies only at
// the properties level. If plan declares a top-level non-properties key (e.g. runtime, _meta)
// and state has Elasticsearch-injected extras *within* that key, scalarSemanticEqual uses
// reflect.DeepEqual and may return false, causing a spurious PUT. This is a pre-existing
// limitation of MappingsSemanticallyEqual; it does not affect the properties superset check.
```

**Severity:** SUGGESTION — no correctness impact; documentation gap only.

---

### SUGGESTION — S2: Dead `var diags` declaration in `updateMappings`

**Location:** `internal/elasticsearch/index/index/update.go:195`

**Description:** `updateMappings` declares `var diags diag.Diagnostics` at line 195 and
then immediately overwrites it via short-variable-declaration `:=` at line 197
(`requiresUpdate, diags := planMappings.RequiresMappingsUpdate(...)`). The `var` declaration
creates an empty slice that is never read before being replaced. This does not affect
correctness (`:=` reassigns the existing variable when the name is already in scope) but
the dead declaration is misleading.

**Suggested fix:** drop the `var diags diag.Diagnostics` line and declare `diags` directly
with the `:=` result from `RequiresMappingsUpdate`:
```go
requiresUpdate, diags := planMappings.RequiresMappingsUpdate(ctx, stateMappings)
```
(already the actual declaration — just remove the earlier `var`).

**Severity:** SUGGESTION — style nit; zero runtime impact.

---

## Correctness confirmations

The following aspects were verified against the spec and design and are confirmed correct:

1. **`RequiresMappingsUpdate` argument order** (`mappings_value.go:184`):
   `!MappingsSemanticallyEqual(planMap, stateMap)` — plan as first arg, state as second.
   `MappingsSemanticallyEqual(planMap, stateMap)` asks "does state cover everything plan
   wants?" Negating gives "update required." Reversing the arguments would silently
   reintroduce `elastic/protections-cloud#19769`. ✓

2. **Null/unknown handling** (`mappings_value.go:170–175`):
   - Plan null or unknown → `!IsKnown(v)` → `false` (nothing planned to add). ✓
   - State null/unknown with known plan → second guard `!IsKnown(state)` → `true`. ✓
   - Both null/unknown → first guard fires → `false`. ✓
   All three cases from the spec are correctly handled. ✓

3. **Branch inversion at wire site** (`update.go:197–204`):
   Old: `areEqual, _ := StringSemanticEquals(...)` then `if areEqual { return nil }`.
   New: `requiresUpdate, diags := RequiresMappingsUpdate(...)` then `if !requiresUpdate { return nil }`.
   Logically correct inversion. ✓

4. **null-plan guard at `update.go:91`** (W1 from previous iteration):
   `if typeutils.IsKnown(planModel.Mappings) { r.updateMappings(...) }` added, matching the
   pattern in `create.go:182`. The asymmetry is resolved. ✓

5. **`create.go` adoption path picks up fix** (`create.go:182–187`):
   `adoptExistingIndexOnCreate` calls `r.updateMappings` which now uses
   `RequiresMappingsUpdate`. No separate call site to update. ✓

6. **`StringSemanticEquals` and `MappingsSemanticallyEqual` untouched**:
   Plan-time semantic equality and `RequiresReplace` behaviour (REQ-022–REQ-024) are
   unchanged. ✓

7. **Unit tests pin argument order** (`mappings_value_test.go:422–499`):
   9 cases covering: plan adds field (true), plan strict superset (true), state superset of
   plan (false), equal (false), plan null/unknown (false), state null/unknown (true). The
   "state superset of plan" case (`field1Plus3` as state, `field1Only` as plan → `false`)
   would return `true` if arguments to `MappingsSemanticallyEqual` were swapped, pinning
   the correct argument order. ✓

8. **Regression acceptance test uses direct API read** (`acc_test.go:1114–1121`):
   `checkIndexHasField` reads the live cluster via `typedClient.Indices.Get`, bypassing
   Terraform state — this correctly catches the case where a PUT was skipped but the planned
   field appears in refreshed state. ✓

9. **Adoption acceptance test exercises both halves of the asymmetry** (`acc_test.go:1130–1162`):
   `new_field` is written (was in plan, absent from live), `legacy_field` is retained (was in
   live, absent from plan). Both checked via direct API read. ✓

10. **Spec delta coverage** (`specs/elasticsearch-index/spec.md`):
    REQ-015–REQ-018 updated with unidirectional language and all three scenarios (template
    extras no PUT, adding field triggers PUT, state covers plan skips PUT). Adoption
    requirement updated with unidirectional language and the mixed add/omit scenario. ✓

---

## Verdict

`approve` — implementation is correct and complete. All spec SHALL requirements are
satisfied, all prior findings are resolved, and tests correctly pin both halves of
the fix. The two SUGGESTION-level findings (doc comment and dead `var`) are editorial;
they do not require a re-review iteration.
