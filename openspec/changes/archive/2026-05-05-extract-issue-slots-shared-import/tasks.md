## 1. Create shared script

- [x] 1.1 Create `.github/workflows-src/lib/compute_issue_slots.inline.js` with the same content as the existing per-workflow copies (i.e. `//include: issue-slots.js` at the top, followed by the script body)

## 2. Update consumer templates

- [x] 2.1 Update `x-script-include:` in `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl` from `scripts/compute_issue_slots.inline.js` to `../lib/compute_issue_slots.inline.js`
- [x] 2.2 Update `x-script-include:` in `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` from `scripts/compute_issue_slots.inline.js` to `../lib/compute_issue_slots.inline.js`
- [x] 2.3 Update `x-script-include:` in `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` from `scripts/compute_issue_slots.inline.js` to `../lib/compute_issue_slots.inline.js`

## 3. Delete per-workflow script files

- [x] 3.1 Delete `.github/workflows-src/duplicate-code-detector/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [x] 3.2 Delete `.github/workflows-src/semantic-function-refactor/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [x] 3.3 Delete `.github/workflows-src/schema-coverage-rotation/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory

## 4. Validate

- [x] 4.1 Run `make workflow-generate` and confirm it succeeds with no errors
- [x] 4.2 Run `make check-workflows` and confirm all generated artifacts are up to date
- [x] 4.3 Confirm no `compute_issue_slots.inline.js` files remain under consumer `scripts/` directories
- [x] 4.4 Commit the change (nothing new to commit — regenerated outputs were already clean)
