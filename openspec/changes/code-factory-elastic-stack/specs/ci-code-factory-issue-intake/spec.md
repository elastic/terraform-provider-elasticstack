## MODIFIED Requirements

### Requirement: Workflow activates the implementation agent only for qualifying `code-factory` issue events
The workflow MAY subscribe to GitHub `issues.opened` and `issues.labeled` events. For issue-event intake, eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `code-factory`, and `issues.opened` when the issue already includes the `code-factory` label at creation time.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `code-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the implementation agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `code-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the implementation agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `code-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the implementation agent for that event

### Requirement: Implementation agent has structured access to Elastic documentation
The `code-factory` workflow SHALL configure the Elastic docs MCP server as an HTTP MCP server in the workflow frontmatter so that the implementation agent can query Elastic documentation during issue investigation and implementation. The workflow frontmatter SHALL declare an `mcp-servers.elastic-docs` entry pointing to `https://www.elastic.co/docs/_mcp/`. The agent prompt SHALL instruct the agent to use the docs MCP tools (`search_docs`, `find_related_docs`, `get_document_by_url`) when investigating the API behavior, parameters, or constraints required to implement a `code-factory` issue.

#### Scenario: Agent investigates API behavior before implementing a resource
- **WHEN** a `code-factory` issue involves an Elastic API endpoint or feature whose full parameter set is not evident from the existing codebase
- **THEN** the agent SHALL use the elastic-docs MCP `search_docs` tool to retrieve authoritative API documentation before writing implementation code

#### Scenario: Elastic docs MCP server is unavailable
- **WHEN** the elastic-docs MCP tools return an error or are unreachable during a `code-factory` run
- **THEN** the agent SHALL proceed with implementation using the information available in the issue and the repository codebase
- **AND** it SHALL NOT block the run solely because the docs MCP is unavailable

#### Scenario: Maintainer inspects compiled workflow for docs MCP configuration
- **WHEN** maintainers inspect the compiled `code-factory-issue.md` workflow
- **THEN** the workflow frontmatter SHALL include `mcp-servers.elastic-docs` with `url: https://www.elastic.co/docs/_mcp/`

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
