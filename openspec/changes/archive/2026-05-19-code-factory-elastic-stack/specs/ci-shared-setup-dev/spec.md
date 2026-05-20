## ADDED Requirements

### Requirement: Terraform CLI SHALL be discoverable inside the agentic sandbox
The `shared/setup-dev.md` workflow component SHALL stage the Terraform binary into the tracked workspace so that agentic sandboxes can discover and execute it, because the AWF container does not mount the GitHub Actions toolcache where `hashicorp/setup-terraform` installs the binary.

#### Scenario: Agent discovers Terraform during verification
- **WHEN** the implementation agent runs `terraform` as part of acceptance tests or provider validation
- **THEN** the binary SHALL be discoverable within the agentic sandbox's PATH at `$GITHUB_WORKSPACE/bin/terraform`

#### Scenario: Terraform is staged before agent activation
- **WHEN** the workflow runs the `Setup Terraform CLI` step
- **THEN** a subsequent step SHALL copy the Terraform binary into `$GITHUB_WORKSPACE/bin/terraform`
- **AND** it SHALL prepend `$GITHUB_WORKSPACE/.bin` to `PATH`

### Requirement: Go toolchain SHALL be discoverable inside the agentic sandbox
The `shared/setup-dev.md` workflow component SHALL export `GOROOT`, `GOPATH`, and `GOMODCACHE` into `GITHUB_ENV` after running `actions/setup-go` so the AWF chroot container can discover the Go toolchain.

#### Scenario: Agent runs Go commands during verification
- **WHEN** the implementation agent runs `go test` or `go build`
- **THEN** the Go binary and module cache SHALL be available within the agentic sandbox

### Requirement: Node.js SHALL be available inside the agentic sandbox
The `shared/setup-dev.md` workflow component SHALL install Node.js via `actions/setup-node` and make it available to the agentic sandbox for any repository scripts that require it.

#### Scenario: Agent runs Node-based scripts
- **WHEN** the implementation agent runs a Node.js script (e.g. `make setup` dependencies)
- **THEN** Node.js SHALL be available in the agentic sandbox

### Requirement: Repository dependencies SHALL be installed before agent activation
The `shared/setup-dev.md` workflow component SHALL run `make setup` after installing Go, Terraform, and Node.js so the agentic sandbox has all required dependencies.

#### Scenario: Agent builds or tests the provider
- **WHEN** the implementation agent runs `make build` or `go test ./...`
- **THEN** all repository dependencies SHALL already be installed
