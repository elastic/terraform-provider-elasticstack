# Delta Spec: Makefile Workflows

On sync or archive, this content is intended to land in `openspec/specs/makefile-workflows/spec.md`.

Implementation: [`Makefile`](../../../../Makefile)

## MODIFIED Requirements

### Requirement: Documentation, workflow, and code generation (REQ-038–REQ-042)

The `docs-generate` target SHALL regenerate Terraform provider website/markdown documentation using **HashiCorp `terraform-plugin-docs`** (`tfplugindocs`) for provider name `terraform-provider-elasticstack`. `docs-generate` SHALL read the Terraform CLI version from the repository root `.terraform-version` file and SHALL pass that exact version to `tfplugindocs` via `--tf-version`, so documentation generation does not depend on whichever Terraform binary happens to be installed locally. The `workflow-generate` target SHALL regenerate the checked-in GitHub workflow artifacts from the repository-authored workflow sources, and it SHALL run only when explicitly requested. Aggregate targets such as `gen`, `lint`, `check-lint`, and `build` SHALL NOT depend on `workflow-generate`. The `workflow-test` target SHALL run the repository tests that cover workflow source generation. The `hook-test` target SHALL run `node --test .agents/hooks/*.test.mjs`. The `check-workflows` target SHALL verify that generated workflow artifacts are up to date without regenerating them. The `gen` target SHALL run documentation generation and `go generate` for the repository.

#### Scenario: Docs generation

- GIVEN `make docs-generate`
- WHEN it succeeds
- THEN `tfplugindocs` SHALL have regenerated provider docs to match the current schema
- AND the Terraform CLI version used for schema extraction SHALL come from `.terraform-version`
