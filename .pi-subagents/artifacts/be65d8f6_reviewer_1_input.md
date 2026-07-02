# Task for reviewer

[Read from: /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/plan.md, /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/progress.md]

Run the openspec-verify-change skill/process for the OpenSpec change `dashboard-description-null-normalization` in the repo at /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization.

Context: This change fixes Kibana 9.5 returning description: "" for omitted dashboard descriptions, causing Terraform inconsistent-result-after-apply. The implementation: in internal/kibana/dashboard/models.go dashboardPopulateFromAPI, added intent-preserving null normalization (API returns "" and prior state null -> store null; explicit "" preserved). Added unit tests and an acceptance test TestAccResourceDashboardDescriptionNormalization.

Read the change artifacts: openspec/changes/dashboard-description-null-normalization/{proposal.md,design.md,tasks.md,specs/kibana-dashboard/spec.md}. Verify implementation matches the proposal/design/specs. Run `openspec status --change dashboard-description-null-normalization --json` and `openspec validate` if available. Return ONLY actionable mismatches, missing work, or notable warnings with severity, finding, evidence, and recommended fix. Do not edit files.

---
**Output:**
Write your findings to exactly this path: /tmp/review-spec.md
This path is authoritative for this run.
Ignore any other output filename or output path mentioned elsewhere, including output destinations in the base agent prompt, system prompt, or task instructions.

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