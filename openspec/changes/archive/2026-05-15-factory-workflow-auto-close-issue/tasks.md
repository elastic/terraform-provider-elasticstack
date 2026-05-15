## 1. Update authored workflow sources

- [x] 1.1 Locate the repository-authored `change-factory` workflow source and set `safe-outputs.create-pull-request.auto-close-issue` to `false`
- [x] 1.2 Locate the repository-authored `reproducer-factory` workflow source and set `safe-outputs.create-pull-request.auto-close-issue` to `false`
- [x] 1.3 Verify both authored workflows still express `Related to #N` linkage and do not rely on GitHub closing keywords

## 2. Regenerate and verify workflow artifacts

- [x] 2.1 Regenerate the compiled workflow artifacts from the authored sources
- [x] 2.2 Verify the generated `.github/workflows/change-factory-issue.md` and `.github/workflows/reproducer-factory-issue.md` preserve `auto-close-issue: false`
- [x] 2.3 Verify the generated lockfiles also reflect the non-closing PR policy where applicable

## 3. Keep requirements aligned

- [x] 3.1 Update or confirm workflow requirements/specs to state that `change-factory` PR creation disables automatic issue-closing references
- [x] 3.2 Update or confirm workflow requirements/specs to state that `reproducer-factory` PR creation disables automatic issue-closing references
- [x] 3.3 Validate the OpenSpec change artifacts before implementation begins
