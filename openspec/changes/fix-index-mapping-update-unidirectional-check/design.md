## Context

`elasticstack_elasticsearch_index` stores `mappings` as a `MappingsValue` (a normalized JSON
string type, `internal/elasticsearch/index/mappings_value.go`). Two independent concerns currently
share the same predicate, `MappingsSemanticallyEqual`, via `StringSemanticEquals`:

1. **Plan-time semantic equality / replace decision** (REQ-022–REQ-024): tolerate
   Elasticsearch/template-injected extras (extra `properties`, `dynamic_templates`, `_meta`,
   `semantic_text` `model_settings`, …) so they don't show up as drift or a forced replace. This is
   legitimately bidirectional-tolerant in intent — either side may carry template/API-injected
   extras the other doesn't know about — and is implemented as
   `MappingsSemanticallyEqual(vMap, newMap) || MappingsSemanticallyEqual(newMap, vMap)`.
2. **Apply-time update decision**: whether to issue `PUT /{index}/_mapping`. `update.go:194`
   (`updateMappings`) and `create.go:183` (`adoptExistingIndexOnCreate`, the `use_existing`
   adoption path) both reuse `StringSemanticEquals` for this. Because that check is bidirectional,
   "plan is a superset of state" (the user added a field) reads as "equal", and the Put Mapping
   API is skipped — silently dropping the addition. This is the reported bug
   (`elastic/protections-cloud#19769`).

The two concerns need different directionality. Plan-time equality legitimately tolerates
extras on *either* side. The apply-time decision must only ask: "does the plan contain anything
state doesn't already have?" — a one-way question.

## Goals / Non-Goals

**Goals:**

- Ensure `PUT /{index}/_mapping` is called whenever the planned config's `mappings` contains
  content the prior state does not already have (field additions, changed content).
- Preserve the existing tolerance for Elasticsearch/template-injected extras so `use_existing`
  adoption and steady-state applies do not fire redundant Put Mapping calls when state already
  covers everything the plan specifies.
- Keep the fix scoped to the update-decision call sites; do not change plan-time semantic equality
  or replacement semantics (REQ-022–REQ-024).

**Non-Goals:**

- Changing `MappingsSemanticallyEqual`'s bidirectional Read/drift-suppression semantics.
- Changing `elasticstack_elasticsearch_index_mappings` (`indexmappings/intersect.go`), which already
  avoids this bug class via read-side intersection.
- Adding an equivalent "skip redundant write" optimization to
  `elasticstack_elasticsearch_index_template` or `elasticstack_elasticsearch_index_component_template`
  — investigated, confirmed unnecessary (see Decisions).
- Changing data stream mapping paths.

## Decisions

### Add a unidirectional `RequiresMappingsUpdate` method on `MappingsValue`

```go
// RequiresMappingsUpdate returns true when the receiver (the planned mappings)
// contains content not already present in state — i.e. state is NOT a
// non-drifting superset of plan. It is the mirror image of the superset check
// used by StringSemanticEquals, kept as a separate method because the two
// callers need opposite directionality: StringSemanticEquals tolerates extras
// on either side for plan-time drift suppression, while an apply-time update
// decision must fire whenever the plan asks for something state doesn't have.
func (v MappingsValue) RequiresMappingsUpdate(ctx context.Context, state MappingsValue) (bool, diag.Diagnostics)
```

Implementation shape: decode both sides to `map[string]any` (mirroring
`StringSemanticEquals`'s existing null/unknown/decode handling), then

```go
return !MappingsSemanticallyEqual(planMap, stateMap), diags
```

`MappingsSemanticallyEqual(userMappings, apiMappings)` returns true when `apiMappings` is a
non-drifting superset of `userMappings`. Calling it with `userMappings = planMap`,
`apiMappings = stateMap` asks "does state already have everything plan wants?" — true means state
covers the plan (no write needed); negating it gives "the plan wants something state doesn't have"
(write needed). This is the opposite argument order from the bidirectional `StringSemanticEquals`
check and is the crux of the fix — care is needed here since a swapped argument order would
silently reintroduce the original bug.

Null/unknown handling, matching `StringSemanticEquals`'s existing guards:

- Plan null or unknown → `false` (nothing planned to add).
- State null or unknown, plan non-null/known → `true` (state has nothing, plan adds everything).
- State null or unknown, plan also null/unknown → `false`.

This is a method on `MappingsValue` (not a free function taking decoded maps), per direction from
`@tobio` on the issue — it mirrors `StringSemanticEquals`'s existing shape and keeps JSON
decoding/error-diagnostic handling colocated with the type.

Call sites: `update.go:194`'s `updateMappings` and, transitively, `create.go:183`'s
`adoptExistingIndexOnCreate` (both funnel through the same `updateMappings` helper) switch from
`planMappings.StringSemanticEquals(ctx, stateMappings)` to
`planMappings.RequiresMappingsUpdate(ctx, stateMappings)`, inverting the resulting branch (the old
code returned early when `areEqual`; the new code returns early when `!requiresUpdate`).

**Alternative considered (rejected):** drop superset tolerance entirely and gate the Put Mapping
call on normalized inequality (`reflect.DeepEqual` after `normalizeMappings`), relying on the Put
Mapping API's additive/merge semantics to make redundant calls harmless. Rejected because it
regresses `use_existing` adoption: a partial declared mapping compared against a fuller live
mapping would differ on every apply, firing a needless `PUT` every time, and would require routing
the plan modifier's merged output through `normalizeMappings` first to avoid formatting-only
deltas falsely triggering updates. The unidirectional method keeps today's tolerance intact and is
a smaller, more targeted diff.

### No change needed for index templates / component templates

Investigated whether `elasticstack_elasticsearch_index_template` and
`elasticstack_elasticsearch_index_component_template`, which reuse the same `MappingsType`/
`MappingsSemanticallyEqual` machinery, have an analogous bug. They do not: both
`template.writeIndexTemplate` (`template/write.go:34-68`) and `componenttemplate`'s create/update
(`componenttemplate/create.go:44`) call `PutIndexTemplate`/`PutComponentTemplate`
**unconditionally** on every Create/Update. They embed `MappingsType` only for plan-time diff
suppression, never as an apply-time "skip the write" gate, so there is no equivalent optimization
to get wrong. No spec or code change is proposed for those resources.

## Risks / Trade-offs

- **Two similarly-shaped, oppositely-directioned predicates now exist on `MappingsValue`**
  (`StringSemanticEquals` and `RequiresMappingsUpdate`) → mitigated with an explicit doc comment on
  `RequiresMappingsUpdate` calling out the argument-order contrast, and a unit test asserting both
  directions independently so a future edit that accidentally swaps the argument order fails CI.
- **Adoption asymmetry**: during `use_existing` adoption, fields present in the live mapping but
  omitted from config are kept (not deleted — the Put Mapping API cannot remove fields), while
  fields added in config that are absent from the live mapping are now written. The acceptance
  test in this change exercises both halves of that asymmetry in one scenario.

## Open questions

- ~~Method vs. free function?~~ **Resolved (`@tobio`):** method on `MappingsValue`.
- ~~Stale-state risk / need for tests?~~ **Resolved (`@tobio`):** add both the acceptance test and
  a unit test for `RequiresMappingsUpdate`. Non-blocking detail: natural homes are
  `mappings_value_test.go` (unit) and `index/acc_test.go` (acceptance) — neither currently has an
  "add field to existing index" case.
- ~~Does `elasticstack_elasticsearch_index_template` have an analogous bug?~~ **Investigated — no.**
  Both `template.writeIndexTemplate` (`template/write.go:34-68`) and `componenttemplate`'s
  create/update (`componenttemplate/create.go:44`) call `PutIndexTemplate`/`PutComponentTemplate`
  **unconditionally** on every Create/Update. They embed `MappingsType` only for plan-time diff
  suppression, never as an apply-time "skip the write" gate — so there's no equivalent optimization
  to get wrong. This class of bug is specific to the plain index resource's redundant-`PUT`
  avoidance.
- Should the acceptance test also cover `use_existing`/Create adoption (`create.go:182-187`) with a
  config that both adds and omits fields vs. the live mapping, to pin the asymmetry Approach A
  depends on? Not yet scoped. **Resolved (`@tobio`):** yes — included in this change's task list
  and proposal scope.
