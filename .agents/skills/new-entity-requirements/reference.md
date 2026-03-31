# New Entity Requirements — API Clients, Docs, and Interview Questions

## API client locations

### Elasticsearch

| What | Where |
|------|--------|
| Provider client | `apiClient.GetESClient()` returns the go-elasticsearch client. |
| Wrappers | `internal/clients/elasticsearch/`: `security.go`, `cluster.go`, `watch.go`, `transform.go`, `index.go`, `enrich.go`, `ml_job.go`, `logstash.go`. |
| Raw API | Dependency `github.com/elastic/go-elasticsearch/v8`; e.g. `esClient.Security.*`, `esClient.Watcher.*`, `esClient.TransformGetTransform.*`. |
| Models | `internal/models/` (e.g. `user.go`, `role.go`, `api_key.go`). Request/response structs used by wrapper functions. |

**Finding an API**: If there is no wrapper in `internal/clients/elasticsearch/`, search the codebase for the API name or inspect the go-elasticsearch package for the relevant namespace (Security, Indices, Cluster, etc.) and HTTP method (Put, Get, Delete).

### Kibana

| What | Where |
|------|--------|
| Generated client | `generated/kbapi/kibana.gen.go` — OpenAPI-generated types and `ClientWithResponses` interface. Large file; search for API/type names. |
| Wrappers | `internal/clients/kibanaoapi/`: `alerting_rule.go`, `connector.go`, `dashboards.go`, `data_views.go`, `exceptions.go`, `maintenance_window.go`, `prebuilt_rules.go`, `security_enable_rule.go`, `security_lists.go`, `client.go`. |
| Client construction | `internal/clients/kibanaoapi/client.go`: `NewClient(cfg)`, `Client.API` is `*kbapi.ClientWithResponses`. |

**Finding an API**: Search `generated/kbapi/kibana.gen.go` for the feature or endpoint name (e.g. “Slo”, “Connector”). Then check `internal/clients/kibanaoapi/` for a wrapper that calls that API.

### Fleet

| What | Where |
|------|--------|
| Client | `internal/clients/fleet/`: `client.go`, `fleet.go`. |
| Fleet APIs | Documented at Elastic Fleet API |
| Generated client | `generated/kbapi/kibana.gen.go` — OpenAPI-generated types and `ClientWithResponses` interface. Large file; search for API/type names. |

---

## Elastic API documentation

### Elastic docs MCP server (preferred)

When the **Elastic docs MCP server** is configured, use it as the primary source for Elastic documentation:

- **Before fetching docs**: Check whether an MCP server exposes Elastic-docs tools (e.g. search, fetch by URL or topic). Use `call_mcp_tool` with that server and the appropriate tool to retrieve doc content.
- **When to use**: For any Elasticsearch, Kibana, or Fleet API documentation needed in step “Examine Elastic API docs”. Prefer MCP over web fetch when the server returns relevant, up-to-date content.
- **How to use**: Invoke the MCP tool with query parameters that match the entity (e.g. API name like “security API put role”, “create API key”, “Kibana SLO API”) or with a specific doc URL if the tool supports it. Use the returned content to extract endpoints, request/response shape, version notes, and error behavior.
- **If MCP is unavailable**: Fall back to URL patterns below and web fetch (`mcp_web_fetch`) or web search for the same information.

### URL patterns (fallback or for citing links)

- **Elasticsearch Reference**: `https://www.elastic.co/guide/en/elasticsearch/reference/current/<topic>.html`  
  Example: `security-api-put-role.html`, `security-api-create-api-key.html`, `security-api-get-role.html`.
- **Kibana Guide**: `https://www.elastic.co/guide/en/kibana/current/<topic>.html`  
  Example: `add-monitor-api.html`, `create-private-location-api.html`, `spaces-api-post.html`.
- **REST APIs (alternative)**: `https://www.elastic.co/docs/api/...` (e.g. `/api/doc/elasticsearch`, `/api/doc/kibana`).

When not using the Elastic docs MCP server, search the web or fetch these URLs for the API name (e.g. “Elasticsearch create API key API”, “Kibana SLO API”) to get the exact doc page. Existing requirements docs and `docs/resources/*.md` often link to these URLs — reuse the same link style in new requirements when citing sources.

### What to pull from docs

- Endpoint path and method (PUT, GET, POST, DELETE).
- Request body: required vs optional fields, nested objects, allowed values.
- Response: shape, identifier(s) returned (e.g. `id`, `name`).
- Version notes (“added in 8.x”, “deprecated in 8.y”).
- Errors: 404 behavior, validation errors, rate limits.

---

## Interview question bank

Use these when the API and client do not fully determine behavior. Prefer **AskQuestion** with concrete options when available; otherwise ask in chat and list options.

### Identity and import

- **Resource identifier**: “How should the Terraform resource be identified in state? Options: (A) name only, (B) composite id (e.g. `cluster_uuid/name`), (C) server-generated id only, (D) other (describe).”
- **Import**: “Should this resource support import? If yes, what format should the import `id` use (e.g. `cluster_uuid/name`, or a single id)?”

### Schema and lifecycle

- **Required vs optional**: “For [field X] from the API: should it be required in Terraform, or optional with a default?”
- **Replacement**: “When [e.g. name] changes, should the resource be replaced (destroy + create) or updated in place?”
- **Sensitive fields**: “Which attributes should be marked sensitive (e.g. passwords, tokens) and not shown in plan/apply output?”

### Connection and compatibility

- **Connection override**: “Should this resource support a resource-level connection block (e.g. `elasticsearch_connection` / `kibana_connection`) to override the provider, or always use the provider client?”
- **Minimum version**: “What is the minimum Elasticsearch/Kibana version for this feature? Should the provider fail with an ‘Unsupported Feature’ error when the server is older?”

### State and mapping

- **Empty vs null**: “When the API returns an empty list for [field], should Terraform store null or an empty list? (Affects drift and optional+computed behavior.)”
- **Read-only fields**: “Which response fields should be computed-only (e.g. `created_at`, `version`) and not configurable?”
- **JSON attributes**: “Should [complex object] be exposed as a JSON string (normalized) or as nested blocks/attributes?”

### Scope and naming

- **Resource vs data source**: “Is this entity a managed resource (create/update/delete) or a read-only data source?”
- **Type name**: “Proposed Terraform type name (e.g. `elasticstack_elasticsearch_security_api_key`). Confirm or suggest alternative.”

### Deferred decisions

- If the user defers: “We’ll leave [X] as TBD and add an Open point in the requirements doc. You can decide later during implementation.”

Record each answer in the requirements doc (update schema, add/revise requirements, or add an Open point).
