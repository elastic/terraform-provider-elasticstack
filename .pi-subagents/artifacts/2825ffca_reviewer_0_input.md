# Task for reviewer

Review the updated `.github/workflows/provider.yml` for the `selective-acceptance-tests` change. Read the current committed file via `git show HEAD:.github/workflows/provider.yml` or the working tree file. Compare against the spec `openspec/changes/selective-acceptance-tests/specs/ci-provider-acceptance-tests/spec.md`. Report any mismatches or issues, especially:
- whether the `compute-packages` step correctly routes PR vs non-PR events
- whether the tool invocation includes `--total-shards=2 --shard-index=${{ matrix.shard }}`
- whether all expensive steps are gated on `steps.targeted.outputs.has_packages == 'true'`
- whether test step routing uses `targeted_pkgs` emptiness
- whether `merge_group` trigger is added
- any YAML/syntax/quoting concerns.
Return findings only; do not edit.

## Acceptance Contract
Acceptance level: attested
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Return concrete findings with file paths and severity when applicable

Required evidence: review-findings, residual-risks

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