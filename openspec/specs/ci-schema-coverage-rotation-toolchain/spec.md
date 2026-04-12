# ci-schema-coverage-rotation-toolchain Specification

## Purpose
TBD - created by archiving change schema-coverage-rotation-toolchain-bootstrap. Update Purpose after archive.
## Requirements
### Requirement: Schema-coverage rotation bootstraps repository toolchains before agent execution
The `schema-coverage-rotation` workflow SHALL provision the repository toolchains before agent reasoning begins. At a minimum, it SHALL set up Go using `actions/setup-go` with `go-version-file: go.mod`, SHALL export `GOROOT`, `GOPATH`, and `GOMODCACHE` after Go setup for AWF chroot mode, SHALL set up Node using `actions/setup-node` with `node-version-file: package.json`, and SHALL run `make setup` at the repository root before the agent executes repository-local schema-coverage commands.

#### Scenario: Go toolchain follows go.mod
- **WHEN** the schema-coverage rotation workflow prepares the runner environment for repository-local Go commands
- **THEN** it SHALL configure `actions/setup-go` with `go-version-file: go.mod`

#### Scenario: AWF chroot mode receives configured Go paths
- **WHEN** the workflow has installed Go for schema-coverage rotation
- **THEN** it SHALL export `GOROOT=$(go env GOROOT)` to `GITHUB_ENV`
- **AND** it SHALL export `GOPATH=$(go env GOPATH)` to `GITHUB_ENV`
- **AND** it SHALL export `GOMODCACHE=$(go env GOMODCACHE)` to `GITHUB_ENV`

#### Scenario: Node toolchain follows package.json
- **WHEN** the schema-coverage rotation workflow prepares the repository bootstrap environment
- **THEN** it SHALL configure `actions/setup-node` with `node-version-file: package.json`

#### Scenario: Repository setup completes before agent reasoning
- **WHEN** the workflow finishes provisioning Go and Node for schema-coverage rotation
- **THEN** it SHALL run `make setup` before the agent begins executing the prompt instructions

### Requirement: Schema-coverage rotation allows repository bootstrap ecosystems
The `schema-coverage-rotation` workflow SHALL declare an AWF network policy that allows the repository bootstrap path to use the default allowlist plus the Node and Go ecosystems.

#### Scenario: Workflow frontmatter allows required ecosystems
- **WHEN** maintainers inspect the schema-coverage rotation workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`
- **AND** `network.allowed` SHALL include `node`
- **AND** `network.allowed` SHALL include `go`

### Requirement: Agent instructions rely on deterministic bootstrap
The `schema-coverage-rotation` workflow prompt SHALL rely on deterministic workflow setup for repository toolchain provisioning and SHALL NOT require the agent to install or discover the required Go or Node toolchains itself before running repository-local commands.

#### Scenario: Prompt begins after bootstrap is complete
- **WHEN** the agent receives the schema-coverage rotation prompt
- **THEN** the workflow SHALL already have completed the repository toolchain setup needed for `go run ./scripts/schema-coverage-rotation ...`

