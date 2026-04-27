## MODIFIED Requirements

### Requirement: Workflow frontmatter allows required agent ecosystems
The `code-factory` issue-intake workflow SHALL declare an authored AWF network policy that allows the default allowlist plus the Node and Go ecosystems, allows `elastic.litellm-prod.ai` for the Claude engine's Anthropic-compatible proxy access, and allows `www.elastic.co` for the Elastic docs MCP server.

#### Scenario: Maintainer inspects workflow frontmatter
- **WHEN** maintainers inspect the authored `code-factory` issue-intake workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`
- **AND** `network.allowed` SHALL include `node`
- **AND** `network.allowed` SHALL include `go`
- **AND** `network.allowed` SHALL include `elastic.litellm-prod.ai`
- **AND** `network.allowed` SHALL include `www.elastic.co`

## ADDED Requirements

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
