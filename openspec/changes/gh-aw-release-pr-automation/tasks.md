## 1. Changelog generator workflow

- [ ] 1.1 Add the authored GH AW changelog-generator workflow source under `.github/workflows-src/` and the compiled workflow artifacts under `.github/workflows/`.
- [ ] 1.2 Configure the changelog generator trigger modes for schedule, `workflow_dispatch`, and same-repository `prep-release-*` pull requests.
- [ ] 1.3 Add deterministic pre-activation steps that resolve the previous release tag, authoritative compare range, merged pull-request set, and target section mode (`Unreleased` vs concrete release section).

## 2. Changelog evidence and validation helpers

- [ ] 2.1 Implement the deterministic helper that builds the changelog evidence manifest in workflow memory, including PR metadata, coarse classification, and inclusion/exclusion rationale.
- [ ] 2.2 Implement the deterministic helper that validates the agent's changelog markdown and provenance against the evidence manifest.
- [ ] 2.3 Implement the deterministic helper that rewrites the full target section in `CHANGELOG.md` for both `Unreleased` mode and release-PR mode.

## 3. Agent prompt and changelog PR behavior

- [ ] 3.1 Author the GH AW prompt so the agent generates changelog output strictly from PR-level summaries and returns structured provenance for every changelog bullet.
- [ ] 3.2 Add singleton branch/PR orchestration for scheduled/manual changelog maintenance on branch `generated-changelog`.
- [ ] 3.3 Add release-PR branch update logic so `prep-release-*` pull requests receive regenerated concrete `## [x.y.z] - <date>` sections on their own branch.

## 4. Deterministic release preparation workflow

- [ ] 4.1 Add the deterministic release-preparation workflow that computes the target version from `major|minor|patch`, creates or reuses `prep-release-x.y.z`, and applies the simple version bump changes.
- [ ] 4.2 Add deterministic branch and pull-request orchestration for release-preparation PR creation or reuse.
- [ ] 4.3 Add the thin `Makefile` target that validates `BUMP` and dispatches the release-preparation workflow via `gh workflow run`.

## 5. CI and auto-approval integration

- [ ] 5.1 Extend `scripts/auto-approve/` and its tests with the narrowly scoped `generated-changelog` approval category.
- [ ] 5.2 Update `.github/workflows/test.yml` so same-repository `generated-changelog` PRs touching only `CHANGELOG.md` can reach the auto-approve path without running the full provider CI suite.
- [ ] 5.3 Enable auto-merge for the generated changelog PR once the lightweight validation path and approval policy succeed.

## 6. Verification and documentation

- [ ] 6.1 Add or update tests for version bumping, previous-tag selection, merged-PR evidence gathering, workflow-memory persistence/loading, provenance validation, changelog section rewriting, and the generated-changelog approval path.
- [ ] 6.2 Add prompt/fixture validation that confirms changelog output remains PR-based and rejects commit-level narration.
- [ ] 6.3 Update maintainer documentation in `dev-docs/high-level/contributing.md` to replace manual changelog maintenance and manual release-PR preparation with the new workflows.
- [ ] 6.4 Regenerate and verify workflow artifacts and any related workflow-source tests or checks required by the repository, then run the relevant OpenSpec validation/check commands.
