## 1. Update change-factory workflow template

- [ ] 1.1 In `.github/workflows-src/change-factory-issue/workflow.md.tmpl`, add `www.elastic.co` to the `network.allowed` list (alongside `defaults`, `node`, `elastic.litellm-prod.ai`)
- [ ] 1.2 In the same template, add an `mcp-servers` frontmatter block with an `elastic-docs` entry: `url: "https://www.elastic.co/docs/_mcp/"` and `allowed: ["*"]`
- [ ] 1.3 In the same template, add a prompt section instructing the agent to use the elastic-docs MCP tools (`search_docs`, `find_related_docs`, `get_document_by_url`) during issue investigation before authoring proposal artifacts, and to proceed gracefully if the MCP is unavailable

## 2. Update code-factory workflow template

- [ ] 2.1 In `.github/workflows-src/code-factory-issue/workflow.md.tmpl`, add `www.elastic.co` to the `network.allowed` list (alongside `defaults`, `node`, `go`, `elastic.litellm-prod.ai`)
- [ ] 2.2 In the same template, add an `mcp-servers` frontmatter block with an `elastic-docs` entry: `url: "https://www.elastic.co/docs/_mcp/"` and `allowed: ["*"]`
- [ ] 2.3 In the same template, add a prompt section instructing the agent to use the elastic-docs MCP tools during issue investigation before implementing provider code, and to proceed gracefully if the MCP is unavailable

## 3. Regenerate compiled workflow artifacts

- [ ] 3.1 Run `make workflow-generate` to recompile both templates and regenerate the four files: `change-factory-issue.md`, `change-factory-issue.lock.yml`, `code-factory-issue.md`, `code-factory-issue.lock.yml`
- [ ] 3.2 Verify the compiled `change-factory-issue.md` frontmatter contains `mcp-servers.elastic-docs` and `www.elastic.co` in `network.allowed`
- [ ] 3.3 Verify the compiled `code-factory-issue.md` frontmatter contains `mcp-servers.elastic-docs` and `www.elastic.co` in `network.allowed`
- [ ] 3.4 Verify the `change-factory-issue.lock.yml` manifest JSON references the MCP gateway configuration (check the `mcpServers` block in the generated `start_mcp_gateway` step)

## 4. Validate OpenSpec artifacts

- [ ] 4.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate factory-workflows-elastic-docs-mcp --type change` and fix any reported problems
