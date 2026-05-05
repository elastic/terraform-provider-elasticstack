## 1. Extend `code-factory` intake for dispatch mode

- [ ] 1.1 Update `.github/workflows-src/code-factory-issue/workflow.md.tmpl` to add a `workflow_dispatch` entrypoint with typed single-issue inputs and to drive the prompt from normalized intake outputs instead of direct `github.event.issue.*` references
- [ ] 1.2 Add deterministic pre-activation logic and any shared helper code needed to resolve normalized intake context for both issue-event and dispatch-triggered runs, including live issue fetch for dispatch mode and current-repository validation
- [ ] 1.3 Preserve and adapt duplicate linked-PR suppression so it applies identically to both issue-event and dispatch-triggered `code-factory` runs
- [ ] 1.4 Update `code-factory` workflow tests under `.github/workflows-src/lib/` to cover dispatch intake, normalized context resolution, and continued manual issue-event behavior

## 2. Add deterministic producer-side dispatch fan-out

- [ ] 2.1 Add repository-authored post-safe_outputs dispatch logic for `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` that parses `temporary-id-map.json` and dispatches one `code-factory` run per created issue
- [ ] 2.2 Add repository-authored post-safe_outputs dispatch logic for `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` that parses `temporary-id-map.json` and dispatches one `code-factory` run per created issue
- [ ] 2.3 Add repository-authored post-safe_outputs dispatch logic for `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl` that parses `temporary-id-map.json` and dispatches one `code-factory` run per created issue
- [ ] 2.4 Extract and test any shared helper logic needed to parse the temporary issue ID map and construct dispatch payloads deterministically

## 3. Remove producer-side `code-factory` label handoff

- [ ] 3.1 Remove `code-factory` from semantic refactor created-issue labels in `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` and update related tests/assertions
- [ ] 3.2 Remove `code-factory` from schema coverage created-issue labels in `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` and update related tests/assertions
- [ ] 3.3 Remove `code-factory` from flaky test catcher created-issue labels in `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl` and update related tests/assertions

## 4. Regenerate and verify workflow artifacts

- [ ] 4.1 Regenerate workflow markdown and compiled lock artifacts for `code-factory-issue`, `semantic-function-refactor`, `schema-coverage-rotation`, and `flaky-test-catcher`
- [ ] 4.2 Run the relevant workflow-source/unit test suites and any OpenSpec validation needed to verify the new dispatch-handoff contracts
