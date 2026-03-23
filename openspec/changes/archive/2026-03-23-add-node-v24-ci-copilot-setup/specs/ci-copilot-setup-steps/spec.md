## MODIFIED Requirements

### Requirement: Toolchain and checkout (REQ-006–REQ-008)

The job SHALL check out the repository using `actions/checkout` pinned by commit SHA. The job SHALL install Node.js using `actions/setup-node` pinned by commit SHA with **`node-version-file` set to the repository root `package.json`** and **without** a conflicting `node-version` input that would override the file (per the action’s documented behavior). The resolved version SHALL follow the action’s documented precedence among `package.json` fields (`volta.node`, then `devEngines.runtime` for node, then `engines.node`). The job SHALL enable npm caching and use `package-lock.json` as the cache dependency path. The job SHALL install Go using `actions/setup-go` with `go-version-file: go.mod` and Go module caching enabled. The job SHALL install Terraform using `hashicorp/setup-terraform` with `terraform_wrapper: false`.

#### Scenario: Go version tracks the module

- GIVEN `go.mod` specifies the toolchain
- WHEN setup-go runs
- THEN the Go version SHALL be derived from `go.version-file` / `go.mod` per the action configuration

#### Scenario: Node satisfies the version read from package.json

- GIVEN the repository root `package.json` declares a Node version requirement via the fields `actions/setup-node` reads for `node-version-file` (in precedence order)
- WHEN setup-node runs
- THEN the job SHALL provision a Node.js version that satisfies the semver range (or exact version) resolved from that file so `node` and `npm` meet the repository’s declared requirement for OpenSpec and npm-based Makefile targets

#### Scenario: npm dependencies cached for setup

- GIVEN `package-lock.json` is present in the repository
- WHEN setup-node runs with npm caching configured
- THEN the action SHALL use `package-lock.json` as the cache dependency path for npm
