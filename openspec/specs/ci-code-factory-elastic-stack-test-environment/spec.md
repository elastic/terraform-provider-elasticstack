# ci-code-factory-elastic-stack-test-environment Specification

## Purpose
TBD - created by archiving change code-factory-elastic-stack. Update Purpose after archive.
## Requirements
### Requirement: code-factory workflow SHALL import shared Elastic Stack and dev-setup components
The `code-factory-issue` workflow SHALL import `shared/elastic-stack.md` so the Elastic Stack is provisioned and reachable, and SHALL import `shared/setup-dev.md` so the agent can discover Go, Terraform, and Node.js tooling. The agent prompt SHALL describe the test environment using proxy ports (`9201` for Elasticsearch, `5602` for Kibana) and SHALL instruct the agent that acceptance tests are runnable.

#### Scenario: Agent runs acceptance tests against the live stack
- **WHEN** the implementation agent reaches the acceptance test verification step
- **THEN** the agent SHALL connect to Elasticsearch at `http://host.docker.internal:9201` and Kibana at `http://host.docker.internal:5602` for test execution
- **AND** `TF_ACC=1` acceptance tests SHALL be expected to pass when correctly implemented

#### Scenario: Agent prompt reflects a reachable test environment
- **WHEN** the implementation agent reads the test environment instructions in the agent prompt
- **THEN** the prompt SHALL describe the Elastic Stack as provisioned and reachable via `host.docker.internal` on its proxy ports
- **AND** the prompt SHALL NOT state that acceptance tests are blocked by a network policy issue

### Requirement: code-factory agent prompt documents correct test endpoints
The `code-factory-issue.md` agent prompt SHALL provide example `go test` commands that use the proxy ports in `ELASTICSEARCH_ENDPOINTS` and `KIBANA_ENDPOINT`, and SHALL list acceptance test execution as a required verification task.

#### Scenario: Maintainer inspects agent prompt for test instructions
- **WHEN** maintainers inspect the compiled `code-factory-issue.md` workflow
- **THEN** the `## Test environment` section SHALL contain `http://host.docker.internal:9201` for Elasticsearch
- **AND** it SHALL contain `http://host.docker.internal:5602` for Kibana
- **AND** the `## Verification tasks` section SHALL instruct the agent to run acceptance tests

