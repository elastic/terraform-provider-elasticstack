# Task for worker

Run the `openspec-verify-change` skill for the `selective-acceptance-tests` change in `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests`.

Use the process from `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/.agents/skills/openspec-verify-change/SKILL.md`:
1. Run `openspec status --change "selective-acceptance-tests" --json` and `openspec instructions apply --change "selective-acceptance-tests" --json`.
2. Read all context files from the instruction output.
3. Verify completeness against `tasks.md` and `specs/**/*.md` from the change context.
4. Report whether implementation matches the change artifacts. Focus on critical/warning issues only; do not edit files.

Report back:
- a concise verification summary
- any CRITICAL or WARNING findings with file references
- whether the implementation appears ready to archive

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