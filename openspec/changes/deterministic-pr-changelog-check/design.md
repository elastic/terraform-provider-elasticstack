## Context

The current `pr-changelog-authoring` workflow is a GitHub Agentic Workflow (gh-aw) that:
1. Triggers via `workflow_run` after `Build/Lint/Test` completes
2. Resolves the PR from the triggering `workflow_run` payload via API search (head_sha + branch lookup)
3. Skips fork PRs entirely (write access unavailable under `workflow_run` for forks)
4. Runs an LLM agent to draft and append a `## Changelog` section when one is missing

The replacement is a plain GitHub Actions workflow — no gh-aw framework, no agent, no PR body mutation.

## Goals / Non-Goals

**Goals:**
- Give PR authors immediate changelog feedback (on open, not after CI)
- Handle fork PRs without special-casing (full write access via `pull_request_target`)
- Remove all non-determinism and LLM involvement from the changelog gate
- Simplify the source footprint (no compiled templates, no lock files)

**Non-Goals:**
- Drafting changelog sections on behalf of authors — authors must supply the section
- Changing the changelog format (`Customer impact` / `Summary` / `### Breaking changes`)
- Adding `PR Changelog Check` as a required status check (follow-on step post-deployment)
- Modifying `test.yml` (the `workflow_run` dependency lived in `pr-changelog-authoring`)

## Decisions

### Trigger: `pull_request_target` over `pull_request`

`pull_request_target` runs in the context of the base repository for all PRs, including forks. This gives the workflow `pull-requests: write` permission unconditionally, enabling PR comments for fork contributors.

`pull_request` was rejected: fork PRs run with restricted permissions, making PR comments impossible for fork contributors.

`pull_request_target` security note: no code checkout happens in this workflow. The workflow only reads `context.payload.pull_request` (event metadata) and calls `github.rest.issues` to upsert comments. There is no untrusted code execution path.

### Single `actions/github-script` step

All logic (label check, parse, validate, comment upsert) lives in one `actions/github-script` step. No separate shell steps, no file I/O, no artifact upload/download.

PR data is read directly from `context.payload.pull_request` — no API call required to fetch the PR body or labels.

### Parser/validator: inline verbatim from deleted source files

`parseChangelogSectionFull` and `validateChangelogSectionFull` are copied inline unchanged from `.github/workflows-src/lib/pr-changelog-parser.js` — the canonical source of these functions. `validate-pr-changelog.inline.js` only included that file; it does not define the functions itself. The existing unit tests in `.github/workflows-src/lib/*.test.mjs` continue to cover this logic and are not affected by the workflow change.

### Comment upsert with hidden HTML marker

Failure and pass comments are identified by the marker `<!-- pr-changelog-check -->` embedded in the comment body. On each run:
- Find existing bot comment by `github-actions[bot]` login containing the marker
- **Fail**: update existing comment or create new one
- **Pass with existing comment**: update to a "check passed" message
- **Pass with no existing comment**: do nothing (no noise on first-valid push)

This avoids comment accumulation on repeated pushes and provides a clean audit trail.

### Plain `.yml` workflow, not gh-aw compiled

Since there is no agent, the gh-aw framework adds no value. A plain `.yml` file is directly readable, directly editable, and has no compilation step. The `manifest.json` entry and the entire `workflows-src/pr-changelog-authoring/` source directory are deleted.

## Risks / Trade-offs

- **Authors must write their own changelog section** → Mitigation: clear failure comment with format hint and link to docs; PR template should document the format
- **`pull_request_target` is a privileged trigger** → Mitigation: no code checkout, no shell execution of PR content; only metadata reads and REST API writes
- **`unlabeled` event is in the trigger list** → Removing the `no-changelog` label immediately re-runs the check; no additional push is required. This risk no longer applies.
- **Parser logic is inlined, not imported** → Unit tests in `lib/*.test.mjs` still cover the shared logic; the inline copy stays in sync because it's verbatim

## Migration Plan

1. Delete the three gh-aw files (`pr-changelog-authoring.md`, `.lock.yml`, `workflows-src/pr-changelog-authoring/`)
2. Remove the entry from `manifest.json`
3. Create `.github/workflows/pr-changelog-check.yml`
4. Verify the old `workflow_run`-based check no longer appears on new PRs
5. (Follow-on) Add `PR Changelog Check` as a required status check in branch protection

Rollback: restore the deleted files from git history and re-add the manifest entry.

## Open Questions

_(none — all design decisions resolved during exploration)_
