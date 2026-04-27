## Context

Both factory workflows (change-factory and code-factory) are authored as gh-aw (GitHub Agentic Workflow) markdown templates under `.github/workflows-src/<name>/workflow.md.tmpl` and compiled to `.github/workflows/<name>.md` + `.github/workflows/<name>.lock.yml` via `make workflow-generate`. The `.md` and `.lock.yml` files must never be hand-edited.

The gh-aw runtime supports three types of MCP servers configurable in the workflow frontmatter:
- **Container**: Docker image run by the MCP gateway
- **HTTP**: Remote server accessed via HTTPS from the MCP gateway  
- **Stdio**: Local command invoked on the runner host

The Elastic docs MCP server (`https://www.elastic.co/docs/_mcp/`) is a public, unauthenticated, stateless HTTP endpoint exposing tools via Streamable HTTP. No Docker image or npm package is required.

The gh-aw MCP gateway runs with `--network host`, so its outbound HTTP calls may or may not traverse the squid firewall that backs `network.allowed`. To be safe, `www.elastic.co` is added to `network.allowed` in both templates; this can be removed if testing shows it is unnecessary.

## Goals / Non-Goals

**Goals:**
- Give both factory agents structured, semantic access to Elastic documentation during issue investigation
- Keep changes confined to the workflow template source files (no Go, no Terraform provider code)
- Remain compliant with the workflows-src compile pattern

**Non-Goals:**
- Adding authentication or rate-limiting to the MCP access (the endpoint is public)
- Modifying other workflows (kibana-spec-impact, openspec-verify-label, etc.)
- Restricting which tools within the MCP the agent may call (all 6 are permitted)

## Decisions

**Decision: HTTP MCP over network-only access**

The alternative (adding `www.elastic.co` to `network.allowed` only, relying on raw WebFetch) gives the agent access to HTML pages it would need to guess the URL for. The HTTP MCP provides `search_docs` for semantic discovery without knowing a URL in advance, which is the common case when investigating an unfamiliar Elastic API feature. The MCP produces structured, token-efficient results vs. full HTML pages.

**Decision: Use `allowed: ["*"]` rather than an explicit tool allowlist**

The elastic-docs MCP exposes only docs-related tools (search_docs, find_related_docs, get_document_by_url, analyze_document_structure, check_docs_coherence, find_docs_inconsistencies). None of these are write operations. Listing `"*"` avoids breakage if the upstream server adds new docs tools, and there is no security concern for a read-only public endpoint.

**Decision: Add `www.elastic.co` to `network.allowed` in both templates**

Whether the MCP gateway's outbound HTTP calls traverse squid is undocumented and may depend on gh-aw version. Adding the host is harmless and avoids a subtle failure mode. This can be revisited and removed once testing confirms it is unnecessary.

**Decision: Apply to both factory workflows symmetrically**

The code-factory agent benefits equally: when implementing a Terraform resource for a new or updated Elastic API, consulting the API reference reduces guesswork about parameter names, types, and behavior.

## Risks / Trade-offs

**[Risk] Elastic docs MCP is Technical Preview** → The endpoint may be unstable or change URL/behaviour. Mitigation: the failure mode is "agent cannot look up docs" (a graceful degradation), not a data-loss or incorrect-output risk. The prompt guidance instructs the agent to proceed without docs if the tools are unavailable.

**[Risk] `www.elastic.co` in network.allowed is unnecessary** → Minor: adds a host to the allowlist that the agent could use via raw WebFetch. The agent prompt does not encourage raw fetching; this is cosmetic overhead. Mitigation: remove the entry after confirming MCP gateway networking in a test run.

**[Risk] MCP gateway network mode changes in a future gh-aw version** → If a future gh-aw release routes MCP gateway traffic through squid, the `www.elastic.co` allowlist entry is already in place and no action is needed.

## Migration Plan

1. Edit both `workflow.md.tmpl` source templates (add `mcp-servers` block, update `network.allowed`, add prompt guidance section).
2. Run `make workflow-generate` to regenerate compiled outputs.
3. Verify compiled `.md` files contain the expected `mcp-servers` JSON in the lock manifest.
4. Commit both source and compiled files together.
5. Trigger a test `change-factory` issue referencing a specific Elastic API to verify `search_docs` is called during the agent run.
6. If `www.elastic.co` proves unnecessary in `network.allowed`, remove it in a follow-up commit.

No rollback strategy beyond reverting the commit is needed — the change is additive and the workflow degrades gracefully if the MCP endpoint is down.
