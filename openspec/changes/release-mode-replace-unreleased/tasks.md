## 1. Rewriter behavior change

- [ ] 1.1 In `.github/workflows-src/lib/changelog-rewriter.js`, update `rewriteChangelogSection` so that in release mode, any existing `## [Unreleased]` section is removed when the new versioned section is written, regardless of whether the target `## [x.y.z]` heading already exists.
- [ ] 1.2 Preserve current behavior for unreleased mode (`mode === 'unreleased'`) and for the "no Unreleased and no target version" prepend path.
- [ ] 1.3 Keep the rewriter line-based; do not introduce a markdown parser dependency.

## 2. Tests

- [ ] 2.1 Invert `rewriteChangelogSection release inserts after Unreleased when section missing` in `.github/workflows-src/lib/changelog-rewriter.test.mjs` to assert that the resulting changelog contains `## [x.y.z]` but no `## [Unreleased]` heading, and that the `## [x.y.z]` block sits where Unreleased used to sit (above the next versioned section).
- [ ] 2.2 Invert `runChangelogRenderAndWrite release inserts section after Unreleased` in `.github/workflows-src/lib/changelog-engine.test.mjs` to assert the same outcome through the engine.
- [ ] 2.3 Invert `runChangelogRenderAndWrite release with zero PRs writes header-only section` in `.github/workflows-src/lib/changelog-engine.test.mjs` so the Unreleased heading is removed even when there are zero PR records (header-only section still replaces Unreleased).
- [ ] 2.4 Add a new test for the re-run case: initial file already contains both `## [Unreleased]` and `## [x.y.z]`; after release-mode rewrite, only `## [x.y.z]` remains.
- [ ] 2.5 Add a regression test mirroring PR #2857: fixture that includes a populated `## [Unreleased]` section identical to the rendered release body, run release-mode rewrite, and assert exactly one occurrence of each release bullet in the output.
- [ ] 2.6 Run `node --test .github/workflows-src/lib/*.test.mjs` and confirm the suite passes.

## 3. Compiled workflow + verification

- [ ] 3.1 Inspect `.github/workflows-src/changelog-generation/` and `.github/workflows/changelog-generation.yml` for any inline duplication of the rewriter logic; if found, update the template and regenerate via `scripts/compile-workflow-sources/main.go` (or the documented compile command).
- [ ] 3.2 Run `make build` and any project lint targets (`make check-lint` if available) to ensure no toolchain regressions.
- [ ] 3.3 Run `npx openspec validate --all` to confirm the change and modified delta spec are structurally valid.
- [ ] 3.4 Manually verify by checking out a recent `prep-release-*` branch (or reproducing the PR #2857 scenario locally) and re-running the engine to confirm the resulting `CHANGELOG.md` contains only the `## [x.y.z]` section with no leftover `## [Unreleased]` heading.

## 4. Sync / archive

- [ ] 4.1 Once merged, sync the delta spec into `openspec/specs/ci-changelog-generation/spec.md` via the **openspec-sync-specs** skill (or archive the change with **openspec-archive-change**, per project convention).
