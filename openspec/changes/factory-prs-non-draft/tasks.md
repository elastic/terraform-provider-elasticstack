## 1. Update change-factory workflow source

- [ ] 1.1 Add `draft: false` to the `safe-outputs.create-pull-request` block in `.github/workflows/change-factory-issue.md`.

## 2. Update reproducer-factory workflow source

- [ ] 2.1 Add `draft: false` to the `safe-outputs.create-pull-request` block in `.github/workflows/reproducer-factory-issue.md`.

## 3. Regenerate compiled lock files

- [ ] 3.1 Run `gh aw compile` (or `make workflow-generate`) for `change-factory-issue.md` to regenerate `.github/workflows/change-factory-issue.lock.yml`.
- [ ] 3.2 Run `gh aw compile` (or `make workflow-generate`) for `reproducer-factory-issue.md` to regenerate `.github/workflows/reproducer-factory-issue.lock.yml`.

## 4. Update OpenSpec delta specs

- [ ] 4.1 Add a new requirement to the `ci-change-factory-issue-intake` delta spec stating that proposal PRs SHALL be created as non-draft.
- [ ] 4.2 Add a new requirement to the `ci-reproducer-factory-issue-intake` delta spec stating that reproduction PRs SHALL be created as non-draft.

## 5. Validate

- [ ] 5.1 Run `make check-openspec` (or `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate factory-prs-non-draft --type change`) and confirm no errors.
