## 1. Update prep-release workflow

- [x] 1.1 Add `--label no-changelog` to the `gh pr create` call in the "Create release PR" step of `.github/workflows/prep-release.yml`
- [x] 1.2 Add a new step after "Check if release PR already exists" that runs `gh pr edit --add-label no-changelog` when the PR already exists (`steps.pr-check.outputs.EXISTS == 'true'`)

## 2. Update spec

- [x] 2.1 Sync the delta spec (`openspec/changes/release-pr-no-changelog-label/specs/ci-release-pr-preparation/spec.md`) into the canonical spec at `openspec/specs/ci-release-pr-preparation/spec.md`
