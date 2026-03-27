## 1. Update the workflow contract

- [ ] 1.1 Update `.github/workflows/openspec-verify-label.md` so the review job provisions Node from `package.json` engines, Go from `go.mod`, and Terraform CLI in the same shape as the `lint` job bootstrap.
- [ ] 1.2 Revise the workflow's repository setup instructions and job steps so the review workspace runs `make setup` after runtime provisioning instead of relying on `npm ci` alone.

## 2. Regenerate generated workflow artifacts

- [ ] 2.1 Recompile `.github/workflows/openspec-verify-label.lock.yml` from the markdown source with `gh aw compile`.
- [ ] 2.2 Inspect the regenerated lock file to confirm the review job includes the expected toolchain bootstrap and dependency-preparation steps.

## 3. Validate the change

- [ ] 3.1 Run the relevant OpenSpec validation checks for the new change artifacts.
- [ ] 3.2 Verify the updated review workflow no longer depends on the runner's default Go installation and that `make setup` completes successfully before agent-invoked repository commands run.
