## ADDED Requirements

### Requirement: Implementation agent can run acceptance tests against the Elastic Stack
The `code-factory` workflow SHALL import `shared/elastic-stack.md` so that the Elastic Stack (Elasticsearch and Kibana) is provisioned and reachable from within the AWF agentic sandbox. The agent prompt SHALL describe the test environment using the proxy ports (`9201` for Elasticsearch, `5602` for Kibana) accessed via `host.docker.internal`, and SHALL instruct the agent that acceptance tests are runnable.

#### Scenario: Agent runs acceptance tests against the live stack
- **WHEN** the implementation agent reaches the acceptance test verification step
- **THEN** the agent SHALL connect to Elasticsearch at `http://host.docker.internal:9201` and Kibana at `http://host.docker.internal:5602` for test execution
- **AND** `TF_ACC=1` acceptance tests SHALL be expected to pass when correctly implemented

#### Scenario: Agent prompt reflects a reachable test environment
- **WHEN** the implementation agent reads the test environment instructions in the agent prompt
- **THEN** the prompt SHALL describe the Elastic Stack as provisioned and reachable via `host.docker.internal` on its proxy ports (`9201` for Elasticsearch, `5602` for Kibana)
- **AND** the prompt SHALL NOT state that acceptance tests are blocked by a network policy issue

### Requirement: Implementation agent can discover the Terraform CLI
The `code-factory` workflow SHALL import `shared/setup-dev.md`, which stages the Terraform binary into the workspace so the agentic sandbox can execute it. The agent SHALL be able to run `terraform --version` and `terraform` subcommands during verification.

#### Scenario: Agent discovers Terraform during verification
- **WHEN** the implementation agent runs `terraform` as part of acceptance tests or provider validation
- **THEN** the binary SHALL be discoverable within the agentic sandbox's PATH at a workspace-relative path
