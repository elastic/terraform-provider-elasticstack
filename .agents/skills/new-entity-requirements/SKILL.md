---
name: new-entity-requirements
description: Gathers initial requirements for a new Terraform resource or data source by examining API clients (go-elasticsearch, generated kbapi), Elastic API docs (Elastic docs MCP server and/or web), then interviewing the user for gaps. Use when designing a new entity, drafting requirements from an API, or before implementing a new resource/data source.
---

# New Entity Requirements

Gather **initial requirements** for an entirely new Terraform resource or data source. Sources: repo API client code, Elastic API documentation (via the **Elastic docs MCP server** when available, otherwise web fetch/search), and user input for decisions the code and docs cannot answer.

## Input

- **Entity concept**: User specifies the target (e.g. “Elasticsearch API key”, “Kibana SLO”, “Fleet integration”). Optionally: resource vs data source, proposed type name.
- **API scope**: Which backend (Elasticsearch vs Kibana/Fleet) and, if known, API name or doc URL.

## Workflow

### 1. Resolve API surface

- **Elasticsearch**: Wrappers live in `internal/clients/elasticsearch/` (e.g. `security.go`, `cluster.go`). They call `apiClient.GetESClient()` and use the go-elasticsearch client (e.g. `esClient.Security.PutUser`). Shared models: `internal/models/`. If no wrapper exists yet, search the go-elasticsearch API (dependency `github.com/elastic/go-elasticsearch/v8`) for the relevant namespace (e.g. Security, Watcher, Transform).
- **Kibana and Fleet**: OpenAPI-generated client: `generated/kbapi/kibana.gen.go`. Higher-level wrappers: `internal/clients/kibanaoapi/` and `internal/clients/fleet/` (e.g. `alerting_rule.go`, `connector.go`). Use kbapi types and ClientWithResponses methods for the target API.

Identify: **create/update** (PUT/POST), **read** (GET), **delete** (DELETE), request/response shapes, identifiers (name, id, composite id). Note any version or feature flags in the client or API.

### 2. Examine Elastic API docs

- **Preferred**: Use the **Elastic docs MCP server** when available. Call its tools to search or fetch Elastic documentation (e.g. by API name, topic, or URL). See [reference.md](reference.md) for how to use the MCP server.
- **Fallback**: If the MCP server is not configured or does not return the needed content, fetch docs via web (e.g. `mcp_web_fetch` for a known URL, or web search for the API name).
- **URL patterns**: See [reference.md](reference.md) for elastic.co/guide and elastic.co/docs URL patterns when constructing or citing links.
- **Extract**: Endpoint names, required vs optional fields, validation rules, version requirements, error semantics (e.g. 404 = not found), and whether create/update are the same or separate.
- If the user provided a doc URL, use it (via MCP if the server supports fetch-by-URL, or via web fetch); otherwise search by API name (e.g. “Elasticsearch security API create API key”, “Kibana SLO API”).

### 3. Draft initial OpenSpec spec

- **Path**: `openspec/specs/<capability>/spec.md` (e.g. `openspec/specs/elasticsearch-security-api-key/spec.md`). Follow [`dev-docs/high-level/openspec-requirements.md`](../../../dev-docs/high-level/openspec-requirements.md): `## Purpose`, optional `## Schema`, `## Requirements` with `### Requirement:` / `#### Scenario:`; requirement bodies MUST include **SHALL** or **MUST** (for `openspec validate`).
- **Schema**: From API request/response and docs, draft an HCL-style schema: required/optional/computed attributes and blocks, types, and notes (e.g. “requires Elasticsearch ≥ 8.x”). Mark unknowns as “TBD” or “(to confirm)”.
- **Requirements**: Draft **API** (which endpoints for create/update/read/delete, link to docs), **Identity** (how `id` is formed), **Import** (if resource; id format), **Connection** (provider client; resource-level override if applicable), **Compatibility** (version gates from docs/client). Add **Create/Update**, **Read**, **Delete**, **Mapping**, **State** only where clearly implied by the API; otherwise leave as open questions. For larger features, optionally start under `openspec/changes/<change-id>/` with proposal/design/tasks per OpenSpec’s change workflow.

### 4. Interview the user

- Collect **unanswered questions** from the draft (schema ambiguities, id format, import support, lifecycle, version support). Use the question bank in [reference.md](reference.md).
- Prefer the **AskQuestion** tool when available (one question per call, clear options). Otherwise ask conversationally and list options.
- Record the user’s answers and update the requirements doc: replace TBDs, add or refine requirements, and add any **Lifecycle**, **Plan/State**, or **StateUpgrade** rules they specify.
- If the user defers a decision, leave that item as TBD in the doc and add a short “Open point” note.

### 5. Finalize

- Ensure every requirement is either derived from the API/client/docs or from user answers. Remove or rewrite any that are speculative.
- Add a short “Sources” note at the end if helpful: API client paths, doc URLs used, and whether Elastic docs were obtained via the MCP server or web.

## Output

- **Deliverable**: One OpenSpec spec at `openspec/specs/<capability>/spec.md` with Purpose, Schema (if useful), and Requirements/Scenarios, with TBD/Open points only where the user deferred.
- **Traceability**: Requirements tied to API docs (links) or user decisions; no invented behavior.

## Reference

- Authoring: `dev-docs/high-level/openspec-requirements.md`
- Example (existing entity): `openspec/specs/elasticsearch-security-role/spec.md`
- API client locations, doc URLs, interview question bank: [reference.md](reference.md)
