## MODIFIED Requirements

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the same core toolchain layers as the `lint` job before agent verification begins. At a minimum, the review environment SHALL set up Node using `actions/setup-node` with `node-version-file: package.json`, SHALL configure Go in the runner environment through `actions/setup-go` with `go-version-file: go.mod`, SHALL export `GOROOT`, `GOPATH`, and `GOMODCACHE` after Go setup for AWF chroot mode, SHALL allow the Go ecosystem in the workflow's AWF network policy, and SHALL NOT use workflow frontmatter `runtimes.go` for Go provisioning. The workflow SHALL also make Terraform CLI available with wrapper behavior disabled so agent-executed commands do not depend on runner-default toolchains.

#### Scenario: Node toolchain follows package.json
- **GIVEN** the repository declares the supported Node version in `package.json`
- **WHEN** the `verify-openspec` review environment is prepared
- **THEN** the workflow SHALL configure `actions/setup-node` with `node-version-file: package.json`

#### Scenario: Go toolchain follows go.mod
- **GIVEN** the workflow prepares the runner environment for repository setup steps
- **WHEN** the Go toolchain is installed
- **THEN** the workflow SHALL configure `actions/setup-go` with `go-version-file: go.mod`

#### Scenario: AWF chroot mode receives the configured Go paths
- **GIVEN** the review workflow has installed Go from `go.mod`
- **WHEN** the agent environment is prepared for AWF chroot mode
- **THEN** the workflow SHALL export `GOROOT=$(go env GOROOT)` to `GITHUB_ENV`
- **AND** the workflow SHALL export `GOPATH=$(go env GOPATH)` to `GITHUB_ENV`
- **AND** the workflow SHALL export `GOMODCACHE=$(go env GOMODCACHE)` to `GITHUB_ENV`

#### Scenario: AWF network policy allows the Go ecosystem
- **GIVEN** agent-executed verification commands may need Go module network access
- **WHEN** maintainers inspect the workflow frontmatter
- **THEN** `network.allowed` SHALL include `go`

#### Scenario: Review bootstrap does not use runtimes.go
- **GIVEN** the review workflow bootstrap is implemented
- **WHEN** maintainers inspect the authored workflow source
- **THEN** it SHALL provision Go from `go.mod` and SHALL NOT declare `runtimes.go`

#### Scenario: Terraform CLI matches repository CI expectations
- **GIVEN** the review workflow uses repository scripts or commands that require Terraform CLI behavior consistent with CI
- **WHEN** the review environment is prepared
- **THEN** Terraform SHALL be available in that environment without wrapper behavior enabled

### Requirement: Review environment installs repository dependencies before verification
Before the agent performs verification, the workflow SHALL run `make setup` in the agent workspace after runtime provisioning completes. This bootstrap SHALL make `npx openspec` available locally, SHALL prepare repository Go dependencies needed by agent-invoked Go commands through the repository's standard setup path, and SHALL preserve access to the prepared Go workspace and module cache for AWF agent commands during verification.

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
- **THEN** that command SHALL run against the provisioned Go toolchain and prepared Go dependencies instead of failing solely because the base runner lacked the required Go version or module setup

#### Scenario: Prepared module cache remains available in AWF
- **GIVEN** the workflow prepared Go dependencies before agent reasoning
- **WHEN** an AWF agent command runs Go module-aware verification in chroot mode
- **THEN** the command SHALL retain access to the configured Go workspace and module cache through the exported Go environment variables
