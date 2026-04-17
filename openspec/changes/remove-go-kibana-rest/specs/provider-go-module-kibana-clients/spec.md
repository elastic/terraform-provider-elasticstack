## ADDED Requirements

### Requirement: No go-kibana-rest module dependency

The root `go.mod` SHALL NOT contain a `require` directive for `github.com/disaster37/go-kibana-rest/v8` and SHALL NOT contain a `replace` directive mapping that module path to `./libs/go-kibana-rest` or any other local path. After removal, `go mod tidy` SHALL leave no requirement edge to that module for the provider.

#### Scenario: Module file excludes legacy client
- **WHEN** a reviewer opens the root `go.mod` after this change
- **THEN** neither `github.com/disaster37/go-kibana-rest/v8` nor a `replace` stanza for it SHALL appear

### Requirement: Vendored go-kibana-rest fork removed

The directory `libs/go-kibana-rest` SHALL NOT remain in the repository when its only purpose was to satisfy the removed `replace` directive for `github.com/disaster37/go-kibana-rest/v8`.

#### Scenario: Fork directory absent
- **WHEN** the change is applied and the branch builds successfully
- **THEN** the path `libs/go-kibana-rest` SHALL not exist in the tree

### Requirement: No legacy Kibana REST import paths in first-party code

No first-party Go source under the provider module SHALL import `github.com/disaster37/go-kibana-rest/v8` or subpackages such as `github.com/disaster37/go-kibana-rest/v8/kbapi`.

#### Scenario: Repository search shows no deprecated imports
- **WHEN** a maintainer searches the repository for `disaster37/go-kibana-rest`
- **THEN** no matches SHALL appear in first-party Go sources, the root `go.mod`, or the root `Makefile`
