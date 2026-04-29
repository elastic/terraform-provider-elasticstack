# Task 2 Compliance Review: remove-7x-support — Makefile Fleet image fallback

## Key Checks

| Check | Expected | Actual | Status |
|-------|----------|--------|--------|
| Makefile matches only `8.0.%` and `8.1.%` | Filter list = `8.0.% 8.1.%` | `filter 8.0.% 8.1.%` | ✅ Pass |
| `7.17.%` no longer matched | `7.17.%` absent from filter | `7.17.%` not present | ✅ Pass |
| Comment no longer mentions 7.17 | Comment text omits 7.17 | "notably 8.0.x, 8.1.x" | ✅ Pass |
| Live OpenSpec specs reflect behavior | Spec matches implementation and delta spec | Live spec still mentions 7.17 | ❌ Mismatch |

---

## Issues

### WARNING — Live OpenSpec spec out of sync with implementation

**Severity:** WARNING  
**Finding:** The live canonical spec `openspec/specs/makefile-workflows/spec.md` still documents the old 7.17 fallback behavior, but the implementation (and the change's delta spec) have removed it.

**Evidence:**
- `Makefile:34` (on `remove-7.x`): `ifneq (,$(filter 8.0.% 8.1.%,$(STACK_VERSION)))`
- `openspec/specs/makefile-workflows/spec.md:97`: `When STACK_VERSION matches 7.17.%, 8.0.%, or 8.1.%...`
- `openspec/changes/remove-7x-support/specs/makefile-workflows/spec.md:5`: `When STACK_VERSION matches 8.0.% or 8.1.%...` (correct delta spec)

**Recommended fix:** Update the live spec `openspec/specs/makefile-workflows/spec.md` Requirement REQ-017 and its scenario to match the delta spec wording:
- Remove `7.17.%` from the version list in the requirement text.
- Rename the scenario heading from "Older 7.17 / 8.0 / 8.1 line" to "Older 8.0 / 8.1 line".
- Remove `7.17.%` from the scenario GIVEN clause.

This also satisfies task 2.2: "Update comments and current OpenSpec wording for Makefile workflow behavior so older-version fallback language no longer mentions 7.17."

---

## Summary

- Implementation of Task 2.1 (Makefile logic) is complete and correct.
- Makefile comment update for Task 2.2 is complete.
- **Remaining work:** Apply the same 7.17 removal to the live canonical OpenSpec spec (`openspec/specs/makefile-workflows/spec.md`).
- No other Task 2-related warnings or critical issues found.
