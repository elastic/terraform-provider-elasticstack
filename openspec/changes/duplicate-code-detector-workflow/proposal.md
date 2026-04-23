## Why

The repository now has a `duplicate-code-detector` GitHub Agentic Workflow derived from the upstream `gh-aw` workflow source at `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md`. This behavior should be captured as an OpenSpec change so the upstream source reference, repository-authored adaptations, generated artifacts, issue-slot gating, and reporting expectations are reviewable and maintainable as requirements rather than implicit implementation details.

## What Changes

- Add a new OpenSpec capability for the `duplicate-code-detector` workflow.
- Define the workflow as repository-authored source under `.github/workflows-src/`, derived from the upstream `gh-aw` duplicate-code detector workflow, and generating checked-in workflow artifacts under `.github/workflows/`, including the compiled `.lock.yml`.
- Define deterministic pre-activation issue-slot gating based on open issues carrying the `duplicate-code` label and a workflow-configured issue cap.
- Record the related schema-coverage workflow contract update that renames the published open-issue-count output to `open_issues` so both workflows share the same issue-slot helper vocabulary.
- Define the analysis scope, significance threshold, and issue-creation behavior for duplicate-code findings.
- Define the reporting contract so each issue covers exactly one actionable duplication pattern with concrete evidence and refactoring guidance.

## Capabilities

### New Capabilities
- `ci-duplicate-code-detector`: scheduled and manually triggered duplicate-code analysis that opens bounded, actionable GitHub issues from a generated GH AW workflow

### Modified Capabilities
- `ci-schema-coverage-rotation-issue-slots`: align the published open-issue-count output with the shared issue-slot helper by renaming `open_schema_coverage_issues` to `open_issues`.

## Impact

- New authored workflow source under `.github/workflows-src/duplicate-code-detector/`, maintained against the upstream `gh-aw` source at `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md`, and generated workflow artifacts under `.github/workflows/`
- Shared workflow helper logic for deterministic issue-slot computation under `.github/workflows-src/lib/`
- The existing `ci-schema-coverage-rotation-issue-slots` OpenSpec capability, whose published open-issue-count output now aligns with the shared helper contract
- GitHub Actions permissions, safe outputs, and compiled lock metadata for the new GH AW workflow
- Maintainer expectations for how duplicate-code issues are capped, labeled, and structured for follow-up remediation
