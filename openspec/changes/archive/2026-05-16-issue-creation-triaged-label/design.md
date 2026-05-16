## Context

The **Issue Classifier** workflow gates its classification run with a single pre-flight check:
`labels.includes('triaged')`. If the label is already present, the classifier skips the issue
entirely. The classifier applies `triaged` as its final output label when it finishes routing an
issue. Producer workflows that create fully defined automated issues should set `triaged` at
creation time so the classifier's gate sees it and skips the issue without wasting a run.

The `safe-outputs.create-issue.labels` field in each workflow source template (`.md.tmpl`) is the
authoritative declaration of the complete label set applied when the workflow creates an issue.
All four affected workflows already use workflow-specific labels (e.g. `schema-coverage`,
`flaky-test`) through this mechanism; adding `triaged` to that list follows the same pattern.

Source templates are compiled into paired `.lock.yml` files via `make workflow-generate`. Both the
template change and the regenerated lock file must be committed together.

## Goals / Non-Goals

**Goals:**

- Eliminate unnecessary Issue Classifier runs caused by automated producer workflows.
- Apply `triaged` atomically at issue creation time, with no race condition.
- Keep the complete label set for each workflow visible in its source template.

**Non-Goals:**

- Changes to the Issue Classifier itself.
- Retroactively labelling existing open issues.
- Adding `triaged` to the `kibana-spec-impact` workflow (different triage lifecycle; explicitly
  out of scope per the research baseline).
- Adding `triaged` to any other producer workflow not named in this change.
- Label permissions or role-based label management.

## Decisions

### 1. Label at creation time (Approach A)

`triaged` is appended to the `labels:` list in the `safe-outputs.create-issue` block of each
affected source template. This is idiomatic for this repo's tooling: the field already controls
the full label set for created issues. Approach B (a reactive `issues: labeled` workflow that adds
`triaged` after issue creation) was considered and rejected because it introduces a race condition
with the classifier and a false-positive risk if contributors manually apply automation-specific
labels to human-filed issues.

### 2. Four workflows in scope

The issue body names three workflows (Schema Coverage Rotation, Duplicate Code Detector, Semantic
Function Refactor). `@tobio` confirmed in issue comments that **Flaky Test Catcher** should also
receive `triaged` — the same rationale applies because that workflow creates equally well-defined
issues and dispatches `code-factory` immediately.

### 3. No shared mechanism

A global default or centralized label injection was considered. It was rejected because:
- The existing `safe-outputs.create-issue.labels` mechanism is already per-workflow and explicit.
- A centralized mechanism would require changes to the workflow compilation framework.
- Per-workflow explicitness is more auditable and easier to reason about.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Future automated workflows forget to add `triaged` | Document the convention in the relevant contributor guide or workflow source comment. Not blocking this change. |
| `triaged` label does not exist in the repository | The Issue Classifier already applies it (confirmed in `issue-classifier/workflow.md.tmpl`), so it exists. Verify before the PR lands. |
| Lock file drift | Both template and lock file are committed in the same PR; CI validates they match. |

## Migration Plan

1. For each of the four source templates, append `triaged` to the `labels:` list in the
   `safe-outputs.create-issue` block.
2. Run `make workflow-generate` to regenerate all four `.lock.yml` files.
3. Commit both the template changes and the regenerated lock files.

**Rollback**: Remove `triaged` from each template's `labels:` list and rerun `make workflow-generate`.
Existing issues that already carry the label are unaffected.

## Open Questions

The following questions were raised during research. Where a resolution is known, it is documented
here for implementer reference.

- **Should the `triaged` label be documented as an automation-applied label?**
  Currently `triaged` is described only in the context of the classifier. As more producer workflows
  apply it at creation time, it may be worth noting in contributor documentation that `triaged` can
  also be applied by producers. This is non-blocking; no documentation change is required for this
  change to land.

- **Does the `triaged` GitHub label exist in the repository?**
  The Issue Classifier already applies it, so it should exist. Verify before the PR lands.

- **Should the `triaged` label be listed in a canonical "automation labels" reference?**
  Non-blocking. Consider updating contributor documentation in a follow-up if this becomes a source
  of confusion.
