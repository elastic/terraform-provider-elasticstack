# Consulting Elastic documentation

Use this reference when the task requires **content inside JSON-shaped Terraform attributes** — mapping field types, query DSL, ingest processor bodies, detection rule params, ILM phase actions, Kibana alerting rule params, inference model parameters, role query DSL. Consult the Elastic docs to verify the shape of that JSON **before** emitting it inside `jsonencode(...)`.

## When to consult Elastic docs

Consult external Elastic documentation when the user asks about:

- Mapping field types, analyzers, or mapping parameters (e.g. `copy_to`, `fields`, `index_options`).
- Query DSL shape for `filter`, `query`, or `post_filter` attributes.
- Ingest processor bodies beyond what's in `elasticstack_elasticsearch_ingest_processor_*` helpers.
- Alerting rule `params` or action parameters for a specific rule type.
- Detection rule `query` / `threshold` / `threat_mapping` content.
- ILM policy phase actions and their parameters.
- Inference endpoint `service_settings` or `task_settings`.
- Model IDs, tokenizer names, or any Elastic Stack version-specific behavior.

## When NOT to consult Elastic docs

Do not consult external docs for:

- Terraform attribute names, types, defaults, required/optional status, or force-new semantics — use `references/resources/<short_name>.md` or `references/data-sources/<short_name>.md`.
- Provider authentication, connection blocks, or `required_providers` — use `references/provider.md`.
- Deletion protection, JSON normalization, or connection precedence — use `references/gotchas.md`.

The per-entity reference files are generated from the provider source and are more accurate than docs for Terraform-surface questions.

## Preferred tool order

1. The `elastic-docs` MCP server, if the agent runtime has it configured. Use `search_docs` to find the right page, then `get_document_by_url` to read it.
2. If the MCP server is not configured, show the user the configuration snippet below and offer to continue without it.
3. If no docs lookup is available, ask the user to share the relevant Elastic docs link and proceed with their content.

## MCP configuration snippet

Endpoint: `https://www.elastic.co/docs/_mcp/`

Cursor / Claude Code:

```json
{
  "mcpServers": {
    "elastic-docs": {
      "url": "https://www.elastic.co/docs/_mcp/"
    }
  }
}
```

VS Code:

```json
{
  "servers": {
    "elastic-docs": {
      "type": "http",
      "url": "https://www.elastic.co/docs/_mcp/"
    }
  }
}
```

## Output rules for docs-sourced content

- Wrap every JSON fragment from docs with `jsonencode({...})` before placing it in HCL. Never embed as a raw string literal.
- Translate JSON `true` / `false` / `null` to HCL `true` / `false` / `null` inside `jsonencode`.
- When docs show a sample across multiple Elastic Stack versions, pick the one matching the user's target version. Surface the version as part of the assumptions in your output.
- Cite the docs URL in a code comment above the attribute when the content is non-obvious.
