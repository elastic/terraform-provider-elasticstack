## 1. Establish `.terraform-version` as the shared Terraform source of truth

- [ ] 1.1 Add a root `.terraform-version` file pinned to the current latest stable Terraform release.
- [ ] 1.2 Update `docs-generate` to read `.terraform-version` and pass that value to `tfplugindocs generate --tf-version ...`.
- [ ] 1.3 Ensure related aggregate targets (`gen`, `lint`, `check-docs`, `check-lint`) continue to exercise the pinned docs-generation path through `docs-generate`.

## 2. Align CI with `.terraform-version`

- [ ] 2.1 Update the lint/docs validation workflow configuration so the Terraform setup step reads and uses the same `.terraform-version` value as local docs generation.
- [ ] 2.2 Regenerate workflow artifacts if the source workflow templates are the canonical edit point.
- [ ] 2.3 Verify CI source/tests covering workflow generation still pass after the Terraform version source change.

## 3. Update requirements, contributor docs, and Renovate expectations

- [ ] 3.1 Update the relevant OpenSpec requirements (at minimum `openspec/specs/makefile-workflows/spec.md`) to state that `docs-generate` uses the Terraform CLI version pinned in `.terraform-version` rather than a developer-installed version.
- [ ] 3.2 Update `dev-docs/high-level/documentation.md` to document the `.terraform-version` policy for docs generation and where it is configured.
- [ ] 3.3 Update any additional contributor guidance that mentions docs generation if needed for consistency.
- [ ] 3.4 Verify the repository's `renovate.json` configuration allows built-in `.terraform-version` updates, and adjust only if necessary.

## 4. Verify behavior

- [ ] 4.1 Run targeted validation for docs generation (for example `make docs-generate` and/or `make check-docs`) and confirm docs generation succeeds while reading the pinned version from `.terraform-version`.
- [ ] 4.2 Run workflow validation/tests required by the repo (`make check-workflows` / `make workflow-test`) if workflow sources changed.
- [ ] 4.3 Run the relevant aggregate validation (`make check-lint` or an equivalent targeted subset) to confirm the deterministic docs-generation path integrates cleanly.
