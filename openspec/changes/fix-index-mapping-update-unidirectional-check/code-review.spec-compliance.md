---
schema: gc.build.code-review.v1
lane: spec-compliance
change_id: fix-index-mapping-update-unidirectional-check
verdict: approve
---

# Spec-Compliance Review

## Requirement: Update flow (REQ-015–REQ-018)

### Unidirectional decision — PASS

`RequiresMappingsUpdate` (`internal/elasticsearch/index/mappings_value.go:167`) implements the
spec's unidirectional gate exactly:

- Returns `false` when plan is null/unknown (nothing planned to add).
- Returns `true` when state is null/unknown with a known plan (state has nothing, plan adds
  everything).
- Returns `!MappingsSemanticallyEqual(planMap, stateMap)` — asks "does state already have
  everything plan wants?"; negating gives "plan adds something state lacks."

The bidirectional form `MappingsSemanticallyEqual(vMap, newMap) || MappingsSemanticallyEqual(newMap, vMap)`
in `StringSemanticEquals` (`mappings_value.go:152`) is untouched and continues to back plan-time
semantic equality and `RequiresReplace`. The update-decision gate (`update.go:197`) now calls
`RequiresMappingsUpdate`, not `StringSemanticEquals`. Both call sites (update and adoption) go
through the shared `updateMappings` helper.

No non-conformance.

### Scenario: Adding a mapping field calls the Put Mapping API — PASS

`TestAccResourceIndexMappingsUpdateRegression` (`acc_test.go:1093`) creates the index with
`field1`, applies a config adding `field2`, then calls `checkIndexHasField` (`acc_test.go:1168`)
to verify `field2` is present in the live cluster via a direct Elasticsearch API read — independent
of Terraform state. This is the correct regression-gate for the original bug.

### Scenario: Template-injected mappings do not cause mapping update — PASS

`TestIndexMappingsValue_RequiresMappingsUpdate` case "state is superset of plan
(template-injected extras) — no update" (`mappings_value_test.go:447–452`) covers this at the
unit level. No new acceptance test was added for this scenario, but `MappingsSemanticallyEqual`'s
superset semantics are separately tested in
`TestMappingsSemanticallyEqual_coverageForMappingSupersetAndDrift` and the
pre-existing `TestAccResourceIndexTemplateNoMappingDrift` suite continues to exercise the
end-to-end path.

### Scenario: State already covering plan skips the Put Mapping API — PASS

Same unit-test case as above. No acceptance-level test for the negative side (PUT not called),
which is expected — verifying a call was not made requires mock instrumentation that is outside
scope here.

---

## Requirement: Opt-in adoption via `use_existing`

### Unidirectional mapping reconciliation during adoption — PASS

`adoptExistingIndexOnCreate` (`create.go:182–187`) calls `r.updateMappings(ctx, client,
concreteName, plan.Mappings, synthetic.Mappings)`. `synthetic.Mappings` holds the existing
index's mappings (populated via `populateFromAPI`). `RequiresMappingsUpdate` therefore sees plan
vs. existing — exactly the spec's intent.

### Scenario: Adoption writes a field the plan adds and retains a field the plan omits — PASS

`TestAccResourceIndexUseExistingAdoptMappings` (`acc_test.go:1130`) creates an out-of-band index
with `legacy_field`, then adopts it with a config that specifies only `new_field`. The test
asserts:
- `checkIndexHasField(..., "new_field")` — PUT Mapping was called and succeeded.
- `checkIndexHasField(..., "legacy_field")` — retained, not deleted.

Both assertions use direct ES API reads.

### `sort.missing` and `sort.mode` in adoption static-setting comparison — PASS (code correct)

`compareStaticPlanAndES` (`use_existing.go:216–230`) handles `SettingSortMissing` and
`SettingSortMode` with the same ordered-slice comparison used for `SettingSortOrder`. Both keys
are present in `StaticSettingsKeys` (`settings_keys.go:67–68`). The code satisfies the spec
requirement that these settings be compared and an error diagnostic returned on mismatch.

### Scenario: Adoption compares `sort.missing` against existing index — WARNING

**WARNING** `use_existing_test.go` contains `Test_compareStaticSettings_sortOrder_orderedEquality`
and `Test_compareStaticSettings_sortField_mismatch` but has no corresponding unit test for
`SettingSortMissing` or `SettingSortMode` mismatch detection. The spec explicitly scenarios a
`sort.missing` mismatch returning an error diagnostic.

The code path is identical in structure to `sort.order` (already tested), and the implementation
pre-dates this change (no code was modified for `sort.missing`/`sort.mode`). The behavioral gap
is therefore low-risk, but the scenario specified in this change's spec delta has no test
anchoring it.

Suggested remediation: add a `Test_compareStaticSettings_sortMissing_mismatch` case in
`use_existing_test.go` following the pattern of `Test_compareStaticSettings_sortOrder_orderedEquality`
(`use_existing_test.go:316`).

---

## Summary

| Check | Result |
|---|---|
| `RequiresMappingsUpdate` unidirectional logic | PASS |
| `updateMappings` call site uses `RequiresMappingsUpdate` | PASS |
| `StringSemanticEquals` unchanged for plan-time equality | PASS |
| Adoption path uses shared `updateMappings` | PASS |
| Scenario: adding a field calls PUT (regression test) | PASS |
| Scenario: template-injected extras skip PUT (unit test) | PASS |
| Scenario: adoption writes new field, retains legacy field | PASS |
| `sort.missing`/`sort.mode` in static comparison code | PASS |
| Scenario: `sort.missing` mismatch has a test | **WARNING** |

All SHALL requirements in the spec delta are satisfied by the implementation. The single WARNING
is a missing unit test for a pre-existing code path that the spec delta now formally scenarios.
It does not block the change.

**Verdict: approve**
