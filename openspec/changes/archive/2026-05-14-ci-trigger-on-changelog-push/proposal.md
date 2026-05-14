## Why

Pushes from the changelog-generation and prep-release workflows use the default `GITHUB_TOKEN`, which does not trigger downstream CI workflows. This means required status checks on the resulting PRs never get satisfied — reviewers see perpetually pending checks and must manually re-trigger CI. The GitHub Agentic Workflows system already documents this limitation and provides a `GH_AW_CI_TRIGGER_TOKEN` PAT convention for solving it via an empty-commit push pattern.

## What Changes

- After each `git push` in the changelog-generation and prep-release workflow templates, push an empty commit re-authenticated with the `GH_AW_CI_TRIGGER_TOKEN` secret to trigger CI.
- Emit a warning if the secret is not configured, preserving current behaviour (no CI trigger) but making the gap visible.
- Use `chore: trigger CI` as the empty commit message.

## Capabilities

### New Capabilities
- `ci-trigger-on-push`: Push an empty commit authenticated with a PAT after workflow pushes to trigger downstream CI and satisfy required status checks.

### Modified Capabilities
- `ci-changelog-generation`: After pushing the changelog commit (release or unreleased mode), the workflow will also push an empty commit with the CI trigger token.
- `ci-release-pr-preparation`: After pushing the release branch, the workflow will also push an empty commit with the CI trigger token.

## Impact

- **Workflow templates**: `.github/workflows-src/changelog-generation/workflow.yml.tmpl` (2 push sites), `.github/workflows-src/prep-release/workflow.yml.tmpl` (1 push site).
- **Compiled workflows**: `.github/workflows/changelog-generation.yml`, `.github/workflows/prep-release.yml` — regenerated via `go run ./scripts/compile-workflow-sources/main.go`.
- **Repository secret**: `GH_AW_CI_TRIGGER_TOKEN` must be configured (fine-grained PAT with `Contents: Read & Write`).
- **PR history**: Each workflow-triggered push will produce one additional empty commit. Since both changelog and release PRs are squash-merged, this has no lasting history impact.
