# Task for worker

Run final local validation for the `selective-acceptance-tests` implementation in `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests`.
Commands:
1. `go vet ./scripts/targeted-testacc/...`
2. `go test ./scripts/targeted-testacc/...`
3. `go build ./scripts/targeted-testacc/...`
4. `go build ./...`
5. `make targeted-testacc-dry-run ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=0 TARGETED_TESTACC_BASE=HEAD~3` (just print the dry-run summary tail)
6. `make targeted-testacc-dry-run ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=1 TARGETED_TESTACC_BASE=HEAD` (should report all packages split or empty for shard 1 depending on threshold; just report result)
7. `npx openspec validate --specs`
Return a concise summary of pass/fail for each and the final line counts / selected package counts.

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