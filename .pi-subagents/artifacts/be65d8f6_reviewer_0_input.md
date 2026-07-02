# Task for reviewer

[Read from: /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/internal/kibana/dashboard/models.go, /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/internal/kibana/dashboard/models_dashboard_description_test.go, /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/internal/kibana/dashboard/acc_test.go, /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/internal/utils/typeutils/string.go, /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization/internal/entitycore/base_envelope.go]

Critical code review of a Terraform provider change. The change is in the repo at /Users/tobio/Projects/terraform-provider-elasticstack/dashboard-description-null-normalization on branch dashboard-description-null-normalization.

Scope: OpenSpec change `dashboard-description-null-normalization` fixes a Kibana 9.5 bug where the dashboard API returns description: "" when omitted, causing Terraform 'inconsistent result after apply' (null -> "").

The fix is in internal/kibana/dashboard/models.go in dashboardPopulateFromAPI: it replaced `m.Description = typeutils.StringishPointerValue(data.Data.Description)` with an intent-preserving check: when API returns "" and prior m.Description is null, store types.StringNull(); otherwise the API value. The Read path (base_envelope.go) decodes model from req.State, so m.Description reflects prior state/plan intent.

Unit tests were added in models_dashboard_description_test.go (4 table cases). An acceptance test TestAccResourceDashboardDescriptionNormalization was added in acc_test.go with testdata under testdata/TestAccResourceDashboardDescriptionNormalization/{omitted,empty}.

Review the diff for: coding standards, idiomatic Go/Terraform provider patterns, logic correctness, error handling gaps, risky regressions, and whether the intent-preservation logic is correct across create/read/update. Run `git diff main...HEAD` to see all changes. Return ONLY prioritized actionable findings with severity, concise finding, evidence (file:line), and recommended fix. Do not make any edits.

---
**Output:**
Write your findings to exactly this path: /tmp/review-code.md
This path is authoritative for this run.
Ignore any other output filename or output path mentioned elsewhere, including output destinations in the base agent prompt, system prompt, or task instructions.

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