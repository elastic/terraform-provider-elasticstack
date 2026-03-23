---
name: new-entity-requirements
description: Gathers initial requirements for a new Terraform resource or data source by examining API clients (go-elasticsearch, generated kbapi), Elastic API docs (Elastic docs MCP server and/or web), then interviewing the user for gaps. Produces an OpenSpec proposal (change with proposal, design, tasks, and delta specs)—not a hand-written spec under openspec/specs/ alone. Use when designing a new entity, drafting requirements from an API, or before implementing a new resource/data source.
---

# New Entity Requirements

Gather **initial requirements** for an entirely new Terraform resource or data source, then **materialize them as an OpenSpec proposal**: a change under `openspec/changes/<name>/` with `proposal.md`, `design.md`, `tasks.md`, and delta capability specs. Sources: repo API client code, Elastic API documentation (via the **Elastic docs MCP server** when available, otherwise web fetch/search), and user input for decisions the code and docs cannot answer.

For the **CLI sequence** (create change, resolve artifact order, run `openspec instructions`, write files), follow [openspec-propose](../openspec-propose/SKILL.md). This skill adds **what to research** and **what to put in each artifact** for a new Terraform entity.

## Input

- **Entity concept**: User specifies the target (e.g. “Elasticsearch API key”, “Kibana SLO”, “Fleet integration”). Optionally: resource vs data source, proposed type name.
- **API scope**: Which backend (Elasticsearch vs Kibana/Fleet) and, if known, API name or doc URL.
- **Change name** (optional): Kebab-case id for `openspec new change` (e.g. `add-elasticsearch-security-api-key-resource`). If missing, derive one from the entity and confirm with the user if ambiguous.

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

### 3. Create the OpenSpec proposal (change + artifacts)

Do **not** write directly to canonical `openspec/specs/<capability>/spec.md` as the primary deliverable. Instead:

1. **Create the change**  
   Run `openspec new change "<name>"` (requires OpenSpec CLI; `make setup` installs it). This creates `openspec/changes/<name>/` with `.openspec.yaml`.

2. **Build all apply-ready artifacts**  
   Follow [openspec-propose](../openspec-propose/SKILL.md) steps 3–4: `openspec status --change "<name>" --json`, then for each artifact in dependency order use `openspec instructions <artifact-id> --change "<name>" --json`, read dependencies, write to `outputPath` using `template` and `instruction`, re-run status until every id in `applyRequires` is `done`.

3. **Entity-specific content**
   - **proposal.md**: What & why for this resource/data source; problem, scope, non-goals if useful; link to Elastic docs URLs gathered in step 2.
   - **design.md**: How the provider will map API ↔ Terraform (client package, identity, import id shape, version gates, error handling). Reference client paths from step 1.
   - **tasks.md**: Concrete implementation steps (schema, CRUD, acceptance tests, docs)—aligned with repo conventions in `dev-docs/high-level/`.
   - **Delta spec(s)** (`openspec/changes/<name>/specs/.../spec.md`): Normative requirements per [`dev-docs/high-level/openspec-requirements.md`](../../../dev-docs/high-level/openspec-requirements.md): `## Purpose`, optional `## Schema`, `## Requirements` with `### Requirement:` / `#### Scenario:`; bodies MUST include **SHALL** or **MUST** (for `openspec validate`). Draft HCL-style schema, API/identity/import/connection/compatibility requirements from steps 1–2. Mark unknowns as “TBD” or “(to confirm)”.

Use **TodoWrite** to track artifact creation, as in openspec-propose.

### 4. Interview the user

- Collect **unanswered questions** from the draft artifacts (schema ambiguities, id format, import support, lifecycle, version support). Use the question bank in [reference.md](reference.md).
- Prefer the **AskQuestion** tool when available (one question per call, clear options). Otherwise ask conversationally and list options.
- Record answers by updating the **delta spec** and, where relevant, **proposal** or **design**—replace TBDs, refine requirements, add **Lifecycle**, **Plan/State**, or **StateUpgrade** if the user specifies them.
- If the user defers a decision, leave TBD in the delta spec and add a short “Open point” note.

### 5. Finalize

- Ensure every requirement is either derived from the API/client/docs or from user answers. Remove or rewrite speculative claims.
- Add a short “Sources” note in **design.md** or the delta spec if helpful: API client paths, doc URLs, and whether Elastic docs came from the MCP server or web.
- Run `openspec status --change "<name>"` and confirm the change is apply-ready. Validate per project docs (`make check-openspec` / `openspec validate` as appropriate for specs in the change).

## Output

- **Deliverable**: A complete OpenSpec change at `openspec/changes/<name>/` with all artifacts required for implementation (`applyRequires`), including delta capability specs—not a standalone file only under `openspec/specs/`.
- **Traceability**: Requirements in delta specs tied to API docs (links) or user decisions; no invented behavior.

## Reference

- Proposal workflow (CLI): [openspec-propose](../openspec-propose/SKILL.md)
- Authoring: `dev-docs/high-level/openspec-requirements.md`
- Example canonical spec (style reference): `openspec/specs/elasticsearch-security-role/spec.md`
- Example archived change layout: `openspec/changes/archive/` (proposal, design, tasks, `specs/…/spec.md`)
- API client locations, doc URLs, interview question bank: [reference.md](reference.md)
