## ADDED Requirements

### Requirement: No standalone generated SLO client in the module graph

The provider’s Go packages SHALL NOT import `github.com/elastic/terraform-provider-elasticstack/generated/slo`. The repository SHALL NOT ship a `generated/slo` Go module tree as part of the build after this change.

#### Scenario: Source tree free of generated/slo imports

- **WHEN** static analysis or `go test`/`go build` runs across `./...` at the repository root
- **THEN** no package SHALL import the `generated/slo` module path

### Requirement: No go-kibana-rest module dependency

The root `go.mod` SHALL NOT contain a `require` directive for `github.com/disaster37/go-kibana-rest/v8` and SHALL NOT contain a `replace` directive mapping that module path to `./libs/go-kibana-rest` (or any other local path). After removal, `go mod tidy` SHALL leave no requirement edge to that module for the provider.

#### Scenario: Module file excludes legacy client

- **WHEN** a reviewer opens the root `go.mod` after this change
- **THEN** neither `github.com/disaster37/go-kibana-rest/v8` nor a `replace` stanza for it SHALL appear

### Requirement: Vendored go-kibana-rest fork removed

The directory `libs/go-kibana-rest` SHALL NOT remain in the repository when its sole purpose was to satisfy the removed `replace` directive for `github.com/disaster37/go-kibana-rest/v8`.

#### Scenario: Fork directory absent

- **WHEN** the change is applied and the branch builds successfully
- **THEN** the path `libs/go-kibana-rest` SHALL not exist in the tree (unless a separate, explicitly scoped follow-up documents a different use — out of scope for this change; default is deletion)

### Requirement: No legacy Kibana REST import paths in provider code

No first-party Go source under the provider module (for example `internal/`, `generated/kbapi` consumers, tools owned by this repo) SHALL import `github.com/disaster37/go-kibana-rest/v8` or subpackages such as `github.com/disaster37/go-kibana-rest/v8/kbapi`.

#### Scenario: Repository search shows no deprecated imports

- **WHEN** a maintainer searches the repository for `disaster37/go-kibana-rest` and `terraform-provider-elasticstack/generated/slo`
- **THEN** no matches SHALL appear in first-party Go sources, the root `Makefile`, or GitHub Actions workflow definitions under `.github/workflows/`

### Requirement: NOTICE file excludes go-kibana-rest attribution

The `NOTICE` file SHALL NOT contain the `github.com/disaster37/go-kibana-rest/v8` attribution entry after the dependency is removed.

#### Scenario: NOTICE attribution removed

- **WHEN** `github.com/disaster37/go-kibana-rest/v8` is no longer a dependency
- **THEN** its attribution block SHALL be absent from the `NOTICE` file
