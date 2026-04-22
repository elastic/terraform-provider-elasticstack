## ADDED Requirements

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the same core toolchain layers as the `lint` job before agent verification begins. At a minimum, the review environment SHALL set up Node using the version range declared in `package.json` engines, Go using the version declared in `go.mod`, and Terraform CLI with wrapper behavior disabled so agent-executed commands do not depend on runner-default toolchains.

#### Scenario: Node toolchain follows the repository declaration
- **GIVEN** the repository declares a Node version range in `package.json` engines
- **WHEN** the `verify-openspec` review environment is prepared
- **THEN** the Node runtime available to the agent SHALL satisfy that declared engine range

#### Scenario: Go toolchain follows the repository declaration
- **GIVEN** the repository declares a Go version in `go.mod`
- **WHEN** the `verify-openspec` review environment is prepared
- **THEN** the Go toolchain available to the agent SHALL satisfy the version declared in `go.mod`

#### Scenario: Terraform CLI matches repository CI expectations
- **GIVEN** the review workflow uses repository scripts or commands that require Terraform CLI behavior consistent with CI
- **WHEN** the review environment is prepared
- **THEN** Terraform SHALL be available in that environment without wrapper behavior enabled

### Requirement: Review environment installs repository dependencies before verification
Before the agent performs verification, the workflow SHALL run `make setup` in the agent workspace after runtime provisioning completes. This bootstrap SHALL make `npx openspec` available locally and SHALL prepare repository Go dependencies needed by agent-invoked Go commands through the repository's standard setup path.

#### Scenario: Review workspace runs repository setup
- **GIVEN** a qualifying `verify-openspec` run reaches the review job after Node, Go, and Terraform have been provisioned
- **WHEN** the workflow prepares the repository for agent verification
- **THEN** it SHALL run `make setup` in the review workspace before agent reasoning begins

#### Scenario: OpenSpec CLI is ready in the agent workspace
- **GIVEN** a qualifying `verify-openspec` run reaches the review job
- **WHEN** `make setup` completes
- **THEN** the agent SHALL be able to run `npx openspec status --change "<id>"` without first performing ad hoc dependency installation in the prompt

#### Scenario: Agent-invoked Go commands use prepared dependencies
- **GIVEN** verification work invokes `go test` or another repository Go command
- **WHEN** `make setup` has completed in the review workspace
- **THEN** that command SHALL run against the provisioned Go toolchain and repository dependencies instead of failing solely because the base runner lacked the required Go version or module setup
