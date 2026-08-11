# Performance Review — fix-index-mapping-update-unidirectional-check

**Lane:** performance  
**Iteration:** 1  
**Status:** implementation review

---

## Summary

The change is narrowly scoped: a new `RequiresMappingsUpdate` method on `MappingsValue`, a shared `decodeMappingPair` helper, and a single call-site swap in `updateMappings`. No hot paths, N+1 queries, or blocking I/O are introduced. The implementation is a marginal performance improvement over the prior code.

---

## Analysis

### `RequiresMappingsUpdate` vs. the replaced `StringSemanticEquals`

`update.go:197` previously called `planMappings.StringSemanticEquals(ctx, stateMappings)`, which under the hood ran:

```go
MappingsSemanticallyEqual(vMap, newMap) || MappingsSemanticallyEqual(newMap, vMap)
```

In the steady-state (no-update) case the first call returns `true` and the short-circuit fires — one traversal. In the update-required case both calls run — two traversals.

`RequiresMappingsUpdate` (`mappings_value.go:167–185`) evaluates exactly one call:

```go
return !MappingsSemanticallyEqual(planMap, stateMap), diags
```

This is never worse than the prior code and is faster in the update-required case (one traversal instead of two).

### `decodeMappingPair` shared helper

The implementation extracted JSON decoding into a shared `decodeMappingPair` helper (`mappings_value.go:190–202`), used by both `StringSemanticEquals` and `RequiresMappingsUpdate`. This eliminates the copy-paste decode logic without any performance impact — the two methods are called in different call stacks (plan time vs. apply time), so no redundant decode is possible within a single call.

### Null/unknown early exits

`RequiresMappingsUpdate` short-circuits at lines 170–175 without touching JSON decode when plan or state is null/unknown. This avoids `json.Unmarshal` entirely for the no-op path, which is cheaper than the previous `StringSemanticEquals` null guard (`v.Normalized.Equal`).

### JSON decode cost

Both methods decode two JSON strings per invocation via `json.Unmarshal`. For `updateMappings` this occurs at most once per `terraform apply` per resource, not in any loop. The decode cost is O(n) in mappings size and entirely normal for an IaC apply.

Strings are already normalized at construction (`NewMappingsValue` applies `normalizeMappings`), so no second normalization pass is performed inside `RequiresMappingsUpdate`.

### `MappingsSemanticallyEqual` complexity

The traversal is O(p × d) where p is the number of properties and d is the nesting depth — linear in user-authored mapping size. For any realistic Elasticsearch index schema this is negligible and bounded.

### Allocation profile

Unchanged: two `json.Unmarshal` calls produce `map[string]any` trees that are GC'd after the call returns. No persistent allocations; no cache.

### No N+1, blocking I/O, or allocation in a loop

`updateMappings` is called once per apply per resource. The downstream `elasticsearch.UpdateIndexMappings` call is conditional on the new check returning `true`; the fix makes the condition correct, not more frequent in steady state. No new API calls per unit of work.

---

## Findings

None. No CRITICAL or WARNING items.

---

## Verdict

`approve` — No performance regressions. The implemented change is a marginal improvement over the prior code in the update-required path (one `MappingsSemanticallyEqual` traversal instead of two), with a minor added benefit of cheaper null/unknown short-circuiting. All costs remain bounded to once-per-apply-per-resource.
