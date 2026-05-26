## ADDED Requirements

### Requirement: Shared setup installs Go
The shared workflow component SHALL install Go using `actions/setup-go@v6` with `go-version-file: go.mod` and `cache: false`.

#### Scenario: Go is installed before agent execution
- **WHEN** a workflow imports the shared setup component
- **THEN** the Go toolchain is available for the agent to run with `go run`, `go test`, `go build`, and `gopls`

### Requirement: Shared setup installs Terraform CLI
The shared workflow component SHALL install Terraform CLI using `hashicorp/setup-terraform@v4` with `terraform_wrapper: false`.

#### Scenario: Terraform CLI is installed before agent execution
- **WHEN** a workflow imports the shared setup component
- **THEN** the `terraform` binary is available in PATH for the agent to invoke

### Requirement: Shared setup exports chroot paths
The shared workflow component SHALL export `GOROOT`, `GOPATH`, `GOMODCACHE`, `TERRAFORM_BIN`, and prepend the Terraform binary directory to `PATH` in `$GITHUB_ENV` so the agent's chroot sandbox can access these tools.

#### Scenario: Chroot sandbox has Go and Terraform access
- **WHEN** the shared setup component runs within an agentic workflow
- **THEN** the agent process inside the chroot can execute `go` and `terraform` commands

### Requirement: Shared setup installs Node.js
The shared workflow component SHALL install Node.js using `actions/setup-node@v6` with `node-version-file: package.json`.

#### Scenario: Node.js is installed before agent execution
- **WHEN** a workflow imports the shared setup component
- **THEN** `node` and `npm` are available for pre-commit hooks, OpenSpec CLI, and other Node-based tooling

### Requirement: Shared setup installs repository dependencies
The shared workflow component SHALL run `make setup` to install all repository-level dependencies including Go modules, OpenSpec CLI, and any other dev tooling defined by the Makefile.

#### Scenario: Repository dependencies are ready before agent execution
- **WHEN** the shared setup component completes
- **THEN** the agent can run `make build`, `make check-lint`, `make check-openspec`, and other Makefile targets without additional setup

### Requirement: Shared setup is unconditional
The shared workflow component SHALL execute all its steps unconditionally. There SHALL be no `import-schema` parameters or `if:` conditions on any step.

#### Scenario: Importing workflow requires no configuration
- **WHEN** a workflow template adds `imports: [shared/setup-dev.yml]` to its frontmatter
- **THEN** the component runs all setup steps without the importing workflow passing any parameters
