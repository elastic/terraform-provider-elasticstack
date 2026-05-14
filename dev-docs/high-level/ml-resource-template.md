# ML Resource Template (Plugin Framework)

This guide captures a repeatable pattern for adding Elasticsearch ML resources in this provider.
Use it as a practical checklist when implementing resources similar to `ml_calendar`, `ml_calendar_event`, and `ml_filter`.

## 1) Start with entitycore envelope by default

Prefer `entitycore.NewElasticsearchResource(...)` unless the resource has non-standard lifecycle behavior.

- Implement a model with:
  - `GetID()`
  - `GetResourceID()`
  - `GetElasticsearchConnection()`
- Provide:
  - schema factory (`getSchema`)
  - read callback
  - delete callback
  - create callback (if straightforward)

Only override `Create`/`Update` directly on the resource when the API behavior requires custom orchestration.

## 2) Decide ID shape early

Resource IDs in state are composite: `<cluster_uuid>/<resource_identifier>`.

- For simple identifiers, use standard composite parsing.
- For identifiers that can contain slashes (for example `<calendar_id>/<event_id>`), use Elasticsearch-specific parsing that splits only on the first slash and then parse the resource segment separately.
- In `ImportState`, set both `id` and key identifying attributes.

## 3) Define explicit model mappings

Keep Terraform and API payload models explicit:

- `TFModel` for state/config
- `CreateAPIModel` for create payload
- `UpdateAPIModel` when update uses patch semantics
- `APIModel` for read mapping

Handle null-vs-empty semantics intentionally for optional strings and optional collections to reduce drift and plan noise.

## 4) Choose an update strategy

Use the envelope update callback only when updates are direct and self-contained.

Override `Update` when you need:

- plan vs state comparison
- remote current-state fetch before update
- add/remove diff computation
- multiple API calls for one Terraform update

If the API has no update endpoint, make mutable fields `RequiresReplace` and return a clear diagnostic in `Update`.

## 5) Keep read authoritative

Read should normalize the final Terraform representation:

- map API response into the TF model
- apply null/empty normalization
- treat not-found as drift (`RemoveResource` behavior)

All create/update flows should re-read and persist canonical state.

## 6) Handle server-generated identifiers robustly

If create does not reliably return a new sub-identifier:

- prefer ID from create response when available
- fallback to deterministic lookup (for example pre/post snapshot + matching on stable fields)
- fail with a clear error if identity cannot be determined

## 7) Validation layering

- Schema validators: structural constraints (regex, length, basic format).
- `ValidateConfig`: cross-field checks (for example `end_time > start_time`).
- API errors: propagate with contextual diagnostics.

## 8) Test matrix (minimum)

For each resource:

- Acceptance tests:
  - create/read
  - update (if supported)
  - import
  - at least one edge case (empty/null/set behavior)
- Unit tests:
  - model conversion and null/empty semantics
  - diff logic for update payloads
  - custom create identity resolution, if applicable

## 9) Docs and CI checklist

- Register the resource in `provider/plugin_framework.go`.
- Regenerate docs (`make docs-generate` or `make gen` as appropriate).
- Ensure lint/docs checks run clean on the branch.
- Keep PR changelog section in required repository format.

## Quick scaffold checklist

- `schema.go`
- `models.go`, `models_test.go`
- `resource.go`
- `create.go`, `read.go`, `delete.go`
- `update.go` (if needed)
- `acc_test.go` (+ test configs)
- provider registration
- docs regeneration
