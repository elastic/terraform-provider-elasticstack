# Task for worker

Validation runner for task 1 of `selective-acceptance-tests`.
Run in `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests`:
1. `go build ./scripts/targeted-testacc/...`
2. `go vet ./scripts/targeted-testacc/...`
3. `go test ./scripts/targeted-testacc/...` (it may have no tests yet; just report result)
4. `go run ./scripts/targeted-testacc/... --base=HEAD --dry-run` (empty diff should exit 0)
5. `go run ./scripts/targeted-testacc/... --base=HEAD~1 --dry-run` (report selected package count)
Return a concise validation summary with pass/fail for each command and any stderr output.

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