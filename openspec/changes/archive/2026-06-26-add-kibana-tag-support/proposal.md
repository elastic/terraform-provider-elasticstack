## Why

Practitioners cannot manage Kibana tags as code today. Tags are used to categorize and filter Kibana saved objects (dashboards, visualizations, etc.), and teams managing large Kibana deployments across multiple environments want to provision and reference tags declaratively alongside the objects that use them.

The existing workaround — `elasticstack_kibana_import_saved_objects` — requires managing internal fields (`typeMigrationVersion`, `coreMigrationVersion`, `buildNum`) that change across upgrades, making it brittle. A first-party community implementation uses the legacy `/api/saved_objects_tagging/tags/...` endpoints, which are now deprecated.

Kibana [PR elastic/kibana#266220](https://github.com/elastic/kibana/pull/266220) ("Tags as code CRUD Endpoints") merged on 2026-05-22 and ships in **Kibana 9.5.0**. It exposes a stable, public REST API at `/api/tags`. The generated `kbapi` client already exposes the full surface (`Client.GetTags`, `PostTags`, `GetTagsId`, `PutTagsId`, `DeleteTagsId`), so we can implement this resource against a stable, documented API.

## What Changes

- Add `elasticstack_kibana_tag` **resource** (Plugin Framework) for full CRUD management of a single Kibana tag.
- Add `elasticstack_kibana_tags` **data source** for listing and filtering existing tags by name/description.
- **Out of scope for this proposal artifact**: editing `openspec/specs/` directly; that happens when the change is synced or archived.

### Schema sketch

**`elasticstack_kibana_tag` resource:**

```hcl
resource "elasticstack_kibana_tag" "example" {
  name        = "production"
  color       = "#FF0000"    # Optional; server-generates a random color if omitted
  description = "Production environment resources"
  tag_id      = "my-custom-uuid"  # Optional; forces PUT-to-create semantics
  space_id    = "default"         # Optional; ForceNew
}
```

Attributes:
- `name` (Required string) — display name of the tag
- `tag_id` (Optional string, ForceNew) — client-specified UUID; if unset, server mints the ID on POST
- `color` (Optional+Computed string, `UseStateForUnknown`) — hex color (e.g. `#772299`); server generates a random color when omitted
- `description` (Optional string) — free-text description
- `space_id` (Optional+Computed string, ForceNew, default `"default"`) — Kibana space
- `id` (Computed string) — composite `"<space_id>/<tag_id>"`
- `created_at` (Computed string) — ISO 8601 timestamp
- `updated_at` (Computed string) — ISO 8601 timestamp
- `managed` is **not exposed**; the resource guards against Kibana-managed tags and refuses to read/update/delete them

Import: `terraform import elasticstack_kibana_tag.example default/my-tag-uuid`

**`elasticstack_kibana_tags` data source:**

```hcl
data "elasticstack_kibana_tags" "all" {
  query    = "production"  # Optional; Elasticsearch simple_query_string on name+description
  space_id = "default"     # Optional
}

output "tag_ids" {
  value = data.elasticstack_kibana_tags.all.tags[*].id
}
```

Attributes:
- `query` (Optional string) — `simple_query_string` filter on `name` and `description`
- `space_id` (Optional string, default `"default"`) — Kibana space
- `tags` (Computed list of objects) — each element: `{ id, name, color, description, managed, created_at, updated_at }`

### Version gating

Both entities require **Kibana ≥ 9.5.0**. The provider SHOULD return a clear diagnostic when connected to an older Kibana version, consistent with the `alert_delay`/`flapping` version-gate pattern.

### Acceptance tests

Acceptance tests MUST be skipped when the connected Kibana version is below 9.5.0.

## Capabilities

### New Capabilities

- `kibana-tag`: `elasticstack_kibana_tag` resource and `elasticstack_kibana_tags` data source covering the full CRUD lifecycle, managed-tag guard, composite ID, `tag_id`-specified creation, version gate (Kibana ≥ 9.5.0), and data-source pagination.

### Modified Capabilities

- _(none)_

## Impact

- **Specs**: Delta under `openspec/changes/add-kibana-tag-support/specs/kibana-tag/spec.md` until merged into canonical spec.
- **Implementation** (future): new package `internal/kibana/tag/` (resource, data source, schema, models, CRUD); new `internal/clients/kibanaoapi/tag.go` (API wrapper and tag domain model); provider registration; documentation; acceptance tests.
