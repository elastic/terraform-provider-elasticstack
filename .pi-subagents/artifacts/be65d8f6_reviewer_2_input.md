# Task for reviewer

[Read from: /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/plan.md, /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/progress.md]

Run the schema-coverage skill/process for the elasticstack_kibana_dashboard resource, focused on the `description` attribute, in the repo at /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization.

Context: OpenSpec change `dashboard-description-null-normalization` adds intent-preserving null normalization for the dashboard root-level `description` attribute. New acceptance test TestAccResourceDashboardDescriptionNormalization in internal/kibana/dashboard/acc_test.go (testdata in testdata/TestAccResourceDashboardDescriptionNormalization/{omitted,empty}) covers omitted-description (null) and explicit description="" round-trips. The schema for `description` is in internal/kibana/dashboard/schema.go (Optional StringAttribute, not Computed).

Analyze the description attribute coverage: is the null/empty/non-empty matrix adequately covered? Are there set-only assertions, missing unset/empty cases, or missing update coverage for description? Identify high-risk untested behaviors. Return a prioritized report of missing or weak coverage with evidence (file:line) and recommended additions. Do not edit files.

---
**Output:**
Write your findings to exactly this path: /tmp/review-coverage.md
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