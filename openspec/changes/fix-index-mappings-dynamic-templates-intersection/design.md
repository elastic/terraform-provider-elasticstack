## Context

`Read` for `elasticstack_elasticsearch_index_mappings` is supposed to keep only the user-declared subset of `mappings` (REQ-004). `intersectProperties` already does this correctly for `properties` by recursing name-by-name. `intersectMappings` (`internal/elasticsearch/index/indexmappings/intersect.go:38-64`) special-cases only `properties` (line 47); every other top-level key falls through to `index.FieldSemanticallyEqual(stateVal, apiVal)` (line 57), which type-asserts to `map[string]any`. `dynamic_templates` is `[]any`, so that assertion always fails and `result[key] = apiVal` (line 61) stores the entire API array — user templates plus anything an index template injected — into state.

Two maintainer decision rounds on the issue, together with the prior implementation-research comment, fully pin down the fix:

1. **Package consolidation direction**: move `intersect.go` into `internal/elasticsearch/index` (not the reverse), so the fix reuses `dynamicTemplatesByName` (added in #4581, `mappings_value.go:427-452`) in-package rather than exporting it as new public API surface.
2. **Drop-on-missing**: a declared `dynamic_templates` name absent from the API (deleted out-of-band) is dropped from state, mirroring `intersectProperties`'s existing omit-on-missing at `intersect.go:73-76`. Drift then surfaces on the next plan instead of state silently pinning a stale value. This applies both when the API array still exists but omits the name, and when the API omits the `dynamic_templates` key entirely (e.g. every declared template was removed out-of-band) — the latter case must be handled explicitly, because `intersectMappings`'s generic "API key absent → keep `stateVal`" early return (`intersect.go:41-45`) runs before any key-specific branch and would otherwise pin the stale array.
3. **Invalid state shape** (duplicate template name, or a template value that isn't an object — only reachable via manual state edits/import, since Elasticsearch itself won't produce this shape): fall back to today's whole-array passthrough for that key, not drop the key outright.
4. **Test tightening**: `checkStateMappingsDynamicTemplates` (`acc_test.go:636-651`) moves from a `len >= minCount` assertion to an exact-name-set assertion, must additionally assert `template_default` is absent from state, and every existing call site (`TestAccResourceIndexMappings_allTopLevelKeys` at `acc_test.go:179` and `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` at `acc_test.go:210`) is updated to pass its fixture's expected declared name(s). `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` additionally gets a third step exercising decision 2 (a declared name deleted out-of-band).
5. **Order-sensitive semantic equality**: `dynamicTemplatesSemanticallyEqual` (`mappings_value.go`) currently ignores array order by design ("API extras and array ordering are intentionally ignored"). Since this change makes `Read` preserve the state's declared order (so it can't itself introduce spurious diffs), but does nothing about *real* reordering in Elasticsearch, `dynamicTemplatesSemanticallyEqual` is extended to also compare the relative order of the state's declared names within the API's array — ignoring API-only extras, consistent with its existing name-based filtering — so a live reorder in Elasticsearch is no longer masked as "no diff" once `Read` reconstructs the old state order.

## Goals / Non-Goals

**Goals:**
- Make `Read`'s `dynamic_templates` projection name-aware, matching the existing `properties` behavior and the documented "user-declared subset only" contract.
- Reuse the exact name-keying already proven correct by #4581's plan-time drift-suppression path (`dynamicTemplatesByName`), so "same template" can't be judged differently by `Read`'s projection vs. `MappingsSemanticallyEqual`.
- Consolidate mapping-comparison logic into a single package (`internal/elasticsearch/index`) rather than growing a new exported seam between packages.
- Tighten test coverage so a regression here is actually caught.

**Non-Goals:**
- Changing the plan/apply consistency path's overall shape (`MappingsSemanticallyEqual`/`StringSemanticEquals`, `RequiresMappingsUpdate`) — already fixed by #4581. The one exception is `dynamicTemplatesSemanticallyEqual`, which this change extends to be order-sensitive (see Decisions) so it stays consistent with `Read`'s new order-preserving behavior; no other part of the consistency path changes.
- Changing `properties` intersection behavior — already correctly name-aware; only its package location changes (it moves alongside `intersectMappings`).
- Changing import behavior (`priorMappingsEmpty` short-circuit in `read.go:54-57`) — the first read after import stores the full API mappings unfiltered by design; this fix only changes behavior once a prior `dynamic_templates` mask exists in state.

## Decisions

### Move `intersect.go` into `internal/elasticsearch/index`

`intersect.go` and `intersect_test.go` relocate from `indexmappings` into `internal/elasticsearch/index`, alongside `MappingsValue` and `dynamicTemplatesByName`/`dynamicTemplatesSemanticallyEqual` (`mappings_value.go`). `intersectMappings` becomes exported `IntersectMappings` (the `indexmappings/read.go:73` call site changes to `index.IntersectMappings(apiMap, stateMap)`); `intersectProperties` and the new dynamic-templates helper stay unexported since everything is now one package. This was decided over exporting `dynamicTemplatesByName` from `index` and keeping `intersect.go` in `indexmappings` (smaller diff, but explicitly rejected by the maintainer decision to consolidate rather than add a public seam).

### Name-keyed `dynamic_templates` intersection

Add a `dynamic_templates` case in `IntersectMappings`, parallel to the existing `properties` case — but note it must handle the fully-absent-key case *before* the generic "API key missing" early return, not alongside the other key-specific branches (which only run once `apiVal, ok := apiMappings[key]` has already succeeded):

```go
for key, stateVal := range stateMappings {
    apiVal, ok := apiMappings[key]
    if !ok {
        if key == dynamicTemplatesKey {
            // API omitted dynamic_templates entirely (e.g. every declared template
            // was removed out-of-band) — drop rather than pin stateVal (decision 2).
            continue
        }
        result[key] = stateVal
        continue
    }
    if key == dynamicTemplatesKey {
        if templates, ok := intersectDynamicTemplates(apiVal, stateVal); ok {
            if len(templates) > 0 {
                result[key] = templates
            }
            continue
        }
        // unparseable shape (decision 3) — fall through to today's passthrough below
    }
    // ... existing propertiesKey case, then FieldSemanticallyEqual passthrough
}
```

`intersectDynamicTemplates` walks the **state's** `[]any` in its original declared order (not `dynamicTemplatesByName`'s map, which would lose that order), looks up each declared name in `dynamicTemplatesByName(apiVal)`, and keeps the API's definition for names found — mirroring `intersectProperties`'s "filter by name, keep API's definition" pattern rather than `dynamicTemplatesSemanticallyEqual`'s equality check. Names absent from the API are omitted (decision 2). If either side fails to parse via `dynamicTemplatesByName` (decision 3's duplicate-name/non-object case), the function returns `ok = false` and the caller passes through `apiVal` unfiltered, unchanged from today's behavior.

Preserving the **state's declared order** (rather than the API's array order) for kept entries is intentional: `dynamic_templates` order is semantically significant to Elasticsearch (first match wins), so re-ordering on read could silently change matching behavior even when no name changes. Because `Read` now always reconstructs the state's own order, this order-preservation is closed over by construction — the remaining gap is `dynamicTemplatesSemanticallyEqual` needing to detect *real* reordering in Elasticsearch (see next decision).

### Order-sensitive `dynamicTemplatesSemanticallyEqual`

`dynamicTemplatesSemanticallyEqual` (`mappings_value.go`) currently compares only by name and ignores order. With `Read` now preserving the state's declared order, a live reorder of the user's own declared templates in Elasticsearch would otherwise round-trip through `Read` back into the old order and `plan` would report no diff via `MappingsSemanticallyEqual` — masking a real behavior change (Elasticsearch dynamic templates are first-match-wins). `dynamicTemplatesSemanticallyEqual` is extended to additionally require that the **relative order of the user's declared names** (as they appear in `userRaw`) matches their relative order in `apiRaw`, while still ignoring API-only extras (index-template-injected names) — i.e. after filtering the API's array down to just the names that are also declared, the two name sequences must match. This is a small, targeted extension of existing #4581 logic, not a change to its overall shape.

### Test tightening

`checkStateMappingsDynamicTemplates` changes signature from `(minCount int)` to an exact-name-set form, e.g. `checkStateMappingsDynamicTemplates(wantNames ...string)`, asserting the state array contains exactly those names (order-insensitive on the assertion, but the underlying fix preserves declared order) and that `template_default` — the out-of-band-injected name used by the existing fixture — is absent. `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` gains a third step that deletes a declared template name out-of-band (via a follow-up `setDynamicTemplatesOutOfBand` call that omits it) and asserts the resulting plan reflects its removal from the declared set rather than pinning a stale value.

## Risks / Trade-offs

- [Risk] Relocating `intersect.go` touches imports in `read.go` and moves unit test coverage into the `index` package (or an `index_test` external test package) — a mechanically larger diff than a same-package fix. Mitigation: this is an explicit, already-resolved maintainer decision (consolidation over a new public seam); the relocation itself has no behavioral effect on `properties` handling.
- [Risk] The invalid-shape passthrough (decision 3) means a malformed `dynamic_templates` in state (e.g. from manual edits) still round-trips the full API array rather than being filtered — this is intentional (matches today's behavior for that edge case) but should be called out in the delta spec so it isn't mistaken for an oversight.
- [Risk] Extending `dynamicTemplatesSemanticallyEqual` to be order-sensitive touches the plan/apply consistency path that #4581 already fixed, slightly widening this change's blast radius beyond `Read`. Mitigation: the extension is additive (existing name-based equality checks are unchanged; only a new order check on the shared-name subsequence is added) and is necessary to keep `Read`'s new order-preserving behavior from masking real drift — without it, decision on declared-order preservation and existing equality semantics would contradict each other, which is exactly the inconsistency this decision closes.

## Open questions

None outstanding. `IntersectMappings` must be exported: `read.go` stays in package `indexmappings` while the intersection logic moves to package `index`, so an unexported symbol would not compile against that production call site (an internal test package cannot grant a *production* caller access to an unexported symbol). The declared-order preservation question is resolved by the order-sensitive `dynamicTemplatesSemanticallyEqual` decision above, which makes order preservation and plan-time equality consistent with each other.
