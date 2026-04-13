## Why

Commenting `@copilot` from GitHub Actions is not a reliable contract for failed-CI remediation, and author-based heuristics such as "Copilot-owned PR" are too narrow for future automation targets like Renovate. The repository needs a deterministic, label-gated workflow that can react to failed CI on explicitly opted-in pull requests and hand structured failure context directly to Copilot.

## What Changes

- Add a GitHub Agentic Workflow that listens for failed CI workflow runs, resolves the associated pull request, and only activates for same-repository pull requests labeled `auto-fix`.
- Add deterministic pre-activation gates that verify the source workflow, confirm the run failed for a pull request, classify supported failure types, and skip unsupported, duplicate, or forked cases before agent reasoning starts.
- Instruct Copilot to remediate supported failures directly on the pull request branch: lint failures should be fixed and pushed, while acceptance test failures should be analyzed with version-specific context and fixed only when there is a clear path.
- Require the workflow to leave clear pull request feedback when remediation is skipped, when analysis concludes there is no safe fix, or when maintainers need to inspect the failure manually.
- Document the repository and Copilot settings needed for follow-up CI execution after agent-authored pushes.

## Capabilities

### New Capabilities
- `ci-pr-auto-fix`: label-gated Copilot remediation for failed CI on pull requests with deterministic failure classification and PR feedback

### Modified Capabilities
<!-- None. -->

## Impact

- New authored workflow source under `.github/workflows-src/` and compiled workflow artifacts under `.github/workflows/`
- CI workflow integration points needed to identify supported failed runs and produce stable remediation context
- Repository automation permissions, safe outputs, and optional CI rerun configuration for agent-authored pushes
- Maintainer-facing documentation for the `auto-fix` label and operational expectations
- Pull request feedback behavior for skipped, fixed, and analysis-only remediation outcomes
