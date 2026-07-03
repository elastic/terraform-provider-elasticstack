# Task for delegate

Monitor GitHub PR 4063 for the terraform-provider-elasticstack repo using this script from the repo root:

  .agents/skills/pr-monitoring-loop/scripts/check-pr-state.py 4063 --watch --interval 120 --max-duration 3600

The script auto-creates/reuses a state file at .agents/skills/pr-monitoring-loop/scripts/state/.pr-monitor-4063.json for new-vs-old detection. Stream the NDJSON ticks.

Success criteria for this run: all CI checks pass (checks.failed == 0, no pending) with no actionable items. verify-openspec is NOT required for this run — do not touch that label.

Poll until one of these happens:
- PR CI is fully green with no actionable items (status: ready)
- a CI check fails (status: delegate — report the failing check name, URL, and any log excerpt)
- a new actionable PR comment, review comment, unresolved review thread, or CHANGES_REQUESTED review appears (status: delegate)
- merge conflict / stale branch (status: delegate)
- timeout (status: timeout)

Print the full script JSON output on the actionable/final tick before deciding. In your final result include:
- status: ready | fixed-and-continued | delegate | blocked | timeout
- PR URL and head SHA
- actionable item summary
- evidence excerpt from check-pr-state.py (actionable, checks.failedChecks, threadDetails as relevant)
- counts of seen vs new IDs
- recommended delegate scope if status is delegate

Do NOT fix anything yourself for this run — just report back. The recent commits fixed build errors and aiops panel type-change-recovery unit tests, so focus especially on whether those CI jobs now pass.

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