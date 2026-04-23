## 1. Author the `code-factory` issue-intake workflow contract

- [x] 1.1 Add the authored workflow source under `.github/workflows-src/code-factory-issue/` and register its generated output in `.github/workflows-src/manifest.json`.
- [x] 1.2 Define deterministic pre-activation handling for `issues.opened` and `issues.labeled`, including `code-factory` label qualification and trusted-actor evaluation for `github-actions[bot]` versus repository collaborators.
- [x] 1.3 Write the workflow prompt contract so the agent treats the triggering issue as the source of truth and creates exactly one linked `code-factory` PR on branch `code-factory/issue-<issue-number>`.

## 2. Implement deterministic helper logic and duplicate suppression

- [x] 2.1 Add extracted helper logic and inline scripts under `.github/workflows-src/lib/` and `.github/workflows-src/code-factory-issue/scripts/` for event qualification, actor trust checks, and duplicate linked-PR detection.
- [x] 2.2 Implement the duplicate linked-PR check so open PRs are matched using the `code-factory` label, deterministic head branch, and explicit issue reference metadata.
- [x] 2.3 Generate and commit the resulting workflow artifacts under `.github/workflows/`, including the compiled `.lock.yml` and any related lock metadata updates.

## 3. Validate workflow behavior and OpenSpec artifacts

- [x] 3.1 Add or update focused workflow-source tests covering issue-opened-with-label detection, trusted-actor gating, duplicate-PR suppression, and generated workflow expectations.
- [x] 3.2 Run the relevant workflow generation and workflow-source tests for the new `code-factory` issue workflow.
- [x] 3.3 Validate the OpenSpec change with `./node_modules/.bin/openspec validate code-factory-issue-workflow --type change` or an equivalent repository OpenSpec check.
