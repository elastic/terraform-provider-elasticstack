# ci-schema-coverage-rotation-toolchain Specification

## Purpose

Specifies how the `schema-coverage-rotation` workflow configures the GitHub Agentic Workflows **engine** (Claude, Anthropic-compatible LiteLLM routing, secret-backed API key, explicit per-tool timeout), prepares the repository (Go, Node, `make setup`) before the agent runs repository-local commands, and which **authored** `network.allowed` entries the workflow declares—including `elastic.litellm-prod.ai` for the Claude engine’s LiteLLM proxy route. (The compiled GH AW lock may still expand AWF egress with additional compiler-managed domains.)

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
The `schema-coverage-rotation` workflow SHALL declare an AWF network policy that allows the repository bootstrap path to use the default allowlist plus the Node and Go ecosystems, and SHALL allow `elastic.litellm-prod.ai` for the Claude engine's Anthropic-compatible proxy access.

#### Scenario: Workflow frontmatter allows required ecosystems
- **WHEN** maintainers inspect the schema-coverage rotation workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`
- **AND** `network.allowed` SHALL include `node`
- **AND** `network.allowed` SHALL include `go`
- **AND** `network.allowed` SHALL include `elastic.litellm-prod.ai`

### Requirement: Schema-coverage rotation uses the Claude engine through the Anthropic-compatible proxy
The `schema-coverage-rotation` workflow SHALL set `engine.id` to `claude` and SHALL route Claude traffic through `ANTHROPIC_BASE_URL` set to `https://elastic.litellm-prod.ai/`. Any configured `ANTHROPIC_API_KEY` value SHALL be sourced from a GitHub Actions secret-backed expression rather than from a checked-in literal.

#### Scenario: Authored workflow selects Claude
- **WHEN** maintainers inspect the authored `schema-coverage-rotation` workflow source
- **THEN** `engine.id` SHALL be `claude`

#### Scenario: Claude traffic is routed through the Elastic LiteLLM endpoint
- **WHEN** maintainers inspect the workflow engine environment
- **THEN** `ANTHROPIC_BASE_URL` SHALL be `https://elastic.litellm-prod.ai/`

#### Scenario: Anthropic authentication is secret-backed
- **WHEN** maintainers inspect the authored workflow source
- **THEN** any configured `ANTHROPIC_API_KEY` value SHALL come from a GitHub Actions secret expression rather than a literal API key value committed to the repository

### Requirement: Schema-coverage rotation sets an explicit Claude execution budget
The `schema-coverage-rotation` workflow SHALL set an explicit per-tool execution timeout of 300 seconds so the repository-local bootstrap and schema-analysis commands are not constrained by the Claude engine's shorter default tool-call timeout.

#### Scenario: Workflow frontmatter defines the Claude tool timeout
- **WHEN** maintainers inspect the schema-coverage rotation workflow frontmatter
- **THEN** `tools.timeout` SHALL be `300`

### Requirement: Agent instructions rely on deterministic bootstrap
The `schema-coverage-rotation` workflow prompt SHALL rely on deterministic workflow setup for repository toolchain provisioning and SHALL NOT require the agent to install or discover the required Go or Node toolchains itself before running repository-local commands.

#### Scenario: Prompt begins after bootstrap is complete
- **WHEN** the agent receives the schema-coverage rotation prompt
- **THEN** the workflow SHALL already have completed the repository toolchain setup needed for `go run ./scripts/schema-coverage-rotation ...`

