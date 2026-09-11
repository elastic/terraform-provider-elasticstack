## Why

The Provider CI acceptance-test matrix (`strategy.matrix.version` plus `include` overrides in `.github/workflows/provider.yml`) is hand-maintained. Updating it for a new patch, a new minor, or a snapshot-to-GA promotion is a manual chore that lags, so CI stops testing the latest published patch of each 8.x/9.x minor. At research time (2026-09-10) the committed matrix had already drifted from the latest GA tags: `8.19.17` vs `8.19.21`, `9.3.6` vs `9.3.8`, `9.4.2` vs `9.4.6`, `9.5.0` vs `9.5.3`.

There is no single public "all product versions" API. The closest deterministic, complete, public source is `elastic/elasticsearch` git tags (`vX.Y.Z`) for GA patches per minor, plus `snapshots.elastic.co/latest/master.json` for the unreleased next-minor snapshot label. A live fetch at Provider CI start would make matrix membership non-reproducible for a given commit, so the desired list must be computed on a schedule and pinned in git — mirroring the existing `changelog-generation` scheduled-PR pattern (`.github/workflows/changelog-generation.yml`, `ci-changelog-generation` capability) rather than hand-edited YAML.

## What Changes

- Add a new scheduled + `workflow_dispatch` workflow, `.github/workflows/version-matrix-generation.yml`, backed by a new `scripts/version-matrix` Go tool, that computes the desired stack-version list (latest GA patch per 8.x/9.x minor from `elastic/elasticsearch` tags, plus the current `snapshots.elastic.co/latest/master.json` SNAPSHOT label) and pins it to a new checked-in JSON artifact, `.github/versions/acceptance-test-matrix.json`.
- When the computed list differs from the pinned artifact, the workflow commits the update to a stable branch (`acceptance-test-version-matrix`), re-triggers CI with `GH_AW_CI_TRIGGER_TOKEN` (same pattern as `changelog-generation.yml`), and opens or updates a single standing pull request whose body asks to make `main` match the full computed list (all-or-nothing — a newly-red GA version holds the whole delta rather than being silently dropped). When the computed list is unchanged, the workflow no-ops without touching any existing open PR.
- Change `.github/workflows/provider.yml`'s `test` job to load `strategy.matrix.version` from the pinned JSON artifact (via `fromJson()` from a preceding job) instead of a hardcoded YAML list, and remove the `include:` overrides for `runner`/`fleetImage`. Per-version environment rules (Docker Hub fleet image for 8.0–8.1, `ubuntu-22.04` runner for 8.0–8.4, forced synthetics install for 8.14–8.17) become numeric major.minor range checks — not `startsWith(matrix.version, '8.1.')`, which also matches `8.10.x`–`8.19.x`. The `load-matrix` job (or equivalent) attaches the derived flags so the `test` job does not re-derive them with unsafe string prefixes.
- Teach `.github/scripts/workflows/lib/classify-changes.js` (and its unit tests) that `.github/versions/acceptance-test-matrix.json` is provider-impacting, so a version-bump PR runs the full matrix against the *proposed* list instead of being classified as a non-impacting `.github/`-only change.
- Add a `version-matrix` auto-approve category to `scripts/auto-approve` (same shape as `generated-changelog`: same-repo PR, head branch exactly `acceptance-test-version-matrix`, every commit authored by `github-actions[bot]`, changed files limited to exactly `.github/versions/acceptance-test-matrix.json`), so the standing PR is auto-approved once Provider Gate is green. Auto-merge is explicitly out of scope; merging stays with existing humans/rulesets.
- Rewrite `acceptance-test-isolation` REQ-ACC-002 from a literal `"9.4.2"` pin into a living invariant: the pinned artifact contains exactly one `9.4.x` GA entry (the latest published 9.4 patch). Drop historical scenarios that narrate the 9.4.0 → 9.4.2 one-off.

## Capabilities

### New Capabilities

- `ci-version-matrix-generation`: Defines the scheduled/`workflow_dispatch` workflow and the `scripts/version-matrix` Go engine that compute the desired GA+SNAPSHOT stack-version list, pin it to `.github/versions/acceptance-test-matrix.json`, and open/update or no-op the standing version-bump pull request.

### Modified Capabilities

- `ci-build-lint-test`: The acceptance-test matrix's `version` list is loaded from the pinned versions artifact rather than hardcoded in the workflow; per-version environment overrides become numeric major.minor range checks (evaluated in `load-matrix`, not `startsWith` prefixes) instead of an exact-version `include:` list; change-classification treats the pinned versions artifact as provider-impacting; snapshot-to-GA promotion is now performed by the generator rewriting the pinned artifact rather than by a human editing YAML.
- `ci-pr-auto-approve`: Adds the `version-matrix` category so the standing version-bump PR is auto-approved once Provider Gate is green, without adding any merge behavior.
- `acceptance-test-isolation`: REQ-ACC-002 changes from pinning the exact string `"9.4.2"` to a living invariant against `.github/versions/acceptance-test-matrix.json` (exactly one latest `9.4.x` GA entry), so it survives automated patch bumps.

## Impact

- **New code**: `scripts/version-matrix/` (Go engine: GA-tag discovery, SNAPSHOT-label lookup, compose-stack image probe with promotion fallback, pinned-artifact diffing/rewriting).
- **New workflow**: `.github/workflows/version-matrix-generation.yml`.
- **New pinned artifact**: `.github/versions/acceptance-test-matrix.json` (git-tracked, generated-but-checked-in, analogous to `CHANGELOG.md`'s `## [Unreleased]` section).
- **Modified workflow**: `.github/workflows/provider.yml` (`test` job matrix source, `include:` removal, per-version conditions rewritten as range expressions).
- **Modified scripts**: `.github/scripts/workflows/lib/classify-changes.js` (+ tests), `scripts/auto-approve/` (`evaluator.go` + tests) for the new category.
- **Modified specs**: `openspec/specs/ci-build-lint-test/spec.md`, `openspec/specs/ci-pr-auto-approve/spec.md`, `openspec/specs/acceptance-test-isolation/spec.md`.
- **Out of scope**: auto-merge of the version-bump PR, computing the matrix at Provider CI start (unpinned), tracking 7.x/10.x or every patch of a minor, using artifacts-api/GCS `releases/current`/Cloud stack-versions/Docker tag listing as the version catalog, chasing SNAPSHOT build IDs, and any change to acceptance-test sharding, Fleet bootstrap, or the existing snapshot-failure PR comment beyond what the range-rule rewrite requires.
- **Backward compatibility**: additive automation plus a mechanical rewrite of how the matrix is sourced; the acceptance-test matrix's effective version *shape* (latest patch per 8.x/9.x minor plus one SNAPSHOT) is unchanged.
