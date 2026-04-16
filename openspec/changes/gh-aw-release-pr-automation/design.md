## Context

Provider releases are currently prepared by manually updating `Makefile` and `CHANGELOG.md`, opening a release PR, and later tagging the merged result. The highest-friction part is changelog maintenance: contributors are expected to keep `## [Unreleased]` current by hand, and maintainers still need to review whether the accumulated entries match the actual user-facing change set.

The repository already uses GitHub Agentic Workflows for narrow, well-gated automation. That pattern is a good fit for changelog synthesis, but it is not necessary for the rest of release preparation. The clean split is therefore:

- an agentic changelog generator that owns `CHANGELOG.md` maintenance
- a deterministic release-preparation workflow that owns version bumps and release PR creation

This also resolves the branch-protection problem cleanly. The changelog generator never needs to write directly to `main`; it can maintain changelog updates through repository-owned branches and PRs.

## Goals / Non-Goals

**Goals:**
- Add a GH AW changelog generator with two modes:
  - scheduled/manual regeneration of the full `## [Unreleased]` section on branch `generated-changelog`
  - `pull_request` regeneration of the concrete `## [x.y.z] - <date>` section for `prep-release-*` branches
- Keep changelog generation proof-carrying: deterministic PR evidence in workflow memory, PR-only summaries, structured provenance, and deterministic validation.
- Add a deterministic release-preparation workflow that creates `prep-release-x.y.z` branches and PRs with simple version bump changes.
- Auto-approve and auto-merge generated changelog PRs once green under a tightly scoped policy.
- Adjust the main CI workflow so changelog-only generated PRs can reach auto-approve without paying for the full build/lint/test matrix.
- Keep the maintainer entrypoint simple through a thin `make` wrapper for the release-preparation workflow.

**Non-Goals:**
- Publishing the GitHub Release object or pushing the final semver tag in this change.
- Preserving manual contributor-maintained `Unreleased` entries after the new changelog generator is rolled out.
- Letting the agent discover release history from raw repository state without deterministic pre-activation context.
- Generating commit-by-commit release notes.
- Allowing broad bot auto-approval beyond the narrowly scoped generated-changelog PR rules.

## Decisions

### 1. Split changelog maintenance and release preparation into two workflows

Use two workflows with different trust and complexity profiles:

- `ci-changelog-generation`: GH AW, owns changelog synthesis
- `ci-release-pr-preparation`: deterministic, owns version bumping and release PR creation

Why:
- Changelog synthesis is the only part that materially benefits from agent reasoning.
- Release preparation becomes smaller, auditable, and easier to retry.
- The changelog generator can operate continuously instead of only at release time.

Alternatives considered:
- One large GH AW release workflow: rejected because it mixes semantic synthesis with otherwise deterministic plumbing.
- Deterministic release workflow that dispatches and waits on a separate changelog workflow: rejected because cross-workflow orchestration adds polling, state handoff, and failure complexity.

### 2. Recompute full target sections, not “since last run” deltas

The changelog generator should regenerate the full target section on each run from an authoritative range.

- `Unreleased` mode: rebuild `## [Unreleased]` from `last_release_tag..main`
- release-PR mode: rebuild the concrete release section for the triggering `prep-release-*` PR branch from the previous release tag to that branch head

Why:
- Reruns become safe and idempotent.
- Missed schedules do not create gaps.
- No persistent “last processed PR” cursor is required.

Alternatives considered:
- Incremental append based on the last successful run: rejected because it is harder to reason about and more fragile under skipped or failed runs.

### 3. Keep changelog synthesis proof-carrying and PR-based

Before the agent runs, build a deterministic release evidence manifest in workflow memory. The agent must consume that manifest, generate changelog text strictly from PR-level summaries, and return structured provenance for every bullet. Deterministic validation must reject unsupported output.

Why:
- It makes changelog generation auditable and rerunnable.
- It prevents hallucinated PR references or drift outside the authoritative range.
- It keeps the output aligned with the repository’s PR-linked changelog style.

Alternatives considered:
- Agent writes only markdown: rejected because it loses machine-checkable provenance.
- Commit-based summarization: rejected because it is noisier and diverges from the repository’s changelog conventions.

### 4. Store release evidence in the GH AW agent directory, not the git worktree

Persist the evidence manifest in the GH AW agent directory (`/tmp/gh-aw/agent/`) rather than in the checked-in worktree or in GH AW `repo-memory`.

The pre-activation job passes the serialized evidence JSON as a job output. A pre-agent-step (running in the agent's execution environment) deserializes it and writes it to `/tmp/gh-aw/agent/evidence.json`, where the agent and pre-agent helper scripts can access it.

Why:
- Generated context should not appear in changelog or release PR diffs.
- Ephemeral evidence keeps reruns and local branch state clean.
- The manifest is runtime support data, not a versioned source artifact.
- `repo-memory` was rejected because content written in the pre-activation job may not survive into the agent execution context; the agent directory is shared between pre-agent-steps and the agent itself.

Alternatives considered:
- Checked-in JSON support file: rejected because it pollutes repository diffs with workflow context.
- GH AW `repo-memory`: rejected because the write step runs in the pre-activation job and that content may not be available in the separate agent execution context. The agent directory (`/tmp/gh-aw/agent/`) is the right boundary.

### 5. Use a singleton changelog PR on branch `generated-changelog`

Scheduled/manual `Unreleased` updates should always use the same branch and PR.

Why:
- It provides a stable place for the generated `Unreleased` state to converge.
- Reruns naturally update the same PR instead of creating noise.
- It makes auto-approval and auto-merge rules straightforward.

Alternatives considered:
- One branch per run: rejected because it creates branch churn and duplicate PRs.

### 6. Auto-approve only the tightly scoped generated-changelog PR shape

Extend the existing `scripts/auto-approve/` policy with a new category that matches only:

- branch name exactly `generated-changelog`
- all commits authored by `github-actions[bot]`
- only `CHANGELOG.md` modified

Why:
- This is specific enough to be trustworthy.
- It allows the singleton changelog PR to auto-merge once green, keeping `main` fresh for release preparation.

Alternatives considered:
- Broad “github-actions PR” approval: rejected as too permissive.
- Human approval for changelog PRs: rejected because stale `main` would undermine the value of continuous changelog generation.

### 7. Adjust `Build/Lint/Test` so generated changelog PRs can reach auto-approve without full CI

The repository currently ignores pull requests whose only changed path is `CHANGELOG.md`, so the existing `auto-approve` job never runs for changelog-only generated PRs. The CI workflow needs a specific exception path for `generated-changelog` PRs: run enough deterministic gating to produce a successful `Test Validation`/auto-approve path, but skip the full provider build/lint/test jobs.

Why:
- It preserves the repository’s intent to avoid expensive CI for changelog-only changes.
- It still allows the auto-approve job to execute.

Alternatives considered:
- Remove `CHANGELOG.md` from `paths-ignore` for all PRs: rejected because it would broaden CI surface for ordinary changelog-only changes unnecessarily.

### 8. Keep release preparation deterministic and focused on version bumps

The release-preparation workflow should:

- run on `workflow_dispatch`
- compute the target version from `major|minor|patch`
- create or reuse `prep-release-x.y.z`
- make the simple version bump changes
- open or reuse the release PR

Then the changelog generator’s `pull_request` mode fills in the `x.y.z` changelog section for that PR branch.

Why:
- It keeps the release workflow small and easy to audit.
- It avoids embedding agent execution in the release-prep path.
- It lets changelog generation logic stay centralized in one workflow family.

Alternatives considered:
- Let the release-preparation workflow call the agent directly: rejected because it duplicates the changelog-generation path.

### 9. Keep the `Makefile` entrypoint as a dispatch-only wrapper

Add a Make target that validates the bump mode and invokes `gh workflow run` for the deterministic release-preparation workflow.

Why:
- It gives maintainers a repo-local command without duplicating workflow logic.
- It keeps local and hosted behavior aligned.

Alternatives considered:
- Local `make` target that performs release logic directly: rejected because it duplicates the hosted automation.

## Risks / Trade-offs

- [Generated changelog PRs could stall and leave `main` stale] -> Mitigation: keep the PR singleton, add narrow auto-approve rules, and enable auto-merge once the lightweight CI path is green.
- [The CI workflow exception for `generated-changelog` could accidentally widen auto-approve coverage] -> Mitigation: tie the behavior to branch name, commit authorship, and `CHANGELOG.md`-only diffs in both CI gating and approval policy.
- [PR classification remains partly semantic] -> Mitigation: constrain the agent to deterministic PR evidence, require PR-level output only, and validate provenance before mutating `CHANGELOG.md`.
- [Agent directory content is ephemeral and not persisted across runs] -> Mitigation: the evidence manifest is regenerated fresh on every run by the pre-activation job and written to the agent directory by a pre-agent-step, so no cross-run persistence is needed.
- [Release PR mode could drift from `Unreleased` mode] -> Mitigation: make both changelog modes share the same evidence schema, prompt contract, and validator.

## Migration Plan

1. Add the GH AW changelog generator, its evidence helpers, and its compiled workflow artifacts.
2. Add the deterministic release-preparation workflow and its thin `make` dispatch target.
3. Extend `scripts/auto-approve/` and `Build/Lint/Test` so generated changelog PRs can auto-approve and auto-merge safely.
4. Update `CHANGELOG.md` maintenance expectations and maintainer docs to retire manual contributor-managed `Unreleased` entries.
5. Roll out by letting the generated changelog PR converge on `main`, then use the new deterministic release-preparation workflow for the next release PR.

## Open Questions

- None for the initial proposal. Follow-on work such as post-merge tagging or GitHub Release publication can be proposed separately.
