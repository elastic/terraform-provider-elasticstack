## Why

The `elasticstack_kibana_dashboard` resource cannot create `links` (navigation) panels. Users assembling curated multi-dashboard hierarchies through Terraform have no first-class way to embed dashboard-to-dashboard or dashboard-to-external navigation. The only workaround today is a panel-level `config_json` blob, which sacrifices typed validation and drift-safe planning.

The Kibana Dashboard API defines a full `links` panel type (`kbn-dashboard-panel-type-links`) with two configuration branches and a per-item link union, none of which is currently reachable through the resource.

## What Changes

- Add a new `links_config` typed block to the `panels` list in `elasticstack_kibana_dashboard`, following the same handler-registry pattern as `vis_config`, `discover_session_config`, `image_config`, etc.
- Implement both API branches:
  - `by_value` — inline link configuration with `layout`, common display fields (`title`, `description`, `hide_title`, `hide_border`), and a `links[]` list of typed link items.
  - `by_reference` — references a Kibana Links library saved object by `ref_id`, with the same common display fields.
- Each link item in `by_value.links[]` is a flat object with a required `type` discriminator (`"dashboard"` or `"external"`) plus type-specific fields at the same level; plan-time validators enforce no cross-contamination.
- Naming: `type = "dashboard"` maps to the API `dashboardLink` discriminator; `type = "external"` maps to `externalLink`. The shorter Terraform enum stays consistent with Terraform idiom (the `links` container already supplies the noun).

## Capabilities

### New Capabilities

- `kibana-dashboard-links-panel`: New `links_config` panel block on `elasticstack_kibana_dashboard` panels supporting `by_value` (inline) and `by_reference` (library-item) configuration.

### Modified Capabilities

- `kibana-dashboard`: extends the panel registry to include the `links` handler.

## Impact

- New package: `internal/kibana/dashboard/panel/links/` — schema, model, api.go (ToAPI/FromAPI), populate.go.
- `internal/kibana/dashboard/registry.go` — register `links.Handler{}`.
- `openspec/specs/kibana-dashboard/spec.md` — add `links_config` schema entry (new REQ).
- Acceptance tests: at least one `by_value` scenario (both link types) and one `by_reference` scenario in `internal/kibana/dashboard/panel/links/acc_test.go`.
