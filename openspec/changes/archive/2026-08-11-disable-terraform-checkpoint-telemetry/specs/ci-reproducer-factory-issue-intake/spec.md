## MODIFIED Requirements

### Requirement: Workflow provides an Elastic Stack acceptance-test environment

The reproduction agent SHALL have access to a live Elastic Stack for running acceptance tests, using the same environment configuration as the `code-factory` workflow. The agent prompt SHALL document the connection parameters: `ELASTICSEARCH_ENDPOINTS=http://host.docker.internal:9201`, `ELASTICSEARCH_USERNAME=elastic`, `ELASTICSEARCH_PASSWORD=password`, `KIBANA_ENDPOINT=http://host.docker.internal:5602`, and `CHECKPOINT_DISABLE=1`. The network allow-list SHALL include `go` to support `go test` downloads. `CHECKPOINT_DISABLE` SHALL be set to `1` so the embedded `terraform` binary used by `terraform-plugin-testing` does not attempt an outbound call to `checkpointapi.hashicorp.com` that the AWF sandbox's egress firewall would otherwise block.

#### Scenario: Agent runs an acceptance test
- **WHEN** the reproduction agent writes a `TestAccReproduceIssue{N}` test and runs it
- **THEN** the test SHALL be able to reach the Elastic Stack at the documented endpoints

#### Scenario: Maintainer inspects workflow network policy
- **WHEN** maintainers inspect the authored workflow frontmatter
- **THEN** `network.allowed` SHALL include `go`

#### Scenario: Agent's acceptance-test invocation disables Terraform checkpoint telemetry
- **WHEN** the reproduction agent runs `TF_ACC=1 go test ...` against the provisioned Elastic Stack
- **THEN** `CHECKPOINT_DISABLE` SHALL be set to `1` in the agent's environment for that invocation
- **AND** the AWF sandbox's egress firewall SHALL NOT report `checkpointapi.hashicorp.com` as a blocked domain for that run
