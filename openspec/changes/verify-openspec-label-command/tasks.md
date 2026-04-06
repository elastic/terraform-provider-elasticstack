## 1. Update workflow trigger and cleanup semantics

- [x] 1.1 Change `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` to use `label_command` for `verify-openspec`, restricted to pull requests.
- [x] 1.2 Remove source-workflow logic, outputs, permissions, and prompt instructions that only support explicit label verification or `remove-labels` cleanup.

## 2. Regenerate and align generated artifacts

- [x] 2.1 Regenerate the compiled workflow artifacts for `openspec-verify-label` so the committed `.md` and `.lock.yml` match the updated source.
- [x] 2.2 Review the generated workflow for expected trigger, automatic label-removal handling, and any permission changes introduced by compilation.

## 3. Validate requirements coverage

- [x] 3.1 Confirm the implementation matches the updated `ci-aw-openspec-verification` requirements for `label_command`, automatic cleanup, and skipped-run behavior.
- [x] 3.2 Run the relevant OpenSpec validation/check command(s) for the change and resolve any issues needed to make the workflow apply-ready.
