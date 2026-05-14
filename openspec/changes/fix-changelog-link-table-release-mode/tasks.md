## 1. Add `rewriteLinkTable` to `changelog-rewriter.js`

- [ ] 1.1 Implement `rewriteLinkTable(content, targetVersion, previousTag)`: guard (no-op when either arg is empty), extract base URL from `[Unreleased]:` line via regex, update `[Unreleased]:` to `vNEW...HEAD`, insert `[NEW]: .../compare/vOLD...vNEW` after it (idempotent — skip if entry already exists)
- [ ] 1.2 Export `rewriteLinkTable` from `changelog-rewriter.js`

## 2. Wire `rewriteLinkTable` into the engine

- [ ] 2.1 Add `previousTag = ''` optional parameter to `runChangelogRenderAndWrite` in `changelog-engine-factory.js`
- [ ] 2.2 Call `rewriteLinkTable(updatedChangelog, targetVersion, previousTag)` after `rewriteChangelogSection` in `runChangelogRenderAndWrite`, writing the result to `CHANGELOG.md`
- [ ] 2.3 Pass `previousTag: renderOutcome`-context `previousTag` from `runChangelogEngine` into its internal `runChangelogRenderAndWrite` call

## 3. Wire `previousTag` in call sites

- [ ] 3.1 Pass `previousTag: ctx.previousTag` to `runChangelogRenderAndWrite` in `run-changelog-engine.inline.js`
- [ ] 3.2 Pass `previousTag: ctx.previousTag` to `runChangelogRenderAndWrite` in `changelog-engine-workflow.js`

## 4. Apply `no-changelog` label to changelog-generation PRs

- [ ] 4.1 In `manageUnreleasedPR`: after creating a new PR call `github.rest.issues.addLabels` to apply `no-changelog`
- [ ] 4.2 In `manageUnreleasedPR`: after updating an existing PR call `github.rest.issues.addLabels` to apply `no-changelog` (idempotent)
- [ ] 4.3 In `refreshReleasePR`: after locating the release prep PR call `github.rest.issues.addLabels` to apply `no-changelog`

## 5. Tests for `no-changelog` label application

- [ ] 5.1 Add test: `manageUnreleasedPR` creates PR and applies `no-changelog` label
- [ ] 5.2 Add test: `manageUnreleasedPR` updates existing PR and applies `no-changelog` label
- [ ] 5.3 Add test: `refreshReleasePR` applies `no-changelog` label to located release prep PR

## 6. Tests for `rewriteLinkTable`

- [ ] 6.1 Add test: standard release — updates `[Unreleased]:` URL and inserts `[NEW]:` entry
- [ ] 6.2 Add test: idempotent re-run — `[NEW]:` not duplicated when entry already present
- [ ] 6.3 Add test: no-op when `[Unreleased]:` line is absent
- [ ] 6.4 Add test: no-op when `previousTag` is empty string
- [ ] 6.5 Add test: no-op when `targetVersion` is empty string (unreleased mode guard)

## 7. Update engine-level release tests

- [ ] 7.1 Update existing release-mode tests in `changelog-engine.test.mjs` that write a changelog with a link table to assert that `[Unreleased]:` is updated and `[NEW]:` is inserted
