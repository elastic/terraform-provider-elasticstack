## MODIFIED Requirements

### Requirement: Implementation agent can run acceptance tests against the Elastic Stack

The `code-factory` workflow SHALL import `shared/elastic-stack.md` so that the Elastic Stack (Elasticsearch and Kibana) is provisioned and reachable from within the AWF agentic sandbox. The agent prompt SHALL describe the test environment using the proxy ports (`9201` for Elasticsearch, `5602` for Kibana) accessed via `host.docker.internal`, and SHALL instruct the agent that acceptance tests are runnable. The workflow SHALL disable the Terraform CLI's checkpoint telemetry call (`CHECKPOINT_DISABLE=1`) for acceptance-test invocations, so the embedded `terraform` binary used by `terraform-plugin-testing` does not attempt an outbound call to `checkpointapi.hashicorp.com` that the AWF sandbox's egress firewall would otherwise block.

#### Scenario: Agent runs acceptance tests against the live stack
- **WHEN** the implementation agent reaches the acceptance test verification step
- **THEN** the agent SHALL connect to Elasticsearch at `http://host.docker.internal:9201` and Kibana at `http://host.docker.internal:5602` for test execution
- **AND** `TF_ACC=1` acceptance tests SHALL be expected to pass when correctly implemented

#### Scenario: Agent prompt reflects a reachable test environment
- **WHEN** the implementation agent reads the test environment instructions in the agent prompt
- **THEN** the prompt SHALL describe the Elastic Stack as provisioned and reachable via `host.docker.internal` on its proxy ports (`9201` for Elasticsearch, `5602` for Kibana)
- **AND** the prompt SHALL NOT state that acceptance tests are blocked by a network policy issue

#### Scenario: Agent's acceptance-test invocation disables Terraform checkpoint telemetry
- **WHEN** the implementation agent runs `TF_ACC=1 go test ...` against the provisioned Elastic Stack
- **THEN** `CHECKPOINT_DISABLE` SHALL be set to `1` in the agent's environment for that invocation
- **AND** the AWF sandbox's egress firewall SHALL NOT report `checkpointapi.hashicorp.com` as a blocked domain for that run
