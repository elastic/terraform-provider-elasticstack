## Why

When the changelog generator runs in release mode for a `prep-release-X.Y.Z` branch, it inserts the new `## [X.Y.Z] - <date>` section *after* the existing `## [Unreleased]` section instead of replacing it. The result is a release PR that contains both the original Unreleased entries and a near-identical copy under the new version heading, forcing maintainers to manually delete the duplicate before merging.

Observed in [PR #2857 — Prepare 0.15.0 release](https://github.com/elastic/terraform-provider-elasticstack/pull/2857): the `## [Unreleased]` section was preserved verbatim and `## [0.15.0] - 2026-05-11` was appended below it with the same ~30 bullets.

A "release cut" semantically means the Unreleased work has shipped under a concrete version. Keeping `[Unreleased]` populated alongside the new version is incoherent.

## What Changes

- **BREAKING (workflow behavior)**: In explicit release mode, `rewriteChangelogSection` SHALL remove any existing `## [Unreleased]` section and replace it with the new `## [X.Y.Z] - <date>` section. Behavior when the Unreleased section is absent (already-cut state, re-run) is unchanged for the existing-target-version branch.
- Re-runs of release mode (target version section already present) SHALL also strip any lingering `## [Unreleased]` section so the changelog converges to a single, correct shape regardless of how many times the workflow executes against the same branch.
- Update the three tests that currently assert the duplicating behavior to assert replacement instead:
  - `.github/workflows-src/lib/changelog-rewriter.test.mjs::rewriteChangelogSection release inserts after Unreleased when section missing`
  - `.github/workflows-src/lib/changelog-engine.test.mjs::runChangelogRenderAndWrite release inserts section after Unreleased`
  - `.github/workflows-src/lib/changelog-engine.test.mjs::runChangelogRenderAndWrite release with zero PRs writes header-only section`
- Regenerate the compiled `.github/workflows/changelog-generation.yml` if any inline workflow logic carries the same assumption (template lives under `.github/workflows-src/`).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-changelog-generation`: Tighten the "Explicit release mode updates only the targeted release section" requirement to specify that the new versioned section replaces the Unreleased section (header + body) and that any pre-existing Unreleased section is removed on every release-mode run.

## Impact

- **Code**: `.github/workflows-src/lib/changelog-rewriter.js` — change the `targetStart === -1 && mode === 'release'` branch (and complement the existing `targetStart !== -1` branch in release mode) to drop the Unreleased section. No other module depends on the existing behavior.
- **Tests**: Three test cases under `.github/workflows-src/lib/` flip their expectations. New tests cover the "re-run with both Unreleased and `[X.Y.Z]` present" case and the existing "no Unreleased at all" case.
- **Compiled workflow**: `.github/workflows/changelog-generation.yml` regenerated via `scripts/compile-workflow-sources/main.go` if the template changes; otherwise unaffected (rewriter is consumed as a module).
- **Operational**: Eliminates a manual cleanup step for every release PR. No effect on unreleased mode, the `generated-changelog` PR, or PR-body parsing.
- **Risk**: Low. The duplicating behavior produces only review noise today; replacing it is information-preserving (the duplicate content already lives in the merged PR history that fed the section). No downstream tooling parses Unreleased and X.Y.Z simultaneously.
