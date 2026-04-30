# Task 1 Coverage Review – `resource_envelope.go`

## Coverage Summary

| File | Statements | Covered | Percentage |
|------|-----------|---------|------------|
| `internal/entitycore/resource_envelope.go` | 40 | 39 | **97.5%** |

### Per-function breakdown

| Function | Coverage |
|----------|----------|
| `NewElasticsearchResource` | 100.0% |
| `Schema` | 100.0% |
| `Create` | 0.0% |
| `Read` | 100.0% |
| `Update` | 0.0% |
| `Delete` | 92.3% |
| `ImportState` | 100.0% |

> **Note:** `Create` and `Update` are intentionally left as no-ops in the envelope. Concrete resources that embed `*ElasticsearchResource[T]` override these methods. They are therefore low-risk uncovered paths for the envelope itself.

---

## Scenario Coverage Verification

All explicitly listed scenarios are covered by existing tests:

| # | Scenario | Test Function | Status |
|---|----------|---------------|--------|
| 1 | Constructor returns valid resource satisfying required interfaces | `TestNewElasticsearchResource_typeAssertions` | ✅ Covered |
| 2 | Metadata type-name composition | `TestNewElasticsearchResource_Metadata` | ✅ Covered |
| 3 | Schema injects `elasticsearch_connection` block | `TestNewElasticsearchResource_schemaInjection` | ✅ Covered |
| 4 | Schema defensive clone | `TestNewElasticsearchResource_schemaDefensiveClone` | ✅ Covered |
| 5 | Read happy path | `TestNewElasticsearchResource_Read_happyPath` | ✅ Covered |
| 6 | Read not-found removes state | `TestNewElasticsearchResource_Read_notFound` | ✅ Covered |
| 7 | Read short-circuits on `state.Get` error | `TestNewElasticsearchResource_Read_shortCircuitStateGetError` | ✅ Covered |
| 8 | Read short-circuits on composite ID parse failure | `TestNewElasticsearchResource_Read_shortCircuitCompositeIDError` | ✅ Covered |
| 9 | Read short-circuits on client resolution failure | `TestNewElasticsearchResource_Read_shortCircuitClientError` | ✅ Covered |
| 10 | Read short-circuits on `readFunc` diagnostic error | `TestNewElasticsearchResource_Read_shortCircuitReadFuncError` | ✅ Covered |
| 11 | Delete happy path | `TestNewElasticsearchResource_Delete_happyPath` | ✅ Covered |
| 12 | Delete short-circuits on composite ID parse failure | `TestNewElasticsearchResource_Delete_shortCircuitCompositeIDError` | ✅ Covered |
| 13 | Delete short-circuits on client resolution failure | `TestNewElasticsearchResource_Delete_shortCircuitClientError` | ✅ Covered |
| 14 | Default `ImportState` passthrough | `TestNewElasticsearchResource_ImportState_defaultPassthrough` | ✅ Covered |

---

## Untested High-Risk Code Paths

### 1. Delete short-circuit on `state.Get` error (1 statement, medium risk)

The `Delete` function does **not** have a test for the branch where `req.State.Get(ctx, &model)` returns an error. This leaves one `return` statement untested. If state decoding fails during a delete operation, the envelope should early-exit without calling `deleteFunc`. A targeted test using a mismatched schema (mirroring the existing `TestNewElasticsearchResource_Read_shortCircuitStateGetError`) would close this gap and bring `Delete` to 100%.

### 2. Delete `deleteFunc` diagnostic error

Although not explicitly listed in the review requirements, there is **no** test verifying behaviour when `deleteFunc` itself returns diagnostics with errors. The current `TestNewElasticsearchResource_Delete_happyPath` only covers the nil-error path. However, the coverage tool shows the statement at the end of `Delete` is already hit (because the `resp.Diagnostics.Append(...)` line executes regardless), so there is no uncovered line per se. A test asserting that diagnostics are propagated would still improve confidence.

### 3. `Create` and `Update` no-ops (intentionally uncovered)

These are placeholder methods. The envelope contract explicitly states that concrete resources override them. Risk is low.

---

## Package Coverage Context

| Metric | Value |
|--------|-------|
| Package (`internal/entitycore`) | 74.6% |
| `resource_envelope.go` | 97.5% |

The package figure is pulled down by `data_source_envelope.go` (`Read` at 53.3%, and several data-source-specific `Configure` / `Metadata` / `Schema` paths at 0%), which is outside the scope of Task 1.

---

## Recommendations

1. **Add `TestNewElasticsearchResource_Delete_shortCircuitStateGetError`** to cover the missing `return` in `Delete`, bringing `resource_envelope.go` to 100%.
2. Consider adding a `Delete` error-propagation test for `deleteFunc` diagnostics to guard against future regressions.
3. No action required for `Create` / `Update` no-ops (by design).
