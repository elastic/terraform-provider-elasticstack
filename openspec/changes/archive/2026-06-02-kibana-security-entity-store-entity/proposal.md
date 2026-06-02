## Why

Practitioners cannot manage **Entity Store entity records** via Terraform today (parent [#2123](https://github.com/elastic/terraform-provider-elasticstack/issues/2123)). The Kibana Security Entity Store API supports creating, updating, and deleting individual entity records (host, user, service, generic), and exposes a list/search endpoint for querying them. Teams that manage Elastic Security environments as code need a way to seed or maintain curated entity records — for example, marking authoritative hosts as assets, attaching risk enrichment anchors, or configuring service identity — alongside their other Terraform-managed Elastic resources.

## What Changes

Add three new Terraform entities backed by the Kibana Security Entity Store API (requires Elastic Stack ≥ 9.1.0):

1. **Resource** `elasticstack_kibana_security_entity_store_entity` — manages a single entity record (create, read, update, delete, import) with first-class typed attributes for the full API body and JSON escape-hatch fallbacks for each top-level section.
2. **Data source** `elasticstack_kibana_security_entity_store_entities` — list/search data source exposing both page-based and cursor-based pagination modes via the `GET /api/security/entity_store/entities` endpoint, with an optional `entity_id` filter for single-entity lookup.

### Schema sketch

**Resource** `elasticstack_kibana_security_entity_store_entity`:

```hcl
resource "elasticstack_kibana_security_entity_store_entity" "example" {
  space_id    = "default"          # optional, computed; RequiresReplace
  entity_type = "host"             # required; "user"|"host"|"service"|"generic"; RequiresReplace
  entity_id   = "host:web-01"      # required; RequiresReplace; must match entity.id

  timestamp   = "2024-01-01T00:00:00Z"  # optional; maps to @timestamp

  # Typed top-level blocks (each conflicts with its matching _json fallback)
  entity { id = "host:web-01"; name = "web-01"; type = "host"; source = ["manual"] }
  host   { name = "web-01"; ip = ["10.0.1.42"] }
  asset  { ... }
  user   { ... }
  service { ... }
  cloud  { ... }
  orchestrator { ... }
  labels = { env = "prod" }        # optional map
  tags   = ["terraform"]           # optional set(string)

  # JSON fallback escape hatches (each conflicts with its typed block)
  entity_json      = jsonencode({...})
  host_json        = jsonencode({...})
  asset_json       = jsonencode({...})
  user_json        = jsonencode({...})
  service_json     = jsonencode({...})
  cloud_json       = jsonencode({...})
  orchestrator_json = jsonencode({...})
  labels_json      = jsonencode({...})

  # Update control
  force = false   # optional bool; passed as ?force=true on PUT only

  # Computed outputs
  id            = "<computed>"    # composite: "<space_id>/<entity_id>"
  document_json = "<computed>"    # assembled API document read back from Kibana
  response_json = "<computed>"    # raw API response for troubleshooting

  kibana_connection { ... }
}
```

**Data source** `elasticstack_kibana_security_entity_store_entities`:

```hcl
data "elasticstack_kibana_security_entity_store_entities" "example" {
  space_id      = "default"    # optional, computed

  # Single-entity lookup (convenience)
  entity_id     = "host:web-01"

  # Search-after mode (cursor pagination)
  filter        = "entity.type: host"
  size          = 20
  search_after  = jsonencode([...])
  source        = ["entity.id", "entity.name"]
  fields        = ["entity.type"]

  # Page mode
  sort_field    = "entity.name"
  sort_order    = "asc"
  page          = 1
  per_page      = 50
  filter_query  = "entity.source: manual"

  # Common filter
  entity_types  = ["host", "user"]

  # Computed
  results_json  = "<computed>"
}
```

### API operations used

- `POST /api/security/entity_store/entities/{entityType}` — create (HTTP 200 on success; 409 if entity ID already exists).
- `PUT /api/security/entity_store/entities/{entityType}` — update (supports `?force=true` query parameter).
- `DELETE /api/security/entity_store/entities/` — delete; entity ID supplied in JSON request body as `{ "entityId": "<id>" }`.
- `GET /api/security/entity_store/entities` — list/search; supports both page-based and cursor-based pagination (modes cannot be combined). Used for resource read (filter by `entity.id`).

### Read strategy

The API has no `GET /entities/{id}` endpoint. The resource's Read callback calls `GET /api/security/entity_store/entities` with the most deterministic filter expression available (`entity.id: "<id>"`) to retrieve a unique record by ID. Single-entity lookup is also supported via the list data source using the `entity_id` filter parameter.

### Version gating

Enforce `EnforceMinVersion` at Elastic Stack `9.1.0`; adjust if acceptance testing reveals a different minimum.

### Response parsing

The generated client (`kbapi`) returns `[]byte` for these endpoints. Implementation uses local response structs or `encoding/json` generic parsing rather than expecting typed response models from the generated client.

### Conflict validation

Each typed top-level block (`entity`, `host`, `user`, `service`, `cloud`, `asset`, `orchestrator`) MUST conflict with its corresponding JSON fallback attribute (`entity_json`, `host_json`, `user_json`, `service_json`, `cloud_json`, `asset_json`, `orchestrator_json`). The `labels` map conflicts with `labels_json`. Conflict is enforced at plan time.

## Capabilities

### New Capabilities

- `kibana-security-entity-store-entity`: Resource CRUD and import for a single Entity Store entity record with typed attributes, JSON escape hatches, conflict validation, `force` update support, and computed `document_json`/`response_json` outputs.
- `kibana-security-entity-store-entities-datasource`: List/search data source with page-based and cursor-based pagination, entity type filtering, optional single-entity lookup via `entity_id`, and computed `results_json`.

### Modified Capabilities

_\(none — both entities are net-new\)_

## Impact

- **New specs** (delta under this change):
  - `openspec/changes/kibana-security-entity-store-entity/specs/kibana-security-entity-store-entity/spec.md` (resource)
  - `openspec/changes/kibana-security-entity-store-entity/specs/kibana-security-entity-store-entities-datasource/spec.md` (list/search data source, including single-entity lookup via `entity_id`)
- **Implementation** (future): `internal/kibana/security_entity_store/` (resource schema, models, CRUD, plan modifier), `internal/clients/kibanaoapi/` or inline JSON parsing for API response handling, provider registration, docs/descriptions, acceptance tests.
- No changes to existing resources or data sources.
