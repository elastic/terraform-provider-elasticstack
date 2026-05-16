## Why

Two gaps in the changelog generation workflow require manual fixup after each release:

1. The generator never updates the Markdown reference link table at the bottom of `CHANGELOG.md` in release mode — confirmed by commit `313a3526` which patched the missing v0.15.0 entries.
2. The PRs created or managed by the changelog workflow (`generated-changelog` and release prep PRs) are not labelled `no-changelog`, meaning the generator will attempt to parse them as feature PRs on the next run and fail.

## What Changes

- Add `rewriteLinkTable(content, targetVersion, previousTag)` to `changelog-rewriter.js` that updates the `[Unreleased]:` compare URL and inserts the new `[x.y.z]:` entry in release mode.
- Pass `previousTag` through to `runChangelogRenderAndWrite` in `changelog-engine-factory.js` and call `rewriteLinkTable` after `rewriteChangelogSection`.
- Wire `previousTag: ctx.previousTag` into the `runChangelogRenderAndWrite` call in `run-changelog-engine.inline.js` and `changelog-engine-workflow.js`.
- Apply the `no-changelog` label to the `generated-changelog` PR (on create and on update) in `manageUnreleasedPR`.
- Apply the `no-changelog` label to the release prep PR in `refreshReleasePR` after it is located.
- Add unit tests for `rewriteLinkTable` in `changelog-rewriter.test.mjs`.
- Update existing release-mode tests in `changelog-engine.test.mjs` to assert the link table is updated.
- Add tests for `no-changelog` label application in `changelog-pr-management.test.mjs`.
- Add delta spec requirements to `ci-changelog-generation`.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `ci-changelog-generation`: Add two requirements — (1) release-mode changelog generation updates the Markdown reference link table at the bottom of `CHANGELOG.md`; (2) changelog-generation PRs (`generated-changelog` and release prep) are labelled `no-changelog`.

## Impact

- `changelog-rewriter.js` — new exported function
- `changelog-engine-factory.js` — signature change to `runChangelogRenderAndWrite` (backward-compatible; `previousTag` defaults to `''`)
- `run-changelog-engine.inline.js`, `changelog-engine-workflow.js` — pass `previousTag` at call site
- `changelog-pr-management.js` — apply `no-changelog` label in `manageUnreleasedPR` and `refreshReleasePR`
- `changelog-rewriter.test.mjs`, `changelog-engine.test.mjs`, `changelog-pr-management.test.mjs` — new and updated tests
- `openspec/specs/ci-changelog-generation/` — delta spec for both new requirements
