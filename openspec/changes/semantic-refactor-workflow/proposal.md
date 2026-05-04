## Why

The repository can benefit from a scheduled Agentic Workflow that finds actionable Go function organization and extraction opportunities before they accumulate into larger maintenance work. The upstream `gh-aw` semantic function refactor workflow provides a useful baseline, but it needs repository-specific adaptation for LiteLLM execution, generated workflow sources, and bounded issue creation.

## What Changes

- Add a new `semantic-function-refactor` GitHub Agentic Workflow derived from `https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md`.
- Author the workflow under `.github/workflows-src/`, generate checked-in workflow artifacts under `.github/workflows/`, and compile the lock file through the existing workflow generation path.
- Configure the workflow to run through the repository's LiteLLM-backed Claude engine configuration instead of the upstream direct engine setting.
- Replace upstream's single-issue output and existing-issue closing behavior with deterministic issue-slot gating for the `semantic-refactor` bucket.
- Allow the workflow to create up to three open `semantic-refactor` issues by counting existing open issues with that label before agent activation and giving the agent only the remaining issue slots.
- Define the semantic refactor analysis scope, issue creation rules, and actionable issue content contract.

## Capabilities

### New Capabilities
- `ci-semantic-refactor-workflow`: scheduled and manually triggered semantic Go function refactor analysis that opens bounded, actionable `semantic-refactor` issues from a generated GH AW workflow

### Modified Capabilities
- None.

## Impact

- New authored workflow source under `.github/workflows-src/semantic-function-refactor/`.
- New generated workflow artifacts under `.github/workflows/`, including the compiled `.lock.yml`.
- `.github/workflows-src/manifest.json` registration for the new workflow source.
- Reuse of the existing workflow-source compiler, `gh aw compile`, issue-slot helper, and workflow test patterns.
- New or updated workflow-source tests validating the `semantic-refactor` bucket, issue cap, LiteLLM engine configuration, generated artifacts, and prompt contract.
