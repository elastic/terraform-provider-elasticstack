## 1. Replace bespoke label cleanup in the workflow source

- [ ] 1.1 Update `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` to declare `remove-labels` safe output for the triggering pull request, constrained to `verify-openspec`.
- [ ] 1.2 Revise the workflow prompt so the agent requests `remove-labels` cleanup as part of its terminal safe outputs and remove prompt text that assumes a separate completion cleanup phase.
- [ ] 1.3 Remove the dedicated `completion_cleanup` job and any inline-script references that exist only to remove `verify-openspec`.

## 2. Regenerate derived workflow artifacts

- [ ] 2.1 Recompile the generated `openspec-verify-label` workflow outputs so the committed lock file matches the updated markdown template.
- [ ] 2.2 Delete or stop compiling any obsolete helper artifact that existed only for the removed label-cleanup script path.

## 3. Validate the new cleanup contract

- [ ] 3.1 Run the relevant OpenSpec and workflow validation checks to confirm the delta spec, workflow source, and compiled outputs stay aligned.
- [ ] 3.2 Verify the updated workflow still has the permissions needed for review submission, push-to-PR-branch, and `remove-labels` cleanup without the old completion job.
