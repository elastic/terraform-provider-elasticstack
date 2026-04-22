## MODIFIED Requirements

### Requirement: Build and lint jobs (REQ-007–REQ-008, REQ-031)

The `build` job SHALL run on `ubuntu-latest`, set up Go from `go.mod`, set up Node.js (24.x), run `make vendor`, run `make workflow-test`, run `make hook-test`, and run `make build-ci`. The `lint` job SHALL run on `ubuntu-latest`, set up Go from `go.mod`, set up Terraform without wrapper mode, install Node.js (24.x), run `npm ci`, run `openspec validate --specs` with telemetry disabled, and run `make check-lint`.

#### Scenario: Build job runs workflow and hook tests

- GIVEN the build job runs after Go and Node setup complete
- WHEN the pre-build verification steps execute
- THEN `make workflow-test` SHALL run before `make build-ci`
- AND `make hook-test` SHALL run before `make build-ci`

#### Scenario: Lint validates OpenSpec

- GIVEN the lint job runs after dependencies are installed
- WHEN OpenSpec specs are present under `openspec/specs/`
- THEN `openspec validate --specs` SHALL run successfully before Go/terraform lint checks
