## MODIFIED Requirements

### Requirement: Review environment bootstraps repository toolchains
The workflow SHALL provision the review environment using explicit workflow frontmatter runtime declarations only. At a minimum, the workflow SHALL set `runtimes.go.version` to `1.26.1` and SHALL set `runtimes.node.version` to `24`. The workflow SHALL NOT require `actions/setup-go` as part of this review-environment bootstrap. Repository make-based runtime validation SHALL confirm that the configured Go runtime matches the version declared in `go.mod` and that the configured Node runtime satisfies the `package.json` `engines.node` range. The workflow SHALL also make Terraform CLI available with wrapper behavior disabled so agent-executed commands do not depend on runner-default toolchains.

#### Scenario: Workflow pins the review Go runtime
- **GIVEN** the workflow source defines runtime provisioning for the review environment
- **WHEN** maintainers configure the workflow frontmatter
- **THEN** it SHALL declare `runtimes.go.version` as `1.26.1`

#### Scenario: Workflow pins the review Node runtime
- **GIVEN** the workflow source defines runtime provisioning for the review environment
- **WHEN** maintainers configure the workflow frontmatter
- **THEN** it SHALL declare `runtimes.node.version` as `24`

#### Scenario: Pinned Node runtime remains compatible with package engines
- **GIVEN** the repository declares a Node engine range in `package.json`
- **WHEN** repository runtime validation runs
- **THEN** it SHALL verify that workflow `runtimes.node.version` satisfies that engine range

#### Scenario: Pinned Go runtime remains aligned with go.mod
- **GIVEN** the repository declares a Go version in `go.mod`
- **WHEN** repository runtime validation runs
- **THEN** it SHALL verify that workflow `runtimes.go.version` matches the version declared in `go.mod`

#### Scenario: Review bootstrap does not use actions setup-go
- **GIVEN** the review workflow bootstrap is implemented
- **WHEN** maintainers inspect the workflow steps required before agent verification
- **THEN** those requirements SHALL rely on the frontmatter runtime declarations and SHALL NOT require an `actions/setup-go` step

#### Scenario: Terraform CLI matches repository CI expectations
- **GIVEN** the review workflow uses repository scripts or commands that require Terraform CLI behavior consistent with CI
- **WHEN** the review environment is prepared
- **THEN** Terraform SHALL be available in that environment without wrapper behavior enabled
