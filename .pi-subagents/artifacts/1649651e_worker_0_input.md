# Task for worker

Implement top-level task 2 of OpenSpec change `selective-acceptance-tests`: add unit tests under `scripts/targeted-testacc/`.

**Scope (do not modify the implementation under test unless tests expose real bugs):**
Complete subtasks 2.1 through 2.5:
- 2.1 `classifier_test.go` — test file-to-package mapping for `.go` files, `testdata/*.tf` files, and non-Go files; test force-all prefix detection for all five prefixes.
- 2.2 `entityname_test.go` — test regex extraction for all four call patterns (`NewResourceBase`, `NewElasticsearchResource`, `NewKibanaResource`, `NewKibanaDataSource`) from source snippets; test component string mapping to `elasticstack_<component>_<name>`.
- 2.3 `depgraph_test.go` — test reverse-dep walk on a synthetic graph (A imports B, B imports C → change C → reverse walk finds B and A).
- 2.4 `selector_test.go` — test force-all short-circuit; test run-all threshold (70%); test union/dedup of phase 1 + phase 2; test `ApplyShard` for all cases: `index >= total` → empty, `count < 30 && index > 0` → empty, `count < 30 && index == 0` → all, `count >= 30` → round-robin split.
- 2.5 `acctestpackages_test.go` — test enumeration using a minimal synthetic directory tree with `*_test.go` files.

**Testing constraints from the spec:**
- Tests must not require a live git repository or `go list` invocation.
- Use `t.TempDir()` for filesystem fixtures.
- A plain `grep` is acceptable for `isAccTestFile`/`FindAccTestPackages` tests.
- For `depgraph` tests, call the pure graph functions (`BuildReverseDepGraph`, `WalkReverseDeps`) with hand-built maps.

**Requirements:**
- Use table-driven tests where appropriate; keep them idiomatic.
- Do not write acceptance tests that require network or a real Elastic stack.
- Run `go test ./scripts/targeted-testacc/...` and ensure all tests pass.
- Run `go vet ./scripts/targeted-testacc/...`.
- Create focused git commits as you finish test files.
- Do NOT push.

**Context files to read first:**
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/specs/selective-acceptance-tests/spec.md`
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/tasks.md`
- All `scripts/targeted-testacc/*.go` files (they are the implementation under test).

Report back:
- test files created
- commits created
- `go test ./scripts/targeted-testacc/...` results
- blockers

## Acceptance Contract
Acceptance level: checked
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope

Required evidence: changed-files, tests-added, commands-run, residual-risks, no-staged-files

Finish with a fenced JSON block tagged `acceptance-report` in this shape:
Use empty arrays when no items apply; array fields contain strings unless object entries are shown.
```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "specific proof"
    }
  ],
  "changedFiles": [
    "src/file.ts"
  ],
  "testsAddedOrUpdated": [
    "test/file.test.ts"
  ],
  "commandsRun": [
    {
      "command": "command",
      "result": "passed",
      "summary": "short result"
    }
  ],
  "validationOutput": [
    "validation output or concise summary"
  ],
  "residualRisks": [
    "none"
  ],
  "noStagedFiles": true,
  "diffSummary": "short description of the diff",
  "reviewFindings": [
    "blocker: file.ts:12 - issue found, or no blockers"
  ],
  "manualNotes": "anything else the parent should know"
}
```