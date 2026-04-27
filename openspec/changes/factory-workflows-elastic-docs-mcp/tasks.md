## 1. Update change-factory workflow template

- [x] 1.1 In `.github/workflows-src/change-factory-issue/workflow.md.tmpl`, add `www.elastic.co` to the `network.allowed` list (alongside `defaults`, `node`, `elastic.litellm-prod.ai`)
- [x] 1.2 In the same template, add an `mcp-servers` frontmatter block with an `elastic-docs` entry: `url: "https://www.elastic.co/docs/_mcp/"` and `allowed: ["*"]`
- [x] 1.3 In the same template, add an `## Elastic documentation` prompt section immediately after `## OpenSpec tooling` (around line 148) that: (a) tells the agent an `elastic-docs` MCP server is available with tools `search_docs`, `find_related_docs`, and `get_document_by_url`; (b) instructs it to use `search_docs` to look up API behaviour, parameters, and constraints for the feature described in the issue before authoring proposal artifacts; and (c) instructs it to proceed without docs if the MCP tools are unavailable rather than blocking the run

## 2. Update code-factory workflow template

- [x] 2.1 In `.github/workflows-src/code-factory-issue/workflow.md.tmpl`, add `www.elastic.co` to the `network.allowed` list (alongside `defaults`, `node`, `go`, `elastic.litellm-prod.ai`)
- [x] 2.2 In the same template, add an `mcp-servers` frontmatter block with an `elastic-docs` entry: `url: "https://www.elastic.co/docs/_mcp/"` and `allowed: ["*"]`
- [x] 2.3 In the same template, add an `## Elastic documentation` prompt section immediately after `## Test environment` and before `## Task` (around line 172) that: (a) tells the agent an `elastic-docs` MCP server is available with tools `search_docs`, `find_related_docs`, and `get_document_by_url`; (b) instructs it to use `search_docs` to look up API behaviour, parameters, and constraints for the feature described in the issue before writing implementation code; and (c) instructs it to proceed without docs if the MCP tools are unavailable rather than blocking the run

## 3. Regenerate compiled workflow artifacts

- [x] 3.1 Run `make workflow-generate` to recompile both templates and regenerate the four files: `change-factory-issue.md`, `change-factory-issue.lock.yml`, `code-factory-issue.md`, `code-factory-issue.lock.yml`
- [x] 3.2 Verify the compiled `change-factory-issue.md` frontmatter contains `mcp-servers.elastic-docs` and `www.elastic.co` in `network.allowed`
- [x] 3.3 Verify the compiled `code-factory-issue.md` frontmatter contains `mcp-servers.elastic-docs` and `www.elastic.co` in `network.allowed`
- [x] 3.4 Verify the `change-factory-issue.lock.yml` manifest JSON references the MCP gateway configuration (check the `mcpServers` block in the generated `start_mcp_gateway` step)

## 4. Validate OpenSpec artifacts

- [x] 4.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate factory-workflows-elastic-docs-mcp --type change` and fix any reported problems
