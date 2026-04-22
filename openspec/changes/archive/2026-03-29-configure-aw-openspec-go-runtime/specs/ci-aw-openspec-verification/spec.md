## MODIFIED Requirements

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the same core toolchain layers as the `lint` job before agent verification begins. At a minimum, the review environment SHALL set up Node using the version range declared in `package.json` engines, SHALL configure Go in the runner environment through an explicit `actions/setup-go` step that reads `go.mod`, and SHALL also declare the Go runtime for the agent environment through workflow frontmatter `runtimes.go.version`. The configured frontmatter Go version SHALL match the repository version declared in `go.mod`. The workflow SHALL also make Terraform CLI available with wrapper behavior disabled so agent-executed commands do not depend on runner-default toolchains.

#### Scenario: Node toolchain follows the repository declaration
- **GIVEN** the repository declares a Node version range in `package.json` engines
- **WHEN** the `verify-openspec` review environment is prepared
- **THEN** the Node runtime available to the agent SHALL satisfy that declared engine range

#### Scenario: Runner Go toolchain follows the repository declaration
- **GIVEN** the workflow prepares the runner environment for repository setup steps
- **WHEN** the explicit `actions/setup-go` step runs
- **THEN** it SHALL read the Go version from `go.mod` so dependency installation and bootstrap commands use the repository-declared Go toolchain

#### Scenario: Agent Go runtime is declared in frontmatter
- **GIVEN** the workflow source defines runtime provisioning for the agent environment
- **WHEN** maintainers configure the workflow frontmatter
- **THEN** the Go toolchain for the agent workspace SHALL be requested through `runtimes.go.version`

#### Scenario: Go runtime stays aligned with the repository declaration
- **GIVEN** the repository declares a Go version in `go.mod`
- **WHEN** the `verify-openspec` review environment is prepared
- **THEN** the Go toolchain available to the agent SHALL match the version maintained in `go.mod` even though the agent runtime is configured explicitly in frontmatter

#### Scenario: Terraform CLI matches repository CI expectations
- **GIVEN** the review workflow uses repository scripts or commands that require Terraform CLI behavior consistent with CI
- **WHEN** the review environment is prepared
- **THEN** Terraform SHALL be available in that environment without wrapper behavior enabled
