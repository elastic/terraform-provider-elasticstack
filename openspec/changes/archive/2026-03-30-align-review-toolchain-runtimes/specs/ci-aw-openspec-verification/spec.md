## MODIFIED Requirements

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the same core toolchain layers as the `lint` job before agent verification begins. At a minimum, the review environment SHALL set up Node using `actions/setup-node` with `node-version-file: package.json`, SHALL configure Go in the runner environment through `actions/setup-go` with `go-version-file: go.mod`, SHALL export `GOROOT` after Go setup for AWF chroot mode, and SHALL NOT use workflow frontmatter `runtimes.go` for Go provisioning. The workflow SHALL also make Terraform CLI available with wrapper behavior disabled so agent-executed commands do not depend on runner-default toolchains.

#### Scenario: Node toolchain follows package.json
- **GIVEN** the repository declares the supported Node version in `package.json`
- **WHEN** the `verify-openspec` review environment is prepared
- **THEN** the workflow SHALL configure `actions/setup-node` with `node-version-file: package.json`

#### Scenario: Go toolchain follows go.mod
- **GIVEN** the workflow prepares the runner environment for repository setup steps
- **WHEN** the Go toolchain is installed
- **THEN** the workflow SHALL configure `actions/setup-go` with `go-version-file: go.mod`

#### Scenario: AWF chroot mode receives the configured GOROOT
- **GIVEN** the review workflow has installed Go from `go.mod`
- **WHEN** the agent environment is prepared for AWF chroot mode
- **THEN** the workflow SHALL export `GOROOT=$(go env GOROOT)` to `GITHUB_ENV`

#### Scenario: Review bootstrap does not use runtimes.go
- **GIVEN** the review workflow bootstrap is implemented
- **WHEN** maintainers inspect the authored workflow source
- **THEN** it SHALL provision Go from `go.mod` and SHALL NOT declare `runtimes.go`

#### Scenario: Terraform CLI matches repository CI expectations
- **GIVEN** the review workflow uses repository scripts or commands that require Terraform CLI behavior consistent with CI
- **WHEN** the review environment is prepared
- **THEN** Terraform SHALL be available in that environment without wrapper behavior enabled
