## Context

`elasticstack_elasticsearch_cluster_settings` implements `resource.ResourceWithValidateConfig` via
`ValidateConfig` → `validateConfigModel` → `categoryBlockEmpty`. The validation fires during
Terraform's `ValidateResourceConfig` RPC, which runs **before** local values are evaluated. When
`persistent` or `transient` is populated by a `dynamic` block whose `for_each` expression
references a local, the block is represented as an **unknown** value at this stage.

The current `categoryBlockEmpty` implementation:

```go
func categoryBlockEmpty(block types.Object) bool {
    if block.IsNull() || block.IsUnknown() {  // ← treats unknown as empty
        return true
    }
    ...
    return settingSet.IsNull() || settingSet.IsUnknown() || ...  // ← same issue
}
```

Because `block.IsUnknown()` is `true` for a dynamic block driven by un-evaluated locals, both
blocks appear "empty", and the error fires even though the configuration is semantically valid.

This is a targeted, semantics-based bug: `null` (block not configured) ≠ `unknown` (block present
but not yet evaluable). The Plugin Framework's convention — followed by
`terraform-plugin-framework-validators` and the project's own `settingNameUniqueValidator` — is
to skip validation (emit no error) when values are unknown.

## Goals / Non-Goals

**Goals:**
- Eliminate the false-positive "No cluster settings configured" error for `dynamic`-driven blocks.
- Align `categoryBlockEmpty` with the Plugin Framework unknown-value convention.
- Provide focused unit tests that catch any future regression.

**Non-Goals:**
- Changing API interaction logic (`expandSettings`, `flattenSettings`, `getConfiguredSettings`).
- Modifying the state upgrade path (`UpgradeState` / `migrateClusterSettingsStateV0ToV1`).
- Altering the `settingNameUniqueValidator` (already handles unknowns correctly).
- Adding acceptance-test coverage for the `dynamic`/`for_each` pattern (unit tests suffice).

## Decisions

### 1. Fix `categoryBlockEmpty` directly (Approach A)

Change `categoryBlockEmpty` to return `false` when the block or the setting set is unknown,
instead of returning `true`. This is the minimal, most targeted fix.

```go
func categoryBlockEmpty(block types.Object) bool {
    if block.IsNull() {
        return true
    }
    if block.IsUnknown() {
        return false  // unknown ≠ absent; defer validation to plan time
    }
    settingAttr, ok := block.Attributes()["setting"]
    if !ok {
        return true
    }
    settingSet, ok := settingAttr.(types.Set)
    if !ok {
        return true
    }
    if settingSet.IsUnknown() {
        return false  // same principle for the nested set
    }
    return settingSet.IsNull() || len(settingSet.Elements()) == 0
}
```

**Why this approach:**
- Smallest blast radius: one function, three lines changed.
- Semantically precise: only `null` means "not configured"; `unknown` means "not yet evaluated".
- Handles both the outer-block unknown case AND the nested-set unknown case (Approach B — adding a
  guard at the top of `validateConfigModel` — only handles the outer case; if the outer block is
  known but the inner set is unknown, the bug could resurface).
- All existing unit tests (`BothNull_Error`, `BothEmpty_Error`, `PersistentSet_OK`,
  `TransientSet_OK`) continue to pass unchanged.

**Approaches not chosen:**
- **Approach B** (guard at top of `validateConfigModel`): handled at a higher level but misses
  the nested-set unknown case.
- **Approach C** (move check to `ModifyPlan`): over-engineering; `ModifyPlan` is for computing
  values, not for emitting validation errors. Delays the error for genuinely empty static
  configs unnecessarily.

### 2. Expose `categoryBlockEmpty` via `export_test.go` for direct unit testing

Add an `ExportedCategoryBlockEmpty` helper in `export_test.go` so the helper can be tested
directly in `helpers_test.go` without relying solely on end-to-end `ExportedValidateConfigModel`
calls. This follows the existing export pattern in the package.

### 3. Three targeted unit tests

| Test name | Input | Expected |
|---|---|---|
| `TestValidateConfigModel_BothUnknown_OK` | both blocks unknown | no error |
| `TestValidateConfigModel_OneUnknown_OK` | one block null, one unknown | no error |
| `TestCategoryBlockEmpty_Unknown_NotEmpty` | unknown block | returns `false` |

These complement the four existing tests without modifying them.

## Risks / Trade-offs

- **Silent deferral for unknown blocks**: when both blocks are unknown, no error is emitted at
  validate time. The cross-block "at least one of `persistent` or `transient` must be non-empty"
  invariant is still enforced by `validateConfigModel` once values are known (i.e., at plan time
  when the unknown blocks have been resolved). `setvalidator.SizeAtLeast(1)` is attached only to
  the nested `setting` set and cannot enforce the cross-block invariant. This is correct behavior
  and is how the Plugin Framework expects validators to behave with unknown values.
- **`setvalidator.SizeAtLeast(1)` false positives**: the standard validator from
  `terraform-plugin-framework-validators` already short-circuits on unknown values, so this is
  not an issue. Confirmed by checking validators usage in `schema.go` and the framework's
  documented behavior.

## Open Questions

1. Does `setvalidator.SizeAtLeast(1)` on the nested `setting` block also produce false positives
   for unknown values? The standard validator already skips unknowns, so this is likely not an
   issue — but the implementor should confirm by checking the version pinned in `go.mod`.

2. Should Approach A also guard against the edge case where the outer block is
   known-non-null but the inner `setting` set is unknown (e.g., a static `persistent {}` block
   containing a `dynamic "setting"` block)? **Yes — Approach A already handles this case** because
   it checks `settingSet.IsUnknown()` separately. The implementor should add a targeted test for
   this pattern if it is considered a realistic user configuration.

## Migration Plan

1. Edit `categoryBlockEmpty` in `resource.go` as described in Decision 1.
2. Add `ExportedCategoryBlockEmpty` to `export_test.go`.
3. Add three unit tests to `helpers_test.go`.
4. Run `go test ./internal/elasticsearch/cluster/settings/...` to verify all tests pass.
