## Context

`Read` for `elasticstack_elasticsearch_index_mappings` is supposed to keep only the user-declared subset of `mappings` (REQ-004). `intersectProperties` already does this correctly for `properties` by recursing name-by-name. `intersectMappings` (`internal/elasticsearch/index/indexmappings/intersect.go:38-64`) special-cases only `properties` (line 47); every other top-level key falls through to `index.FieldSemanticallyEqual(stateVal, apiVal)` (line 57), which type-asserts to `map[string]any`. `dynamic_templates` is `[]any`, so that assertion always fails and `result[key] = apiVal` (line 61) stores the entire API array — user templates plus anything an index template injected — into state.

Two maintainer decision rounds on the issue, together with the prior implementation-research comment, fully pin down the fix:

1. **Package consolidation direction**: move `intersect.go` into `internal/elasticsearch/index` (not the reverse), so the fix reuses `dynamicTemplatesByName` (added in #4581, `mappings_value.go:427-452`) in-package rather than exporting it as new public API surface.
2. **Drop-on-missing**: a declared `dynamic_templates` name absent from the API (deleted out-of-band) is dropped from state, mirroring `intersectProperties`'s existing omit-on-missing at `intersect.go:73-76`. Drift then surfaces on the next plan instead of state silently pinning a stale value. This applies both when the API array still exists but omits the name, and when the API omits the `dynamic_templates` key entirely (e.g. every declared template was removed out-of-band) — the latter case must be handled explicitly, because `intersectMappings`'s generic "API key absent → keep `stateVal`" early return (`intersect.go:41-45`) runs before any key-specific branch and would otherwise pin the stale array.
3. **Invalid state shape** (duplicate template name, or a template value that isn't an object — only reachable via manual state edits/import, since Elasticsearch itself won't produce this shape): fall back to today's whole-array passthrough for that key, not drop the key outright.
4. **Test tightening**: `checkStateMappingsDynamicTemplates` (`acc_test.go:636-651`) moves from a `len >= minCount` assertion to an exact-name-set assertion, must additionally assert `template_default` is absent from state, and every existing call site (`TestAccResourceIndexMappings_allTopLevelKeys` at `acc_test.go:179` and `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` at `acc_test.go:210`) is updated to pass its fixture's expected declared name(s). `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` additionally gets a third step exercising decision 2 (a declared name deleted out-of-band).
5. **Order of declared names on Read**: emit declared names in **API relative order** (still drop undeclared extras; still use the API definition for each kept name). Walking state order would rewrite a live Elasticsearch reorder back to config and hide first-match-wins drift. Happy path (no live reorder) is unchanged because API relative order of declared names equals state order. `dynamicTemplatesSemanticallyEqual` remains order-sensitive on the shared-name subsequence so plan surfaces the reorder once Read has written it into state. Empty vs nonempty `dynamic_templates` is not a subset match, so a drop persisted as `[]` is not re-pinned by Framework Read semantic equality.

## Goals / Non-Goals

**Goals:**
- Make `Read`'s `dynamic_templates` projection name-aware, matching the existing `properties` behavior and the documented "user-declared subset only" contract.
- Reuse the exact name-keying already proven correct by #4581's plan-time drift-suppression path (`dynamicTemplatesByName`), so "same template" can't be judged differently by `Read`'s projection vs. `MappingsSemanticallyEqual`.
- Consolidate mapping-comparison logic into a single package (`internal/elasticsearch/index`) rather than growing a new exported seam between packages.
- Tighten test coverage so a regression here is actually caught.

**Non-Goals:**
- Changing the plan/apply consistency path's overall shape (`MappingsSemanticallyEqual`/`StringSemanticEquals`, `RequiresMappingsUpdate`) — already fixed by #4581. The one exception is `dynamicTemplatesSemanticallyEqual`, which this change extends to be order-sensitive (see Decisions) so it stays consistent with `Read`'s new order-preserving behavior; no other part of the consistency path changes.
- Changing `properties` intersection behavior — already correctly name-aware; only its package location changes (it moves alongside `intersectMappings`).
- Changing the first-read-after-import short-circuit (`priorMappingsEmpty` in `read.go`) — the first Read after import still stores the full API mappings unfiltered. Exact `dynamic_templates` name-set equality does change the **first plan** after import when that unfiltered state includes undeclared extras: the plan shows a one-off `mappings` diff, apply writes the declared subset, and later Reads intersect extras away.

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
            // API omitted dynamic_templates entirely — persist [] rather than
            // omit the key, so Framework semantic equality cannot re-pin prior.
            result[key] = []any{}
            continue
        }
        result[key] = stateVal
        continue
    }
    if key == dynamicTemplatesKey {
        if templates, ok := intersectDynamicTemplates(apiVal, stateVal); ok {
            result[key] = templates // including empty
            continue
        }
        // unparseable shape (decision 3) — fall through to today's passthrough below
    }
    // ... existing propertiesKey case, then FieldSemanticallyEqual passthrough
}
```

`intersectDynamicTemplates` walks the **API's** `[]any` in its relative order (not `dynamicTemplatesByName`'s map, which has no defined order, and not the state's array, which would hide a live reorder), keeps names that are also declared in state, and uses the API's definition for each kept name — mirroring `intersectProperties`'s "filter by name, keep API's definition" pattern. Names absent from the API are omitted (decision 2). If either side fails to parse via `dynamicTemplatesByName` (decision 3's duplicate-name/non-object case), the function returns `ok = false` and the caller passes through `apiVal` unfiltered, unchanged from today's behavior.

Emitting the **API's relative order of declared names** is required so a live reorder in Elasticsearch is written into state and then surfaces as plan drift. Walking state order would reconstruct `[alpha, beta]` from API `[beta, alpha]` and `dynamicTemplatesSemanticallyEqual` would never see the API order on this resource. Happy path (no live reorder) is unchanged because the API relative order of declared names already matches state.

### Exact `dynamic_templates` name-set equality for this resource

Shared `MappingsType{}` extras-tolerant equality (config `[alpha]` vs state `[alpha, extra]`) is required by `elasticstack_elasticsearch_index` and the template resources, which store the full API mappings. That same subset predicate cannot tell "beta was declared and disappeared" from "extra is template-injected", so Framework Read would re-pin prior `[alpha, beta]` over intersected `[alpha]`.

`elasticstack_elasticsearch_index_mappings` therefore uses `MappingsType{ExactDynamicTemplateNames: true}`. Read-constructed values must carry the same flag (`WithExactDynamicTemplateNames`) or `proposedNew.StringSemanticEquals(prior)` still uses extras-tolerant equality. When the flag is set, `dynamicTemplatesSemanticallyEqual` requires matching name sets in addition to definition and relative-order checks. Bidirectional `StringSemanticEquals` and `RequiresMappingsUpdate` are otherwise unchanged.

`dynamicTemplatesSemanticallyEqual` (`mappings_value.go`) compares by name and, after filtering the API array down to declared names (when extras are tolerated), requires that relative order to match `userRaw`. With `Read` storing the API's relative order of declared names, a live reorder is visible in state; equality then reports it as drift rather than masking it. An empty declared list is **not** a subset of a nonempty API list.

### Test tightening

`checkStateMappingsDynamicTemplates` changes signature from `(minCount int)` to an exact-name-set form, e.g. `checkStateMappingsDynamicTemplates(wantNames ...string)`, asserting the state array contains exactly those names (empty `wantNames` requires an empty array) and that `template_default` — the out-of-band-injected name used by the existing fixture — is absent. `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` gains a third step that deletes a declared template name out-of-band (via a follow-up `setDynamicTemplatesOutOfBand` call that omits it) and asserts the resulting plan reflects its removal from the declared set rather than pinning a stale value.

## Risks / Trade-offs

- [Risk] Relocating `intersect.go` touches imports in `read.go` and moves unit test coverage into the `index` package (or an `index_test` external test package) — a mechanically larger diff than a same-package fix. Mitigation: this is an explicit, already-resolved maintainer decision (consolidation over a new public seam); the relocation itself has no behavioral effect on `properties` handling.
- [Risk] The invalid-shape passthrough (decision 3) means a malformed `dynamic_templates` in state (e.g. from manual edits) still round-trips the full API array rather than being filtered — this is intentional (matches today's behavior for that edge case) but should be called out in the delta spec so it isn't mistaken for an oversight.
- [Risk] Extending `dynamicTemplatesSemanticallyEqual` to be order-sensitive (and to reject empty-vs-nonempty as a subset) touches the plan/apply consistency path that #4581 already fixed. Mitigation: bidirectional `StringSemanticEquals` is unchanged, so the index resource extras contract remains; only the shared-name subsequence order check and the empty-vs-nonempty rule are added, which are required so Read's API-order projection and drop-as-`[]` persist through Framework semantic equality.

## Open questions

None outstanding. `IntersectMappings` must be exported: `read.go` stays in package `indexmappings` while the intersection logic moves to package `index`, so an unexported symbol would not compile against that production call site (an internal test package cannot grant a *production* caller access to an unexported symbol). Declared names are emitted in API relative order so a live reorder is visible in state and then in plan.
