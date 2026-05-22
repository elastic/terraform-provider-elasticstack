## Why

Three issue-producing agentic workflows (`flaky-test-catcher`, `semantic-function-refactor`, `schema-coverage-rotation`) each contain an identical `safe-outputs.jobs.dispatch-code-factory` block that fans out `code-factory` workflow runs for each created issue. A fourth workflow, `duplicate-code-detector`, creates issues in the same pattern but lacks this block entirely — meaning issues it opens never receive automated `code-factory` treatment.

The inline block is copy-pasted verbatim across the three existing consumers, with only a per-workflow `SOURCE_WORKFLOW: <slug>` literal differing. This duplication makes every future change to the dispatch logic a three-file edit.

## What Changes

- Extract the `dispatch-code-factory` safe-outputs job block into a new shared gh-aw fragment at `.github/workflows/shared/dispatch-code-factory.md`, replacing the hardcoded `SOURCE_WORKFLOW` literal with a runtime derivation from `${{ github.workflow }}`.
- Replace the inline block in `flaky-test-catcher.md`, `semantic-function-refactor.md`, and `schema-coverage-rotation.md` with `imports: [shared/dispatch-code-factory.md]`.
- Add `imports: [shared/dispatch-code-factory.md]` to `duplicate-code-detector.md`, making it the fourth consumer.
- Add the "Dispatch" instruction paragraph to `duplicate-code-detector.md`'s agent prompt so the agent knows to call `dispatch_code_factory` after creating issues.
- Recompile all four lock files.
- Update each workflow's test file to assert on the import declaration and the compiled lock's `dispatch_code_factory` presence; add corresponding assertions to `duplicate-code-detector.test.mjs`.

## Capabilities

### New Capabilities
- `gh-aw-shared-dispatch-code-factory`: shared gh-aw fragment that canonically defines the `dispatch-code-factory` safe-outputs job

### Modified Capabilities
- `ci-duplicate-code-detector`: gains `dispatch-code-factory` support via shared import and updated agent prompt

## Non-Goals

- Behavioral changes to `producer-dispatch.js` or `code-factory-issue.lock.yml`.
- Adding `dispatch-code-factory` to workflows beyond the four identified.
- Extracting other duplicated safe-output blocks (e.g., `noop` config, issue-slot patterns).

## Impact

- New shared workflow fragment: `.github/workflows/shared/dispatch-code-factory.md`
- Four workflow source files updated (three to replace inline blocks; one to add the import and prompt section)
- All four corresponding compiled lock files recompiled
- Four test files updated with dispatch assertions
