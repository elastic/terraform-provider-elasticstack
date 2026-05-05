## Why

There is no automated mechanism to detect broken or flaky acceptance tests accumulating on `main`. Failures surface only when developers notice CI noise or a test blocks a PR, leading to slow detection and manual triage.

## What Changes

- New GitHub Actions Agentic Workflow (`flaky-test-catcher`) that runs daily, inspects CI results for `main` over the last 3 days, classifies test failures as **broken** (100% fail rate) or **flaky** (≥ 20% fail rate), and opens one GitHub issue per affected resource.
- New workflow source template at `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl`, compiled to `.github/workflows/flaky-test-catcher.md`.
- New pre-activation JS script at `.github/workflows-src/flaky-test-catcher/scripts/check_ci_failures.inline.js`.
- New agent skill document at `.agents/skills/flaky-test-catcher/SKILL.md`.
- Updated `.github/workflows-src/manifest.json` to register the new workflow.
- Issues are labelled `flaky-test` + `code-factory`, which triggers the existing code-factory workflow to attempt automated fixes.

## Capabilities

### New Capabilities

- `flaky-test-catcher`: Daily workflow that identifies broken and flaky acceptance tests on `main`, performs commit-based fix detection, and opens structured GitHub issues per affected resource.

### Modified Capabilities

## Impact

- New files in `.github/workflows-src/flaky-test-catcher/` and `.github/workflows/`.
- `.github/workflows-src/manifest.json` gains one entry.
- No changes to provider Go code, existing resources, or acceptance tests.
- Downstream: `code-factory` workflow is triggered by the `code-factory` label on created issues.
