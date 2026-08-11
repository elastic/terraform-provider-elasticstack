# API Contract Review — fix-index-mapping-update-unidirectional-check

**Lane:** api-contract  
**Iteration:** 2 (post-implementation)  
**Status:** implementation reviewed

---

## Summary

Implementation is confirmed clean against the API contract. The single WARNING from the pre-implementation design review (W1 — argument-direction contract) is fully resolved by the doc comment on `RequiresMappingsUpdate`. No breaking changes. No new findings.

---

## W1 Resolution — Verified

**Pre-implementation requirement:** The doc comment on `RequiresMappingsUpdate` must state explicitly that the receiver is the planned mappings and the argument is the prior state.

**Implemented at:** `internal/elasticsearch/index/mappings_value.go:155–164`

The doc comment reads:

> The receiver is the planned mappings (user intent). state is the prior recorded state (API observation). MappingsSemanticallyEqual(planMap, stateMap) asks "does state already have everything plan wants?" Negating gives "plan wants something state doesn't have." Reversing the arguments to MappingsSemanticallyEqual silently reintroduces elastic/protections-cloud#19769.

This exceeds the minimum requirement: it names the reversal hazard explicitly and cites the issue number it reintroduces. W1 is closed.

---

## Confirmed Clean (no findings)

### Breaking changes

None. `RequiresMappingsUpdate` is a new exported method on `MappingsValue` in `internal/elasticsearch/index/`. Existing method signatures — `StringSemanticEquals`, `Equal`, `Type`, `ValueFromString`, `ValueFromTerraform` — are unchanged. The compile-time interface guards at `mappings_value.go:36–38` are unaffected: `RequiresMappingsUpdate` is not part of `basetypes.StringValuableWithSemanticEquals`.

### Call-site direction

`update.go:197`: `planMappings.RequiresMappingsUpdate(ctx, stateMappings)` — receiver is the plan, argument is the prior state. Correct direction.

### Type signature

`func (v MappingsValue) RequiresMappingsUpdate(ctx context.Context, state MappingsValue) (bool, diag.Diagnostics)` takes concrete `MappingsValue` on both sides. This is intentional and correct: it is strictly safer than `StringSemanticEquals`'s `basetypes.StringValuable` interface, providing compile-time argument-type checking at call sites.

### Null/unknown semantics

Documented in the function comment and verified by tests:
- Receiver null or unknown → `false` (nothing planned, no PUT needed). Tested by "plan null" and "plan unknown" cases.
- Receiver known, state null or unknown → `true` (state empty, plan adds everything). Tested by "state null with non-null plan" and "state unknown with non-null plan" cases.

The null-state→true path is the correct safe default: if the recorded state is absent, we cannot assume state covers plan.

### Test coverage for asymmetric direction

`mappings_value_test.go:422–498` (`TestIndexMappingsValue_RequiresMappingsUpdate`) includes:
- "plan adds field not in state — update required": plan=field1And2, state=field1Only → `true`
- "state is superset of plan (template-injected extras) — no update": plan=field1Only, state=field1Plus3 → `false`

These two cases together confirm asymmetric behaviour: adding a field in plan is detected (true), but extra fields in state are tolerated (false). The direction-reversal hazard is therefore covered by the test suite.

### `FieldSemanticallyEqual` and `MappingsSemanticallyEqual` exported functions

Unchanged in signature and semantics. No API contract concerns.

---

## Verdict

**approve** — W1 is fully resolved by the implementation. The API contract is sound: no breaking changes, the new surface is additive, the type design is strictly safer than the `StringSemanticEquals` pattern, the call-site direction is correct, and the doc comment explicitly documents the argument-direction invariant and the consequence of violating it.
