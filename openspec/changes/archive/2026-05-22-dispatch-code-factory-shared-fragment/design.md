## Context

The `dispatch-code-factory` safe-outputs job is a deterministic post-agent job that reads the safe-outputs temporary-ID artifact and dispatches `code-factory-issue.lock.yml` once per created issue. It is functionally identical across the three existing consumers (`flaky-test-catcher`, `semantic-function-refactor`, `schema-coverage-rotation`), differing only in the `SOURCE_WORKFLOW` environment variable used to identify the calling workflow.

The precedent for shared gh-aw fragments already exists in this repository: `shared/setup-dev.md` demonstrates that the gh-aw compiler merges `steps` and `network.allowed` entries from imported fragments. The `safe-outputs.jobs` key is the merge dimension needed here.

## Goals / Non-Goals

**Goals:**
- Eliminate the copy-pasted `dispatch-code-factory` block from three workflow source files.
- Give `duplicate-code-detector` the same code-factory dispatch behavior the other three workflows already have.
- Keep the shared fragment self-contained so adding a fifth consumer is a one-line `imports:` addition.

**Non-Goals:**
- Changes to `producer-dispatch.js`, `code-factory-issue.lock.yml`, or any other workflow not in the four identified.
- Extracting other duplicated safe-output configurations (issue-slot scripts, noop declarations, etc.).

## Decisions

### 1. Dynamic `SOURCE_WORKFLOW` derivation via `${{ github.workflow }}`

Replace the hardcoded `SOURCE_WORKFLOW: <slug>` literal in each consumer with a runtime slug derived from the calling workflow's display name. `gh aw compile` rejects `${{ }}` inside `run:` scripts, so the fragment uses env indirection:

```yaml
env:
  GITHUB_WORKFLOW_NAME: ${{ github.workflow }}
run: |
  SOURCE_WORKFLOW=$(echo "$GITHUB_WORKFLOW_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
```

The four workflow `name:` values produce the expected identifier strings:
- `"Flaky Test Catcher"` → `flaky-test-catcher` ✓
- `"Semantic Function Refactor"` → `semantic-function-refactor` ✓
- `"Schema Coverage Rotation"` → `schema-coverage-rotation` ✓
- `"Duplicate Code Detector"` → `duplicate-code-detector` (new, no prior hardcode to match)

Why: Avoids a new hardcoded string in the shared fragment while making the slug derivation self-documenting for future consumers. The derivation is verifiable ahead of time because all current workflow display names slug-map cleanly.

Alternative considered — Keep `SOURCE_WORKFLOW` as a parameter the importing workflow must supply: rejected because it adds configuration burden to every consumer and the derivation is already unambiguous.

### 2. Shared fragment file path

Place the shared fragment at `.github/workflows/shared/dispatch-code-factory.md`, following the established `shared/setup-dev.md` precedent. The `imports:` declaration in each consumer will be `imports: [shared/dispatch-code-factory.md]`.

### 3. Confirm gh-aw compiler supports `safe-outputs.jobs` merging before committing

The research comment notes this as a blocking open question. The implementation task list includes a compiler smoke-test step: compile one workflow (e.g., `schema-coverage-rotation`) with the shared import before doing the full four-workflow update. If the compiler rejects `safe-outputs.jobs` merging, pivot to Approach B (inline copy into DCD only) as a fallback.

### 4. `duplicate-code-detector.md` prompt update

Add a `## Dispatch` section to the agent prompt matching the wording already present in the other three workflows:

> After creating all issues for this run (or if no issues were created), call the `dispatch_code_factory` safe output tool once to dispatch the `code-factory` workflow for each created issue.

### 5. Test assertions for `dispatch-code-factory`

The test pattern from `flaky-test-catcher.test.mjs` (line 209) and `schema-coverage-rotation-bootstrap.test.mjs` (line 92) is:

```js
assert.match(source, /dispatch_code_factory/);
assert.match(source, /Dispatch/);
assert.match(lock, /dispatch_code_factory/);
assert.match(lock, /"dispatch-code-factory":\{"description":"Dispatch code-factory for each created issue"\}/);
assert.match(lock, /"dispatch_code_factory"/);
```

Add this pattern to `duplicate-code-detector.test.mjs` as a new test case. Verify the existing assertions in the other three test files still pass after the inline-to-import refactor (the lock file output should be identical; the source file will now contain `imports:` instead of the inline block).

## Risks / Trade-offs

- [gh-aw compiler may not merge `safe-outputs.jobs` from shared fragments] → Mitigation: confirm with a single-workflow smoke-test before doing the full change; fall back to inline DCD copy if not supported.
- [Workflow rename would silently change `SOURCE_WORKFLOW` slug] → Mitigation: document in the shared fragment that the slug is derived from the workflow `name:` and must match `code-factory-issue.lock.yml` routing expectations. Test assertions on lock content provide a regression signal.
- [Three existing consumers' lock files change, potentially causing compile-time regressions] → Mitigation: run all four workflow test suites after recompile to confirm lock content is functionally identical (only the inline block vs. import changes, not the compiled output).

## Open Questions

- **Blocking**: Does gh-aw's `imports:` mechanism support merging `safe-outputs.jobs` entries from a shared fragment? Confirm by running `gh aw compile` on one workflow with the shared import before the full change.
- **Non-blocking**: Does `code-factory-issue.lock.yml` route or filter based on the `source_workflow` literal value, or is it purely metadata/display? If it routes on specific strings, verify the slug derivation produces identical output for the three existing consumers.
- **Non-blocking**: Should `duplicate-code-detector.test.mjs` also assert that the dispatch instruction paragraph appears in the agent prompt (matching the pattern in `schema-coverage-rotation-bootstrap.test.mjs`)?

## Migration Plan

1. Create `.github/workflows/shared/dispatch-code-factory.md` with the shared job block and dynamic slug derivation.
2. Smoke-test: add `imports: [shared/dispatch-code-factory.md]` to one existing consumer (e.g., `schema-coverage-rotation.md`), run `gh aw compile`, and verify the lock output is functionally identical.
3. Update the remaining two existing consumers (`flaky-test-catcher.md`, `semantic-function-refactor.md`) to use the import; remove their inline `safe-outputs.jobs.dispatch-code-factory` blocks.
4. Add `imports: [shared/dispatch-code-factory.md]` and the `## Dispatch` prompt section to `duplicate-code-detector.md`.
5. Recompile all four lock files.
6. Update all four test files with dispatch assertions; run tests to confirm they pass.
7. Verify OpenSpec validation passes.
