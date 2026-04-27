## Why

The change-factory and code-factory agent workflows have no access to Elastic documentation during issue investigation, forcing the agent to operate without authoritative API reference when authoring OpenSpec proposals or implementing provider code for unfamiliar or newly-introduced Elastic features. The Elastic docs MCP server provides a public, unauthenticated HTTP endpoint with structured semantic search that solves this directly.

## What Changes

- Add the Elastic docs MCP server (`https://www.elastic.co/docs/_mcp/`) to the `mcp-servers` frontmatter block of both factory workflow templates, giving the agent access to `search_docs`, `find_related_docs`, `get_document_by_url`, and related tools.
- Add `www.elastic.co` to the `network.allowed` list in both workflow templates (required for MCP gateway outbound calls if squid is in the path; to be confirmed by testing).
- Update the agent prompt in both workflow templates to instruct when and how to use the docs tools during issue investigation.
- Regenerate the compiled `.md` and `.lock.yml` workflow artifacts from the updated templates.

## Capabilities

### New Capabilities

None — this change does not introduce new Terraform resources or data sources.

### Modified Capabilities

- `ci-change-factory-issue-intake`: Adds a requirement that the proposal agent has access to Elastic documentation via the elastic-docs MCP server during issue investigation.
- `ci-code-factory-issue-intake`: Adds a requirement that the implementation agent has access to Elastic documentation via the elastic-docs MCP server during issue investigation.

## Impact

- `.github/workflows-src/change-factory-issue/workflow.md.tmpl` — source template to modify
- `.github/workflows-src/code-factory-issue/workflow.md.tmpl` — source template to modify
- `.github/workflows/change-factory-issue.md` — regenerated compiled output
- `.github/workflows/change-factory-issue.lock.yml` — regenerated compiled output
- `.github/workflows/code-factory-issue.md` — regenerated compiled output
- `.github/workflows/code-factory-issue.lock.yml` — regenerated compiled output
- No Go source, Terraform provider code, tests, or generated clients are affected.
