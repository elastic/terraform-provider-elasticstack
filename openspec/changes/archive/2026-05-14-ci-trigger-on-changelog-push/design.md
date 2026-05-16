## Context

The changelog-generation and prep-release workflows push commits using the default `GITHUB_TOKEN`. GitHub intentionally suppresses push events from the default token to prevent recursive workflow triggers, which means required status checks on the resulting PRs never get satisfied unless someone manually re-triggers CI.

The GitHub Agentic Workflows (gh-aw) system already documents this limitation and provides a solution: after pushing with `GITHUB_TOKEN`, re-authenticate git with a PAT and push an empty commit. The empty-commit push fires a real push event, triggering CI. The `GH_AW_CI_TRIGGER_TOKEN` secret convention is already in use by the gh-aw agent workflows (change-factory, code-factory, openspec-verify-label).

Both changelog PRs and release PRs are squash-merged, so the extra empty commit has no lasting history impact.

## Goals / Non-Goals

**Goals:**
- Required status checks on changelog and release PRs are satisfied automatically after workflow pushes.
- Graceful degradation: if `GH_AW_CI_TRIGGER_TOKEN` is not configured, workflows still push successfully but emit a visible warning.
- Reuse the existing `GH_AW_CI_TRIGGER_TOKEN` secret convention already established by gh-aw workflows.

**Non-Goals:**
- Changing the checkout or API token (stays `GITHUB_TOKEN`).
- Changing commit author attribution (stays `github-actions[bot]` via `git config`).
- Using `workflow_dispatch` to trigger CI — this does not satisfy required status checks on PRs.
- Full PAT override for the entire push (Option A from exploration) — we prefer preserving bot author attribution.

## Decisions

### 1. Empty-commit push pattern (not remote re-auth or workflow_dispatch)

**Decision**: Push the real commit with `GITHUB_TOKEN`, then push an empty commit re-authenticated with the PAT.

**Alternatives considered**:
- **Re-auth remote + push real commit with PAT**: Changes committer identity to the PAT user. Simpler, but loses `github-actions[bot]` attribution on the push event.
- **`workflow_dispatch` API call**: Does not satisfy required status checks — GitHub only matches check runs triggered by `push` or `pull_request` events against PRs.
- **Full `github-token` override on safe outputs**: Only applicable to gh-aw agent workflows; these are deterministic workflows.

**Rationale**: The empty-commit pattern is the gh-aw recommended approach. It preserves bot authorship on the substantive commit while using the PAT identity only for the CI-triggering push. Since both PR types are squash-merged, the extra commit is invisible in final history.

### 2. Reuse `GH_AW_CI_TRIGGER_TOKEN` secret name

**Decision**: Use `GH_AW_CI_TRIGGER_TOKEN` as the secret name, consistent with the gh-aw convention.

**Rationale**: The secret is already referenced in the gh-aw lock workflows. Using the same name avoids requiring a second PAT and secret configuration.

### 3. Empty commit message: `chore: trigger CI`

**Decision**: Use the minimal conventional-commit message `chore: trigger CI`.

**Rationale**: Clear purpose, matches the repo's commit style, and is easy to identify in branch history.

### 4. Re-auth pattern via `git remote set-url`

**Decision**: Re-authenticate by rewriting the remote URL before the empty-commit push, then skip restoring it (the runner is ephemeral).

```
SERVER_URL="${GITHUB_SERVER_URL#https://}"
git remote set-url origin "https://x-access-token:${GH_AW_CI_TRIGGER_TOKEN}@${SERVER_URL}/${GITHUB_REPOSITORY}.git"
git commit --allow-empty -m "chore: trigger CI"
git push origin HEAD:${BRANCH} [--force]
```

**Rationale**: Matches the pattern used by the gh-aw agent workflows for git push re-auth. No need to restore the original remote URL since the Actions runner is discarded after the job.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| `GH_AW_CI_TRIGGER_TOKEN` not configured | Guard with `if [ -n "${GH_AW_CI_TRIGGER_TOKEN}" ]` and emit `::warning::` — same behaviour as today (no CI trigger) but now visible |
| PAT expires or is revoked | Same graceful degradation — warning emitted, push still succeeds, CI just won't auto-trigger |
| Empty commit adds noise to branch history | Squash-merge makes this invisible in final history |
| Force-push timing (unreleased mode): two pushes in quick succession | The GITHUB_TOKEN push completes before the PAT push starts (sequential steps); GitHub processes them in order |
| Token leaked in logs via set-url | The `set-url` command itself doesn't echo the URL. The `git push` output may show the remote — GitHub Actions automatically masks known secrets in logs |
