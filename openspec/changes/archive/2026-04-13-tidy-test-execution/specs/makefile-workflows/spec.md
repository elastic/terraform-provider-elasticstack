## MODIFIED Requirements

### Requirement: Unit tests (REQ-022)

The `test` target SHALL run all repository unit-style test suites. It SHALL run Go unit tests for `TEST` with a bounded wall-clock timeout, fixed `-count`, and repository-chosen parallelism; extra arguments MAY be supplied via `TESTARGS`. It SHALL also run workflow source generation tests and hook JavaScript tests so `make test` provides a single entry point for unit-level verification.

#### Scenario: Go unit tests

- GIVEN `make test`
- WHEN the Go unit-test portion runs
- THEN packages under `TEST` SHALL have been executed under the configured timeout, count, and parallelism constraints

#### Scenario: Aggregate unit-style test coverage

- GIVEN `make test`
- WHEN the target completes successfully
- THEN `workflow-test` SHALL have been executed
- AND hook JavaScript tests SHALL have been executed

### Requirement: Documentation, workflow, and code generation (REQ-038–REQ-042)

The `docs-generate` target SHALL regenerate Terraform provider website/markdown documentation using **HashiCorp `terraform-plugin-docs`** (`tfplugindocs`) for provider name `terraform-provider-elasticstack`. The `workflow-generate` target SHALL regenerate the checked-in GitHub workflow artifacts from the repository-authored workflow sources, and it SHALL run only when explicitly requested. Aggregate targets such as `gen`, `lint`, and `build` SHALL NOT depend on `workflow-generate`. The `workflow-test` target SHALL run the repository tests that cover workflow source generation. The `hook-test` target SHALL run `node --test .agents/hooks/*.test.mjs`. The `check-workflows` target SHALL verify that generated workflow artifacts are up to date without regenerating them. The `gen` target SHALL run documentation generation and `go generate` for the repository.

#### Scenario: Docs generation

- GIVEN `make docs-generate`
- WHEN it succeeds
- THEN `tfplugindocs` SHALL have regenerated provider docs to match the current schema

#### Scenario: Manual workflow generation

- GIVEN `make workflow-generate`
- WHEN it succeeds
- THEN the checked-in workflow artifacts SHALL be regenerated from the repository-authored workflow sources

#### Scenario: Hook test target

- GIVEN `make hook-test`
- WHEN the target runs
- THEN Node's test runner SHALL execute `.agents/hooks/*.test.mjs`

#### Scenario: Workflow drift check without regeneration

- GIVEN generated workflow sources are out of date with their checked-in templates
- WHEN `make check-workflows` runs
- THEN it SHALL fail without regenerating workflow artifacts
