## 1. Workflow source updates

- [x] 1.1 Change `DUPLICATE_LINKAGE_MODE` from `'github-keywords'` to `'related-literal'` in `.github/workflows-src/change-factory-issue/intake-constants.js`
- [x] 1.2 Update the JSDoc comment in the same file to reflect that the workflow uses literal `Related to #N` linkage rather than GitHub closing keywords
- [x] 1.3 In `.github/workflows-src/change-factory-issue/workflow.md.tmpl`, replace both prompt-body references to `` `Closes #${{ github.event.issue.number }}` `` with `` `Related to #${{ github.event.issue.number }}` `` and update the surrounding sentences so they explain why `Related to` is used (the proposal does not resolve the issue; it produces an OpenSpec change that still needs implementation)
- [x] 1.4 Update `.github/workflows-src/lib/change-factory-issue.test.mjs` so the duplicate-PR fixtures, mode assertion, and prompt-contract assertion match the new `'related-literal'` mode (closing keywords no longer match; `Related to #N` does)

## 2. Spec deltas

- [x] 2.1 The MODIFIED Requirements for the `Workflow suppresses duplicate linked pull requests` requirement and the `Agent creates exactly one linked OpenSpec proposal pull request` requirement are authored in this change's delta spec at `specs/ci-change-factory-issue-intake/spec.md`. No manual edit of the canonical `openspec/specs/ci-change-factory-issue-intake/spec.md` is required during implementation; the delta will be applied to the canonical spec when the change is archived.

## 3. Build and validation

- [x] 3.1 Run `make workflow-generate` to recompile `.github/workflows/change-factory-issue.md` and `.github/workflows/change-factory-issue.lock.yml`
- [x] 3.2 Run `make check-openspec` to confirm the modified spec passes structural validation
- [x] 3.3 Run `make check-workflows` to confirm no generated workflow markdown is stale
- [x] 3.4 Run `make workflow-test` to confirm the existing factory-issue-shared library tests still pass (the lib was modified separately by the reproducer-factory-workflow change; no further lib changes are needed here)
