# Standards Review — fix-index-mapping-update-unidirectional-check

**Lane:** standards  
**Iteration:** 2 (implementation present)  
**Status:** complete

---

## Lint / Format

- `gofmt -l` on changed files — **pass** (no output; all files correctly formatted)
- `go vet ./internal/elasticsearch/index/...` — **pass** (no output)
- `go build ./...` — **pass** (exit 0)
- `go test ./internal/elasticsearch/index/...` — **pass** (446 passed; 4 skipped due to missing `ELASTICSEARCH_ENDPOINTS` env var, which is expected in local dev)

---

## File Changes Map (design.md vs actual commit)

| File | Expected | Actual | Status |
|---|---|---|---|
| `internal/elasticsearch/index/mappings_value.go` | Add `RequiresMappingsUpdate` | ✓ added at line 167 | resolved |
| `internal/elasticsearch/index/mappings_value_test.go` | Add `TestIndexMappingsValue_RequiresMappingsUpdate` | ✓ added at line 422 (name convention correct) | resolved |
| `internal/elasticsearch/index/index/update.go` | Replace `StringSemanticEquals` with `RequiresMappingsUpdate` | ✓ done at line 197 | resolved |
| `internal/elasticsearch/index/index/create.go` | No code change (fix is transitive) | ✓ confirmed: `adoptExistingIndexOnCreate` calls `updateMappings` at line 183 | resolved |
| `internal/elasticsearch/index/index/acc_test.go` | Regression + adoption acceptance tests | ✓ both present | resolved |
| testdata/TestAccResourceIndexMappingsUpdateRegression | create + update step dirs | ✓ both present with correct .tf configs | resolved |
| testdata/TestAccResourceIndexUseExistingAdoptMappings | create step dir | ✓ present with use_existing=true config | resolved |
| `openspec/specs/elasticsearch-index/spec.md` | Spec sync (REQ-015–REQ-018 + adoption) | ✓ both updated | resolved |
| `.gitignore` | **not in design** | present in commit | see W1 |

Previous pre-implementation WARNINGs (W1 test name, W2 `typeutils.IsKnown`): both **resolved** in implementation.

---

## Findings

### WARNING — W1: `.gitignore` change is unrelated to this fix

**Location:** `.gitignore` (new lines 70–79)

**Description:** The commit adds ten lines to `.gitignore` for Gas City / Beads tooling (`*.db`, `.beads/`, `.gc/`, `.beads-credential-key`, etc.). These entries are unrelated to the `RequiresMappingsUpdate` fix and do not appear in the design's File Changes map. Including tooling scaffolding in a feature commit obscures the change surface during review and may conflict with other branches that also need these entries.

**Required fix:** Extract the `.gitignore` additions into a separate commit or PR targeting the main branch.

---

### SUGGESTION — S1: Inline comment at `mappings_value.go:183` is redundant

**Location:** `internal/elasticsearch/index/mappings_value.go:183`

```go
// planMap is receiver (plan intent), stateMap is argument (API state). Reversing reintroduces #19769.
return !MappingsSemanticallyEqual(planMap, stateMap), diags
```

**Description:** The method doc comment (lines 155–165) already documents the argument-order invariant in full, including the bug reference. The inline comment repeats the same information and adds an external issue number that will be opaque to contributors without access to the referenced tracker. Removing or trimming it would not reduce understanding.

**Suggested:** drop the inline comment, or replace it with `// argument order: plan first, state second — see doc comment above.` if the proximity is considered useful.

---

## Previous WARNINGs (resolved)

- **W1 (iteration 1):** Test name `TestMappingsValue_RequiresMappingsUpdate` — **resolved**: implementation uses `TestIndexMappingsValue_RequiresMappingsUpdate`, matching the `TestIndexMappingsValue_*` convention.
- **W2 (iteration 1):** Use `typeutils.IsKnown` — **resolved**: implementation uses `typeutils.IsKnown(v)` and `typeutils.IsKnown(state)` at lines 170 and 173.

---

## Verdict

`approve`

The implementation is correct, idiomatic, and conforms to project standards. Lint, format, vet, and build all pass. Both previous `iterate` findings are resolved. One WARNING (`.gitignore`) and one Suggestion (inline comment) are noted but neither blocks acceptance of the fix itself.
