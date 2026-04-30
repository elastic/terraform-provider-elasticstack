# Task 1 Spec Review — `elasticsearch-resource-envelope`

**Scope:** Subtasks 1.1 – 1.8 only (envelope implementation + tests + docs).
**Date:** 2026-04-30

## Executive Summary

Task 1 is **fully implemented and correct**. The code in `internal/entitycore/resource_envelope.go` satisfies every requirement and scenario in the canonical spec (`openspec/changes/elasticsearch-resource-envelope/specs/entitycore-resource-envelope/spec.md`). All 14 envelope unit tests pass. `make build` and `make check-openspec` pass.

One minor **test gap** was found (missing coverage for "Delete function error is appended to response diagnostics"). It does not affect correctness.

---

## Subtask-by-Subtask Verification

| Subtask | Status | Notes |
|---------|--------|-------|
| **1.1** Constructor + `ElasticsearchResourceModel` constraint + struct shape | ✅ Complete | `NewElasticsearchResource[T]`, `ElasticsearchResource[T]`, and `ElasticsearchResourceModel` are all present. |
| **1.2** `Schema` with `elasticsearch_connection` block injection | ✅ Complete | Maps copy + injection in `Schema`. Defensive clone verified by `TestNewElasticsearchResource_schemaDefensiveClone`. |
| **1.3** `Read` prelude (state decode → composite ID parse → client resolve → readFunc → Set/RemoveResource) | ✅ Complete | All 5 spec scenarios covered by tests. |
| **1.4** `Delete` prelude (same gates → deleteFunc → append diagnostics) | ✅ Complete | Happy path + composite-ID gate + client-resolution gate tested. |
| **1.5** Default `ImportState` passthrough on `id` | ✅ Complete | `resource.ImportStatePassthroughID` used. Tested. |
| **1.6** Interface assertions (`resource.Resource`, `ResourceWithConfigure`, `ResourceWithImportState`) | ✅ Complete | Compile-time assertions at bottom of `resource_envelope.go`. |
| **1.7** Update `doc.go` | ✅ Complete | Package doc covers envelope patterns, callback signatures, model constraint, default ImportState, and example usage. |
| **1.8** Unit test coverage | ✅ Complete (see Gap below) | 14 tests; all pass. |

---

## Spec ↔ Implementation Comparison

### Requirements fully matched

1. **Constructor returns valid resource** — Verified by `TestNewElasticsearchResource_typeAssertions`.
2. **Configure stores provider client factory / leaves prior on failure** — Inherited from embedded `*ResourceBase`. Behavior is tested in `configure_test.go` (`invalid_provider_data_leaves_prior_client`, `typed_nil_factory_pointer_leaves_prior_client`).
3. **Metadata builds type name** — Verified by `TestNewElasticsearchResource_Metadata` (`elasticstack_elasticsearch_test_entity`).
4. **Schema injects connection block** — Verified by `TestNewElasticsearchResource_schemaInjection`.
5. **Read happy path → `State.Set`** — Verified by `TestNewElasticsearchResource_Read_happyPath`.
6. **Read not-found → `State.RemoveResource`** — Verified by `TestNewElasticsearchResource_Read_notFound`.
7. **Read error → diagnostics appended, no state mutation** — Verified by `TestNewElasticsearchResource_Read_shortCircuitReadFuncError`.
8. **Read composite-ID parse failure → short-circuit** — Verified by `TestNewElasticsearchResource_Read_shortCircuitCompositeIDError`.
9. **Read client resolution failure → short-circuit** — Verified by `TestNewElasticsearchResource_Read_shortCircuitClientError`.
10. **Delete happy path → nil diagnostics** — Verified by `TestNewElasticsearchResource_Delete_happyPath`.
11. **Delete composite-ID parse failure → short-circuit** — Verified by `TestNewElasticsearchResource_Delete_shortCircuitCompositeIDError`.
12. **Delete client resolution failure → short-circuit** — Verified by `TestNewElasticsearchResource_Delete_shortCircuitClientError`.
13. **Default ImportState passthrough on `id`** — Verified by `TestNewElasticsearchResource_ImportState_defaultPassthrough`.
14. **Model constraint (`GetID`, `GetElasticsearchConnection`)** — Implemented as `ElasticsearchResourceModel` interface. Test model uses value receivers.
15. **Type-name preservation** — Delegates to `ResourceBase.Metadata`, which is unchanged. Verified by existing `ResourceBase` tests for component + name composition.
16. **Coexistence with `ResourceBase`-only entities** — Confirmed: the four security resources still embed `*ResourceBase` directly (migration not yet started).

### Gap: Missing test for delete callback error diagnostics

**Spec scenario:** *"Delete function error is appended to response diagnostics"*  
**Current state:** No test exercises a `deleteFunc` that returns error-level diagnostics and asserts they appear in `resp.Diagnostics` (while state remains untouched).  
**Impact:** Low. The implementation is obviously correct (`resp.Diagnostics.Append(r.deleteFunc(...)...)`, identical pattern to Read).  
**Action:** Add a `TestNewElasticsearchResource_Delete_deleteFuncError` test when convenient (non-blocking).

---

## Notable Warnings / Observations

1. **Delete callback nil enforcement** — The spec says "The system SHALL require concrete resources to supply a non-nil delete callback." The constructor signature enforces a value must be supplied at the type level; there is no runtime nil guard, and the envelope does **not** special-case nil (it will panic if called). This matches the spec’s "SHALL NOT special-case a nil callback" clause.
2. **No-op Create/Update** — The envelope defines empty `Create` and `Update` methods so that the compile-time interface assertions for `resource.Resource` succeed. Concrete resources are expected to override these. This is a necessary Go mechanism and is not mentioned as a concern in the spec.
3. **Migrations pending** — Tasks 2–5 (migrating `user`, `systemuser`, `role`, `rolemapping`) have not started. All four security packages still embed `*ResourceBase`. This is expected because the review scope is Task 1 only.

---

## Validation Results

```
go test ./internal/entitycore/... -run TestNewElasticsearchResource
  => 14/14 PASS

make build
  => PASS (no issues)

make check-openspec
  => 140 passed, 0 failed

npx openspec validate --changes --strict
  => 7 passed, 0 failed
```

---

## Recommended Next Step

Proceed to **Task 2** (`Migrate elasticsearch_security_user`). The envelope substrate is stable and spec-compliant.
