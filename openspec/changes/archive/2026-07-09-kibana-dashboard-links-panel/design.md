## Context

The Kibana Dashboard API exposes a `links` panel type (`KibanaHTTPAPIsKbnDashboardPanelTypeLinks`) in `generated/kbapi/kibana.gen.go`. Its `config` field is a union of two branches:

- **By-value** (`KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0`) — inline link list. Fields: `layout` (enum, required: `"horizontal"` | `"vertical"`), `links[]` (required, non-empty), plus optional `title`, `description`, `hide_title`, `hide_border`.
- **By-reference** (`KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1`) — library saved object. Fields: `ref_id` (required), plus optional `title`, `description`, `hide_title`, `hide_border`.

Each link item is a union:

- `KibanaHTTPAPIsKbnLinkPanelTypeDashboardLink`: `type = "dashboardLink"`, `destination` (dashboard saved-object id), optional `label`, optional `options { open_in_new_tab, use_filters, use_time_range }`.
- `KibanaHTTPAPIsKbnLinkTypeExternalLink`: `type = "externalLink"`, `destination` (URL), optional `label`, optional `options { open_in_new_tab, encode_url }`.

The sibling panel implementations (`vis_config`, `discover_session_config`, `image_config`) provide the established pattern: a per-panel package under `internal/kibana/dashboard/panel/<type>/` with `schema.go`, `model.go`, `api.go`, and `populate.go`, registered in `internal/kibana/dashboard/registry.go` via `panelHandlers`.

## Goals / Non-Goals

**Goals:**
- Full parity with both API branches (`by_value`, `by_reference`) in the initial delivery.
- Plan-time validators that enforce the by-value/by-reference mutual exclusion and the per-link-type field constraints.
- REQ-009 null-preservation: optional display fields (`title`, `description`, `hide_title`, `hide_border`) stay null in state when the user omits them, even if Kibana echoes defaults.

**Non-Goals:**
- A separate `elasticstack_kibana_links_library_item` resource for authoring library items. `by_reference` only accepts an id.
- Panel-level drilldowns on `links` panels — the API model does not expose them.

## Decisions

**1. Both branches ship together.**
Rationale: parity with `vis_config` and `discover_session_config`, which both ship `by_value`/`by_reference`. The `by_reference` branch adds ~15 schema lines (`ref_id` + shared display fields), reuses panelkit helpers, and costs little.

**2. Flat link items with a `type` discriminator.**
Each item in `links_config.by_value.links[]` is a flat object:
```hcl
links = [
  { type = "dashboard", destination = "dash-uuid", label = "Overview",
    open_in_new_tab = true, use_filters = true, use_time_range = false },
  { type = "external",  destination = "https://...", label = "Docs",
    open_in_new_tab = true, encode_url = false },
]
```
The two item types share `destination`, `label`, `open_in_new_tab`. Only `use_filters`/`use_time_range` (dashboard-only) and `encode_url` (external-only) diverge. Nested wrapper blocks would add indentation for negligible schema-level gain. Flat items stay closest to the API JSON shape.

**3. Terraform `type` enum shortens the API discriminator.**
`"dashboard"` maps to API `"dashboardLink"`; `"external"` maps to API `"externalLink"`. The `links = [...]` container supplies the noun; the model layer translates between representations.

**4. Validators for type-specific fields.**
- `type = "dashboard"`: reject `encode_url` at plan time.
- `type = "external"`: reject `use_filters` and `use_time_range` at plan time.
- `destination` and `type` are required on every item; `label` is optional.
- Panel-level validator: exactly one of `by_value` or `by_reference` (same `discoverSessionConfigModeValidator` pattern).

**5. Handler implements all `iface.Handler` methods.**
`AlignStateFromPlan`, `ClassifyJSON`, `PinnedHandler`, `PopulateJSONDefaults` all delegate to standard panelkit no-ops (matching existing simple panel handlers).

**6. No drilldowns block.**
The `links` panel API does not expose drilldowns; the `panelkit.URLDrilldownListAttribute` is not added.

## Risks / Trade-offs

- [Risk] API union discrimination (`by_value` vs `by_reference`) relies on field presence, not a discriminator field. Mitigation: the model layer inspects `ref_id` presence; tests cover both branches.
- [Risk] Link item type-specific validator rejects fields set to their zero value (false/null). Mitigation: only reject when explicitly set (non-null, non-unknown); match existing panelkit validator patterns.

## Open questions

_None — design decisions above were specified by the issue author (see human direction comment on issue #3999)._
