## 1. Update Schema Coverage Rotation source template

- [ ] 1.1 In `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` (line 43), change:
  ```
  labels: [testing, acceptance-tests, schema-coverage]
  ```
  to:
  ```
  labels: [testing, acceptance-tests, schema-coverage, triaged]
  ```

## 2. Update Duplicate Code Detector source template

- [ ] 2.1 In `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl` (line 31), change:
  ```
  labels: [duplicate-code, code-quality, automated-analysis]
  ```
  to:
  ```
  labels: [duplicate-code, code-quality, automated-analysis, triaged]
  ```

## 3. Update Semantic Function Refactor source template

- [ ] 3.1 In `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` (line 29), change:
  ```
  labels: [semantic-refactor, refactoring, code-quality, automated-analysis]
  ```
  to:
  ```
  labels: [semantic-refactor, refactoring, code-quality, automated-analysis, triaged]
  ```

## 4. Update Flaky Test Catcher source template

- [ ] 4.1 In `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl` (line 31), change:
  ```
  labels: [flaky-test]
  ```
  to:
  ```
  labels: [flaky-test, triaged]
  ```

## 5. Recompile lock files

- [ ] 5.1 Run `make workflow-generate` to regenerate all four `.lock.yml` files from the updated
  source templates:
  - `.github/workflows/schema-coverage-rotation.lock.yml`
  - `.github/workflows/duplicate-code-detector.lock.yml`
  - `.github/workflows/semantic-function-refactor.lock.yml`
  - `.github/workflows/flaky-test-catcher.lock.yml`
- [ ] 5.2 Verify that only the expected label additions appear in the compiled diff (the
  `GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG` env var in each lock file should reflect the updated label
  lists); confirm no other unintended template changes were included.

## 6. Validation

- [ ] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate issue-creation-triaged-label --type change` and resolve any reported issues.
- [ ] 6.2 Verify that the `triaged` GitHub label exists in the repository before the PR lands (the
  Issue Classifier already applies it, so it should exist).
- [ ] 6.3 Confirm `make build` still passes if any non-workflow source files were modified
  (unexpected for this change; skip if no Go files were touched).
