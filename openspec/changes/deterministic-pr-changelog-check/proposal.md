## Why

The current PR changelog authoring workflow uses a GitHub Agentic Workflow (gh-aw) that triggers after `Build/Lint/Test` completes and invokes an LLM agent to draft missing `## Changelog` sections. This introduces non-determinism, latency (feedback arrives after CI, not on PR open), unnecessary LLM cost, and significant workflow complexity (gh-aw compilation, lock files, PR body mutation). The goal is to replace it with a purely deterministic check that fails fast with actionable feedback.

## What Changes

- **Remove** `.github/workflows/pr-changelog-authoring.md` (compiled gh-aw workflow)
- **Remove** `.github/workflows/pr-changelog-authoring.lock.yml` (gh-aw lock file)
- **Remove** `.github/workflows-src/pr-changelog-authoring/` (source template and inline scripts)
- **Remove** the `pr-changelog-authoring` entry from `.github/workflows-src/manifest.json`
- **Create** `.github/workflows/pr-changelog-check.yml` — a plain GitHub Actions workflow triggered by `pull_request_target` that validates the changelog section and fails with a PR comment if it is missing or malformed

The workflow no longer drafts changelog sections on behalf of authors. Authors must supply the section themselves; the workflow tells them why it failed if they do not.

## Capabilities

### New Capabilities

_(none — this change replaces an existing capability)_

### Modified Capabilities

- `ci-pr-changelog-authoring`: Requirements change substantially — trigger moves from `workflow_run` to `pull_request_target`, the agent authoring phase is removed entirely, and failing checks now post a PR comment with the failure reason instead of drafting a section.

## Impact

- `.github/workflows/pr-changelog-authoring.md` and `.lock.yml` — deleted
- `.github/workflows-src/pr-changelog-authoring/` — deleted (source template, `resolve-pr.inline.js`, `validate-pr-changelog.inline.js`)
- `.github/workflows-src/manifest.json` — one entry removed
- `.github/workflows/pr-changelog-check.yml` — created
- `test.yml` — no changes required (the `workflow_run` dependency was in `pr-changelog-authoring`, not `test.yml`)
- Branch protection — `PR Changelog Check` is not yet a required status check; enabling it is a follow-on step after deployment
