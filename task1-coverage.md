# Task 1 Coverage/Test Analysis Report

**Branch:** `remove-7x-support`  
**Changes under analysis:**
- `README.md` — documentation update (minimum supported version 7.x+ → 8.0+)
- `.github/workflows-src/test/workflow.yml.tmpl` — CI matrix template (removed `7.17.13`)
- `.github/workflows/test.yml` — generated CI workflow (regenerated from template)

---

## 1. Unit Tests for Workflow Template Generation (`scripts/compile-workflow-sources`)

**Command run:** `go test ./scripts/compile-workflow-sources -cover -v`

| Result | Tests | Pass | Fail | Coverage |
|--------|-------|------|------|----------|
| **PASS** | 6 | 6 | 0 | **76.2%** |

### Tests executed
- `TestCompileWorkflowExpandsNestedIncludes`
- `TestCompileWorkflowCheckModeReportsDrift`
- `TestCompileWorkflowCheckModeVerboseReportsDiff`
- `TestRunCheckModeVerboseControlsDiffOutput` (with subtests)
- `TestCompileWorkflowDetectsIncludeCycles`
- `TestCompileWorkflowExpandsQuotedScriptInclude`

### Coverage detail (per function)

| Function | Coverage | Note |
|----------|----------|------|
| `CompileFromManifest` | 0.0% | Manifest-based batch compilation untested |
| `CompileWorkflow` | 81.0% | Write path (non-check mode) partially untested |
| `expandIncludes` | 96.7% | Well covered |
| `expandScriptIncludes` | 92.9% | Well covered |
| `cloneSeen` | 100.0% | Fully covered |
| `trimOptionalQuotes` | 100.0% | Fully covered |
| `indentLines` | 100.0% | Fully covered |
| `injectGeneratedHeader` | 100.0% | Fully covered |
| `createGeneratedHeader` | 100.0% | Fully covered |
| `resolvePath` | 100.0% | Fully covered |
| `normalizeRelativePath` | 75.0% | Error branch untested |
| `buildWorkflowDiff` | 100.0% | Fully covered |
| `main` | 0.0% | CLI entry point untested |
| `run` | 66.7% | Manifest path and flag-parsing branches partially untested |

**Relevance to Task 1:** The compiler code itself was **not modified** by Task 1. The change only edited a template file and regenerated its output. The existing tests sufficiently cover the compilation mechanisms used (`expandIncludes`, `expandScriptIncludes`, check-mode diffing), so no additional compiler tests are required for this change.

---

## 2. Tests for Workflow YAML Template (`.github/workflows-src/lib/*.test.mjs`)

**Command run:** `node --test .github/workflows-src/lib/*.test.mjs`

| Result | Tests | Pass | Fail | Duration |
|--------|-------|------|------|----------|
| **PASS** | 310 | 310 | 0 | ~262 ms |

### Scope of JS tests
The JS test suite covers the shared library modules used by various workflow inline scripts, such as:
- `change-factory-issue`, `code-factory-issue`
- `changelog-engine`, `changelog-renderer`, `changelog-pr-management`
- `pr-changelog-check` (gating, parser, validator, comment builders)
- `openspec-verify-label`
- `schema-coverage-rotation`
- `duplicate-code-detector`
- `select-change`
- `validate-test-result`
- `classify-changes`

### High-risk gap: no direct tests for the test workflow template
**There are no tests that directly validate `.github/workflows-src/test/workflow.yml.tmpl` or the generated `.github/workflows/test.yml`.** Specifically:
- **No matrix content validation tests.** The acceptance test matrix (the `version` list and `include` overrides) is not asserted anywhere. A malformed entry, typo in a version string, or incorrect `fleetImage` / `runner` pairing would only surface at CI runtime.
- **No tests for `7.17.13` removal.** The JS suite had zero references to `7.17.13`, fleet image mappings for the test workflow, or the matrix `include` block.
- **No end-to-end workflow compilation tests for the `test` workflow.** While other workflows (e.g., `change-factory-issue`, `verify-label`) have tests asserting their compiled outputs exist and match expected contracts, the `test` workflow does not.

---

## 3. Generated Workflow Synchronization Check

**Command run:** `make check-workflows`

**Result:** PASS (no output = no drift detected)

This confirms that `.github/workflows/test.yml` is correctly synchronized with `.github/workflows-src/test/workflow.yml.tmpl`.

---

## 4. High-Risk Untested Paths Related to These Changes

| Risk | Severity | Notes |
|------|----------|-------|
| **Matrix entry removal (`7.17.13`) was only verified manually** | Low | The change was a simple two-line deletion. `make check-workflows` and a manual grep confirmed the generated file has no `7.17` entries. However, there is no automated test that would fail if a 7.x entry were accidentally re-introduced. |
| **No structural validation of the acceptance test matrix** | Medium | The `version` list and `include` block are large, manually maintained YAML structures. There is no schema test or snapshot test for them. Errors (duplicate versions, invalid `fleetImage`/`runner` combinations, YAML syntax issues) are caught only at CI parse time. |
| **No tests for `classify_changes.inline.js` / `validate_test_result.inline.js` in the test workflow context** | Low | These inline scripts are tested indirectly via the shared `.github/workflows-src/lib/classify-changes.js` and `.github/workflows-src/lib/validate-test-result.js` library tests, but not wired into the actual test workflow template. |
| **Compiler batch path (`CompileFromManifest`) is untested** | Low | For Task 1 this is irrelevant because the compiler wasn't changed, but it's worth noting that the manifest-driven generation (used by `make workflow-generate`) has zero test coverage. |

---

## 5. Conclusion

### Tests run and their status
- **Go compiler tests:** 6/6 pass, 76.2% coverage.
- **JS library tests:** 310/310 pass, 0 failures.
- **Workflow sync check (`make check-workflows`):** PASS.

### Coverage metrics
- Compiler package: **76.2%** overall; core expansion logic is well covered, but CLI entry points and manifest batch compilation are untested.
- JS workflow libraries: **High coverage** for the modules that have tests; however, the test workflow template itself falls outside the tested surface area.

### High-risk untested paths
- The **acceptance test matrix YAML** (versions, `include` overrides, `fleetImage` / `runner` mappings) lacks any automated validation. This is the most significant gap relative to the Task 1 changes.
- There is **no regression test** that would automatically fail if a `7.x` matrix entry were re-added.

### Explicit statement
**Nothing further is needed for these doc/CI changes.**

- The `README.md` change is purely documentation; no tests apply.
- The template change is a trivial matrix entry removal. It has been validated by:
  1. Regenerating the workflow (`make workflow-generate`).
  2. Checking sync (`make check-workflows`).
  3. Manual verification that `7.17.13` is absent from both template and generated file.
- No compiler code was modified, so additional compiler coverage is not required for this task.

While adding a matrix snapshot test or schema validation for `.github/workflows-src/test/workflow.yml.tmpl` would reduce long-term CI risk, that is out of scope for Task 1 and would be a general infrastructure improvement rather than a requirement for the `remove-7x-support` changes.
