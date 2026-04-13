## ADDED Requirements

### Requirement: Verification engine uses Copilot BYOK LiteLLM routing
The `openspec-verify-label` workflow SHALL keep `engine.id: copilot`, SHALL set `engine.model` to `llm-gateway/gpt-5.4`, and SHALL configure Copilot CLI to use an OpenAI-compatible BYOK provider for model inference by setting `COPILOT_PROVIDER_TYPE` to `openai` and `COPILOT_PROVIDER_BASE_URL` to `https://elastic.litellm-prod.ai/v1`. Any provider credential passed as `COPILOT_PROVIDER_API_KEY` SHALL be sourced from a GitHub Actions secret-backed expression rather than from a checked-in literal.

#### Scenario: Authored workflow preserves the Copilot engine
- **WHEN** maintainers inspect the authored `openspec-verify-label` workflow source
- **THEN** `engine.id` SHALL be `copilot`
- **AND** `engine.model` SHALL be `llm-gateway/gpt-5.4`

#### Scenario: Copilot BYOK provider targets the Elastic LiteLLM endpoint
- **WHEN** maintainers inspect the verification workflow's engine environment
- **THEN** `COPILOT_PROVIDER_TYPE` SHALL be `openai`
- **AND** `COPILOT_PROVIDER_BASE_URL` SHALL be `https://elastic.litellm-prod.ai/v1`

#### Scenario: Provider authentication is secret-backed
- **WHEN** maintainers inspect the authored workflow source
- **THEN** any configured `COPILOT_PROVIDER_API_KEY` value SHALL come from a GitHub Actions secret expression rather than a literal API key value committed to the repository

## MODIFIED Requirements

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the same core toolchain layers as the `lint` job before agent verification begins. At a minimum, it SHALL set up Node using `actions/setup-node` with `node-version-file: package.json`, SHALL configure Go in the runner environment through `actions/setup-go` with `go-version-file: go.mod`, SHALL export `GOROOT`, `GOPATH`, and `GOMODCACHE` after Go setup for AWF chroot mode, SHALL allow the Go ecosystem and `elastic.litellm-prod.ai` in the workflow's AWF network policy, and SHALL NOT use workflow frontmatter `runtimes.go` for Go provisioning.

#### Scenario: Node toolchain follows package.json
- **GIVEN** the repository declares the supported Node version in `package.json`
- **WHEN** the `verify-openspec` review environment is prepared in workspace mode
- **THEN** the workflow SHALL configure `actions/setup-node` with `node-version-file: package.json`

#### Scenario: Go toolchain follows go.mod
- **GIVEN** the workflow prepares the runner environment for repository setup steps in workspace mode
- **WHEN** the Go toolchain is installed
- **THEN** the workflow SHALL configure `actions/setup-go` with `go-version-file: go.mod`

#### Scenario: AWF chroot mode receives the configured Go paths
- **GIVEN** the review workflow has installed Go from `go.mod` in workspace mode
- **WHEN** the agent environment is prepared for AWF chroot mode
- **THEN** the workflow SHALL export `GOROOT=$(go env GOROOT)` to `GITHUB_ENV`
- **AND** the workflow SHALL export `GOPATH=$(go env GOPATH)` to `GITHUB_ENV`
- **AND** the workflow SHALL export `GOMODCACHE=$(go env GOMODCACHE)` to `GITHUB_ENV`

#### Scenario: AWF network policy allows the Go ecosystem and LiteLLM host
- **GIVEN** agent-executed verification commands may need Go module network access and LiteLLM provider access
- **WHEN** maintainers inspect the workflow frontmatter
- **THEN** `network.allowed` SHALL include `go`
- **AND** `network.allowed` SHALL include `elastic.litellm-prod.ai`

#### Scenario: Review bootstrap does not use runtimes.go
- **GIVEN** the review workflow bootstrap is implemented
- **WHEN** maintainers inspect the authored workflow source
- **THEN** it SHALL provision Go from `go.mod` and SHALL NOT declare `runtimes.go`

#### Scenario: Terraform CLI matches repository CI expectations
- **GIVEN** the review workflow uses repository scripts or commands that require Terraform CLI behavior consistent with CI
- **WHEN** the review environment is prepared in workspace mode
- **THEN** Terraform SHALL be available in that environment without wrapper behavior enabled
