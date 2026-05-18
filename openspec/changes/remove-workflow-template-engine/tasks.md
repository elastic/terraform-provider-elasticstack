## 1. Bootstrap `.github/scripts/workflows/` module tree

- [x] 1.1 Create directory structure: `.github/scripts/workflows/lib/`, `.github/scripts/workflows/lib/intake/`, `.github/scripts/workflows/provider/`, `.github/scripts/workflows/workflows/`, `.github/scripts/workflows/issue-classifier/`, `.github/scripts/workflows/pr-changelog-check/`, `.github/scripts/workflows/flaky-test-catcher/`, `.github/scripts/workflows/change-factory/`, `.github/scripts/workflows/code-factory/`, `.github/scripts/workflows/research-factory/`, `.github/scripts/workflows/reproducer-factory/`, `.github/scripts/workflows/changelog/`, `.github/scripts/workflows/openspec-verify/`
- [x] 1.2 Move `workflows-src/lib/*.js` to `.github/scripts/workflows/lib/` (exclude `factory-issue-module.gh.js`, `compute_issue_slots.inline.js`, `set-phase-label.inline.js`, `set-phase-label.js`)
- [x] 1.3 Rename `workflows-src/lib/compute_issue_slots.inline.js` → `.github/scripts/workflows/lib/compute-issue-slots.js` (remove `//include: issue-slots.js` and require it)
- [x] 1.4 Rename `workflows-src/lib/set-phase-label.inline.js` → `.github/scripts/workflows/lib/set-phase-label.js` (remove `//include: set-phase-label.js` and require it)
- [x] 1.5 Rename `workflows-src/lib/set-phase-label.js` → `.github/scripts/workflows/lib/phase-label.js` (pure logic, update internal references)
- [x] 1.6 Move `workflows-src/change-factory-issue/intake-constants.js` → `.github/scripts/workflows/lib/intake/change-factory-constants.js`
- [x] 1.7 Move `workflows-src/code-factory-issue/intake-constants.js` → `.github/scripts/workflows/lib/intake/code-factory-constants.js`
- [x] 1.8 Move `workflows-src/research-factory-issue/intake-constants.js` → `.github/scripts/workflows/lib/intake/research-factory-constants.js`
- [x] 1.9 Move `workflows-src/reproducer-factory-issue/intake-constants.js` → `.github/scripts/workflows/lib/intake/reproducer-factory-constants.js`
- [x] 1.10 Create `.github/scripts/workflows/provider/classify-changes.js` containing the orchestration logic from `workflows-src/provider/scripts/classify_changes.inline.js` (wrapping `lib/classify-changes.js`)
- [x] 1.11 Create `.github/scripts/workflows/provider/gate.js` containing the orchestration logic from `workflows-src/provider/scripts/gate.inline.js` (wrapping `lib/gate-provider.js`)
- [x] 1.12 Create `.github/scripts/workflows/workflows/classify-changes.js` containing the orchestration logic from `workflows-src/workflows/scripts/classify_changes.inline.js`
- [x] 1.13 Create `.github/scripts/workflows/workflows/gate.js` containing the orchestration logic from `workflows-src/workflows/scripts/gate.inline.js` (wrapping `lib/gate-workflows.js`)
- [x] 1.14 Create `.github/scripts/workflows/issue-classifier/classify-issues.js` from `workflows-src/issue-classifier/scripts/compute_issues.inline.js`
- [x] 1.15 Create `.github/scripts/workflows/pr-changelog-check/check.js` from `workflows-src/pr-changelog-check/scripts/pr-changelog-check.inline.js`
- [x] 1.16 Create `.github/scripts/workflows/flaky-test-catcher/catch.js` from `workflows-src/flaky-test-catcher/scripts/check_ci_failures.inline.js` (wrapping `lib/flaky-test-catcher.js` and `lib/issue-slots.js`)
- [x] 1.17 Create `.github/scripts/workflows/changelog/` modules from `workflows-src/changelog-generation/scripts/*.inline.js`
- [x] 1.18 Create `.github/scripts/workflows/openspec-verify/` modules from `workflows-src/openspec-verify-label/scripts/*.inline.js`
- [x] 1.19 Create `.github/scripts/workflows/change-factory/` modules from `workflows-src/change-factory-issue/scripts/*.inline.js`
- [x] 1.20 Create `.github/scripts/workflows/code-factory/` modules from `workflows-src/code-factory-issue/scripts/*.inline.js`
- [x] 1.21 Create `.github/scripts/workflows/research-factory/` modules from `workflows-src/research-factory-issue/scripts/*.inline.js`
- [x] 1.22 Create `.github/scripts/workflows/reproducer-factory/` modules from `workflows-src/reproducer-factory-issue/scripts/*.inline.js`
- [x] 1.23 Verify all new modules use `require()` for their lib dependencies (no remaining `//include:`)

## 2. Delete obsolete files and directories

- [x] 2.1 Delete `scripts/compile-workflow-sources/` (compiler source code and tests)
- [x] 2.2 Delete `.github/workflows-src/` entirely
- [x] 2.3 Delete `lib/factory-issue-module.gh.js` if it was copied (it is superseded by `factory-issue-shared.js`)

## 3. Update Makefile targets

- [x] 3.1 Remove `go run ./scripts/compile-workflow-sources` from `workflow-generate`; keep `gh aw compile`
- [x] 3.2 Remove `go test ./scripts/compile-workflow-sources` from `workflow-test`; update `node --test` path to `.github/scripts/workflows/lib/*.test.mjs`
- [x] 3.3 Remove `check-workflows` target entirely (or convert to a no-op / reminder)
- [x] 3.4 Verify `make workflow-test` passes after path updates

## 4. Update `.github/workflows/*.yml` files

### 4.1 `provider.yml`
- [x] 4.1.1 Remove generated header comment
- [x] 4.1.2 Add `actions/checkout` step to the `classify` job before `actions/github-script`
- [x] 4.1.3 Replace `classify` inline `script:` with `require('${{ github.workspace }}/.github/scripts/workflows/provider/classify-changes.js')` call
- [x] 4.1.4 Add `actions/checkout` step to the `gate` job before `actions/github-script`
- [x] 4.1.5 Replace `gate` inline `script:` with `require('${{ github.workspace }}/.github/scripts/workflows/provider/gate.js')` call

### 4.2 `workflows.yml`
- [x] 4.2.1 Remove generated header comment
- [x] 4.2.2 Add `actions/checkout` step to the `classify` job
- [x] 4.2.3 Replace `classify` inline `script:` with `require()` call to `.github/scripts/workflows/workflows/classify-changes.js`
- [x] 4.2.4 Add `actions/checkout` step to the `gate` job
- [x] 4.2.5 Replace `gate` inline `script:` with `require()` call to `.github/scripts/workflows/workflows/gate.js`

### 4.3 `pr-changelog-check.yml`
- [x] 4.3.1 Remove generated header comment
- [x] 4.3.2 Add `actions/checkout` step to the job before `actions/github-script`
- [x] 4.3.3 Replace massive inline `script:` with `require()` call to `.github/scripts/workflows/pr-changelog-check/check.js`

### 4.4 `changelog-generation.yml`
- [x] 4.4.1 Remove generated header comment
- [x] 4.4.2 Replace inline scripts with `require()` calls to `.github/scripts/workflows/changelog/` modules

### 4.5 `prep-release.yml`
- [x] 4.5.1 Remove generated header comment
- [x] 4.5.2 Replace inline script with `require()` call to `.github/scripts/workflows/changelog/` module

### 4.6 `openspec.yml`
- [x] 4.6.1 Remove generated header comment

## 5. Update `.github/workflows/*.md` files

- [x] 5.1 Remove generated header comments from all `.md` workflow files
- [x] 5.2 Replace all `x-script-include:` directives with `script:` blocks using `require('${{ github.workspace }}/.github/scripts/workflows/...')`
- [x] 5.3 Ensure `actions/checkout` step is present before any `actions/github-script` step that uses `require()` (already present for most `.md` workflows via `gh aw compile`, verify each)
- [x] 5.4 Update `openspec-verify-label.md`
- [x] 5.5 Update `change-factory-issue.md`
- [x] 5.6 Update `code-factory-issue.md`
- [x] 5.7 Update `research-factory-issue.md`
- [x] 5.8 Update `reproducer-factory-issue.md`
- [x] 5.9 Update `flaky-test-catcher.md`
- [x] 5.10 Update `issue-classifier.md`
- [x] 5.11 Update `schema-coverage-rotation.md`
- [x] 5.12 Update `duplicate-code-detector.md`
- [x] 5.13 Update `ci-deadcode-removal-rotation.md`
- [x] 5.14 Update `semantic-function-refactor.md`
- [x] 5.15 Update `kibana-spec-impact.md`

## 6. Regenerate `.lock.yml` files

- [x] 6.1 Run `gh aw compile` for all `.md` workflows to regenerate `.lock.yml` files
- [x] 6.2 Verify no manual edits were made to `.lock.yml` files

## 7. Update and prune tests

- [ ] 7.1 Move `.github/workflows-src/lib/*.test.mjs` to `.github/scripts/workflows/lib/*.test.mjs`
- [ ] 7.2 Update `require()` paths inside all moved `.test.mjs` files
- [ ] 7.3 Delete `code-factory-inline-scripts.test.mjs` (tests `//include:` expansion, no longer applicable)
- [ ] 7.4 Delete `research-factory-inline-scripts.test.mjs` (tests `//include:` expansion, no longer applicable)
- [ ] 7.5 Delete compiler Go tests (deleted with compiler)
- [ ] 7.6 Delete any test that asserts on `x-script-include` presence in templates
- [ ] 7.7 Run `make workflow-test` and fix any failing tests

## 8. Final validation

- [ ] 8.1 Run `make build` and ensure no Go compilation errors
- [ ] 8.2 Run `grep -r "x-script-include" .github/workflows/` and confirm zero matches
- [ ] 8.3 Run `grep -r "//include:" .github/scripts/workflows/` and confirm zero matches
- [ ] 8.4 Run `grep -r "compile-workflow-sources" Makefile` and confirm zero matches (except possibly in comments)
- [ ] 8.5 Verify `scripts/compile-workflow-sources/` does not exist
- [ ] 8.6 Verify `.github/workflows-src/` does not exist
- [ ] 8.7 Verify all `.github/workflows/*.yml` files lack generated-by headers referencing compile-workflow-sources
