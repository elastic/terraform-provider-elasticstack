## 1. Create shared script

- [ ] 1.1 Create `.github/workflows-src/lib/compute_issue_slots.inline.js` with the same content as the existing per-workflow copies (i.e. `//include: issue-slots.js` at the top, followed by the script body)

## 2. Update consumer templates

- [ ] 2.1 Update `x-script-include:` in `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl` from `scripts/compute_issue_slots.inline.js` to `../lib/compute_issue_slots.inline.js`
- [ ] 2.2 Update `x-script-include:` in `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` from `scripts/compute_issue_slots.inline.js` to `../lib/compute_issue_slots.inline.js`
- [ ] 2.3 Update `x-script-include:` in `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` from `scripts/compute_issue_slots.inline.js` to `../lib/compute_issue_slots.inline.js`

## 3. Delete per-workflow script files

- [ ] 3.1 Delete `.github/workflows-src/duplicate-code-detector/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [ ] 3.2 Delete `.github/workflows-src/semantic-function-refactor/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory
- [ ] 3.3 Delete `.github/workflows-src/schema-coverage-rotation/scripts/compute_issue_slots.inline.js` and the now-empty `scripts/` directory

## 4. Validate

- [ ] 4.1 Run `make workflow-generate` and confirm it succeeds with no errors
- [ ] 4.2 Run `make check-workflows` and confirm all generated artifacts are up to date
- [ ] 4.3 Confirm no `compute_issue_slots.inline.js` files remain under consumer `scripts/` directories
- [ ] 4.4 Commit the change
