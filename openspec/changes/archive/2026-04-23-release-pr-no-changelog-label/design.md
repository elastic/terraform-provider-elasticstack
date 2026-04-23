## Context

The `prep-release.yml` workflow creates (or reuses) a pull request targeting `main` that bumps the provider version. The changelog generator workflow (`ci-changelog-generation`) runs on pull requests and may generate a changelog entry for this PR. Since the release-prep PR only contains a version bump (and changelog content is injected separately), it must carry the `no-changelog` label so the changelog generator knows to skip it.

The `gh` CLI (already available in the workflow) supports adding labels via `--label` on `gh pr create` and via `gh pr edit --add-label` for existing PRs.

## Goals / Non-Goals

**Goals:**
- Ensure the `no-changelog` label is present on the release PR whenever the workflow runs — whether the PR is newly created or already exists.

**Non-Goals:**
- Creating the `no-changelog` label in the repo (it must already exist; documented as a pre-condition).
- Modifying how the changelog generator identifies release PRs.

## Decisions

**Apply label at creation and on reuse** — `gh pr create` accepts `--label` directly, so new PRs get the label atomically. For reused PRs, a dedicated `gh pr edit --add-label` step is added immediately after the existing-PR check to ensure idempotency regardless of how the PR was originally created.

Applying the label separately (rather than only at creation) means the label will be present even when the workflow is rerun on an existing PR that lacked the label.

## Risks / Trade-offs

- **Label must exist in the repo**: `gh pr create --label no-changelog` fails if the label doesn't exist. → Pre-condition is documented in the spec; the label is expected to already be present in the repository.
- **Minimal blast radius**: Only two lines of the workflow change — one `--label` flag addition and one new `gh pr edit` step. No other workflow logic is affected.
