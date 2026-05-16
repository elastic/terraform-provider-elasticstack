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

The agent prompt SHALL also describe the test environment with ports that are reachable from within the AWF sandbox, including `ELASTICSEARCH_ENDPOINTS` on an AWF-allowed port (such as `8080`) and `KIBANA_ENDPOINT` on an AWF-allowed port (such as `80`), and SHALL instruct the agent to run acceptance tests using those reachable endpoints.

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

#### Scenario: Agent runs acceptance tests against reachable stack endpoints
- **WHEN** the implementation agent reaches the acceptance test verification step
- **THEN** the agent SHALL connect to Elasticsearch at `http://host.docker.internal:8080` and Kibana at `http://host.docker.internal` for test execution
