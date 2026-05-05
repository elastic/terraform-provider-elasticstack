## 1. Shared Component — Files and Structure

- [ ] 1.1 Create `.github/workflows-src/shared/` directory
- [ ] 1.2 Create `.github/workflows-src/shared/scripts/` directory
- [ ] 1.3 Write `.github/workflows-src/shared/scripts/compute_issue_slots.inline.js` with `//include: ../../lib/issue-slots.js`
- [ ] 1.4 Write `.github/workflows-src/shared/issue-slots.md.tmpl` containing:
  - `import-schema` with `label` (string, required) and `cap` (number, default: 3)
  - `steps` block with `compute_issue_slots` action using `x-script-include: scripts/compute_issue_slots.inline.js`
  - `jobs.pre-activation` with outputs `open_issues`, `issue_slots_available`, `gate_reason`
  - Body section `## Pre-activation context` parameterized with label and cap via `${{ github.aw.import-inputs.* }}`
- [ ] 1.5 Update `scripts/compile-workflow-sources/compiler.go` to create parent output directories via `os.MkdirAll` before `os.WriteFile`, so shared workflows can be generated under new paths like `.github/workflows/shared/`
- [ ] 1.6 Update `.github/workflows-src/manifest.json` to include the new shared component entry with output path `.github/workflows/shared/issue-slots.md`
- [ ] 1.7 Verify `make workflow-generate` produces `.github/workflows/shared/issue-slots.md` correctly (script is inlined, no `x-script-include` remains in output)

## 2. Consumer Workflow — duplicate-code-detector

- [ ] 2.1 Remove the `steps` block containing `compute_issue_slots` from `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl`
- [ ] 2.2 Remove `jobs.pre-activation` from `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl`
- [ ] 2.3 Remove the `## Pre-activation context` body section from `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl`
- [ ] 2.4 Add `imports` to frontmatter with `uses: shared/issue-slots.md` and `with: label: duplicate-code, cap: 3`
- [ ] 2.5 Delete `.github/workflows-src/duplicate-code-detector/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [ ] 2.6 Verify generated `.github/workflows/duplicate-code-detector.md` compiles and contains the merged pre-activation job

## 3. Consumer Workflow — semantic-function-refactor

- [ ] 3.1 Remove the `steps` block containing `compute_issue_slots` from `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl`
- [ ] 3.2 Remove `jobs.pre-activation` from `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl`
- [ ] 3.3 Remove the `## Pre-activation context` body section from `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl`
- [ ] 3.4 Add `imports` to frontmatter with `uses: shared/issue-slots.md` and `with: label: semantic-refactor, cap: 3`
- [ ] 3.5 Delete `.github/workflows-src/semantic-function-refactor/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [ ] 3.6 Verify generated `.github/workflows/semantic-function-refactor.md` compiles and contains the merged pre-activation job

## 4. Consumer Workflow — schema-coverage-rotation

- [ ] 4.1 Remove the `steps` block containing `compute_issue_slots` from `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl`
- [ ] 4.2 Remove `jobs.pre-activation` from `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl`
- [ ] 4.3 Remove the `## Pre-activation context` body section from `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl`
- [ ] 4.4 Add `imports` to frontmatter with `uses: shared/issue-slots.md` and `with: label: schema-coverage, cap: 3`
- [ ] 4.5 Delete `.github/workflows-src/schema-coverage-rotation/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [ ] 4.6 Verify generated `.github/workflows/schema-coverage-rotation.md` compiles and contains the merged pre-activation job

## 5. Validation and Cleanup

- [ ] 5.1 Run `make check-workflows` and confirm all generated workflow sources are up to date
- [ ] 5.2 Run `make workflow-test` (unit tests for workflow source generation) and confirm no regressions
- [ ] 5.3 Confirm `.github/workflows/shared/issue-slots.md` exists and is correctly generated
- [ ] 5.4 Confirm no orphaned `compute_issue_slots.inline.js` files remain in consumer directories
- [ ] 5.5 Verify `kibana-spec-impact` workflow is untouched (not in scope)
- [ ] 5.6 Commit the change
