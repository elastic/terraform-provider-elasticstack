## Why

The repository now has a `duplicate-code-detector` GitHub Agentic Workflow that scans recent code changes for meaningful duplication and opens focused refactoring issues. This behavior should be captured as an OpenSpec change so the workflow contract, generated artifacts, issue-slot gating, and reporting expectations are reviewable and maintainable as requirements rather than implicit implementation details.

## What Changes

- Add a new OpenSpec capability for the `duplicate-code-detector` workflow.
- Define the workflow as authored source under `.github/workflows-src/` that generates checked-in workflow artifacts under `.github/workflows/`, including the compiled `.lock.yml`.
- Define deterministic pre-activation issue-slot gating based on open issues carrying the `duplicate-code` label and a workflow-configured issue cap.
- Define the analysis scope, significance threshold, and issue-creation behavior for duplicate-code findings.
- Define the reporting contract so each issue covers exactly one actionable duplication pattern with concrete evidence and refactoring guidance.

## Capabilities

### New Capabilities
- `ci-duplicate-code-detector`: scheduled and manually triggered duplicate-code analysis that opens bounded, actionable GitHub issues from a generated GH AW workflow

### Modified Capabilities
<!-- None. -->

## Impact

- New authored workflow source under `.github/workflows-src/duplicate-code-detector/` and generated workflow artifacts under `.github/workflows/`
- Shared workflow helper logic for deterministic issue-slot computation under `.github/workflows-src/lib/`
- GitHub Actions permissions, safe outputs, and compiled lock metadata for the new GH AW workflow
- Maintainer expectations for how duplicate-code issues are capped, labeled, and structured for follow-up remediation
