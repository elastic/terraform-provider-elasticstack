## 1. Create shared dispatch-code-factory fragment

- [ ] 1.1 Create `.github/workflows/shared/dispatch-code-factory.md` with the canonical `safe-outputs.jobs.dispatch-code-factory` block, replacing the hardcoded `SOURCE_WORKFLOW: <slug>` with a runtime bash derivation: `SOURCE_WORKFLOW=$(echo "${{ github.workflow }}" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')`.
- [ ] 1.2 Smoke-test compiler support: add `imports: [shared/dispatch-code-factory.md]` to `schema-coverage-rotation.md`, run `gh aw compile` (or the repository's equivalent compile command), and verify the lock output is functionally identical to the current compiled lock. If `safe-outputs.jobs` merging is not supported, stop and document the blocker — do not proceed to 1.3+.

## 2. Update existing consumers to use the shared fragment

- [ ] 2.1 In `flaky-test-catcher.md`: add `imports: [shared/dispatch-code-factory.md]` (or append to an existing `imports:` list if present) and remove the inline `safe-outputs.jobs.dispatch-code-factory` block.
- [ ] 2.2 In `semantic-function-refactor.md`: add `imports: [shared/dispatch-code-factory.md]` to the existing `imports:` list and remove the inline `safe-outputs.jobs.dispatch-code-factory` block.
- [ ] 2.3 Complete the `schema-coverage-rotation.md` update from step 1.2 (smoke-test step already updated this file; remove the inline block if it wasn't already removed during smoke-test).

## 3. Add dispatch support to duplicate-code-detector

- [ ] 3.1 In `duplicate-code-detector.md`: add `imports: [shared/dispatch-code-factory.md]` to the frontmatter.
- [ ] 3.2 In `duplicate-code-detector.md`: append a `## Dispatch` section to the agent prompt with the instruction: "After creating all issues for this run (or if no issues were created), call the `dispatch_code_factory` safe output tool once to dispatch the `code-factory` workflow for each created issue."

## 4. Recompile lock files

- [ ] 4.1 Recompile all four workflow lock files (`flaky-test-catcher.lock.yml`, `semantic-function-refactor.lock.yml`, `schema-coverage-rotation.lock.yml`, `duplicate-code-detector.lock.yml`) using the repository's compile command.

## 5. Update tests

- [ ] 5.1 In `duplicate-code-detector.test.mjs`: add a new test case that asserts `dispatch_code_factory` appears in the workflow source and compiled lock, and that the lock contains the `dispatch-code-factory` job descriptor — matching the pattern used in `flaky-test-catcher.test.mjs` (line 209) and `schema-coverage-rotation-bootstrap.test.mjs` (line 92).
- [ ] 5.2 Verify that existing test assertions in `flaky-test-catcher.test.mjs`, `semantic-function-refactor.test.mjs` (or equivalent), and `schema-coverage-rotation-bootstrap.test.mjs` still pass against the updated source and recompiled locks. Update any assertions that checked for the inline block structure if they no longer match the import-based source.

## 6. Validate

- [ ] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate dispatch-code-factory-shared-fragment --type change` and fix any reported problems.
- [ ] 6.2 Run the workflow test suite (`node --test .github/scripts/workflows/lib/duplicate-code-detector.test.mjs` and the other three affected test files) to confirm all assertions pass.
