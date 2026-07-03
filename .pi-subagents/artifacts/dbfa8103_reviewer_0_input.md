# Task for reviewer

Critical code review of the new Go tool scaffold in `scripts/targeted-testacc/` (files: main.go, classifier.go, depgraph.go, entityname.go, testconsumers.go, acctestpackages.go, selector.go, gitdiff.go).
Focus on:
- Coding standards and idiomatic Go
- Likely logic bugs, especially in classifier, depgraph, entity regex, and shard selection
- Error handling gaps
- Risks for CI usage (stdout/stderr, exit codes, shelling out to git/go)
Read the context spec at `openspec/changes/selective-acceptance-tests/specs/selective-acceptance-tests/spec.md` first.
Return up to 10 prioritized findings with severity (high/medium/low), concise description, evidence (file:line if possible), and recommended fix. Do not edit files.

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