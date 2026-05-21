## Context

The `elasticstack_kibana_dashboard` resource includes a full implementation for `type = "lens-dashboard-app"` panels, added when the Kibana Dashboard OpenAPI spec mistakenly exposed this internal type publicly. Upstream Kibana has removed `lens-dashboard-app` from its public API spec. The provider's implementation is tied to a supplementary spec overlay (`generated/kbapi/dashboards.json`) that merges 386 dashboard-related schemas (including all panel types) into the main Kibana API spec for code generation purposes. Removing this overlay changes how the code generator names its output types, making the cleanup more involved than a simple file deletion.

The `vis` panel type is the correct public migration target for all `lens-dashboard-app` configurations. The existing unknown-panel fallback (`config_json`) handles any Kibana dashboards that still contain `lens-dashboard-app` panels at the API level after the provider is upgraded.

## Goals / Non-Goals

**Goals:**

- Remove `lens-dashboard-app` from the provider's public surface: schema block, Go handler package, model types, registry entry, constant, and examples.
- Remove the `dashboards.json` supplementary spec overlay and its merge/transform infrastructure from the kbapi generator, so the generated client derives all dashboard types from the upstream Kibana spec alone.
- Update `fixDashboardPanelItemRefs` in `transform_schema.go` to use the `Kibana_HTTP_APIs_` prefixed schema key names that appear in the upstream spec once the overlay is removed.
- Remove `fixVisualizationIdParam` (a transformer for a path that only existed in the overlay).
- Update all provider code that references renamed generated types after regeneration.
- Remove REQ-035 and all `lens-dashboard-app` references from the canonical requirements spec.
- Write an upgrade guide section for the breaking change.

**Non-Goals:**

- Converting existing Kibana saved dashboards that use `lens-dashboard-app` internally — that is a Kibana-side concern.
- Adding automated Terraform state migration (`StateUpgrader`) — practitioners must update their HCL configs manually.
- Removing shared Lens infrastructure (`lenscommon/`, `vis_config`) — those packages remain in use for `type = "vis"` panels.
- Adding any deprecation cycle — the dashboard resource is in technical preview and breaking changes are permitted without a deprecation release.

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|-------------------------|
| Removal strategy | Hard removal in a single PR with no deprecation cycle. The dashboard resource is in technical preview; `type = "vis"` is the complete 1:1 migration target. | Deprecation cycle (Approach B) was considered but rejected by @tobio: retains ~2 100 lines of dead code for at least one more release, unnecessarily conservative for a technical preview resource. |
| `dashboards.json` removal | Delete the file entirely; do not replace it with a newer version. There is no newer Kibana spec version that still includes `lens-dashboard-app`. | Keeping the file and marking `lens-dashboard-app` as removed within it was rejected as fragile and not aligned with the upstream intent. |
| `fixVisualizationIdParam` | Remove from the transformer pipeline and delete the function body. The `/api/visualizations/{id}` path was provided exclusively by `dashboards.json`; with the overlay removed, the function becomes a no-op at best. | Keeping it as a no-op was rejected: dead code in a critical code-generation pipeline is a maintenance hazard. |
| `fixDashboardPanelItemRefs` key names | Update to `Kibana_HTTP_APIs_` prefixed names after running `make -C generated/kbapi transform` to confirm the exact keys in the upstream spec. The function is not removed because the upstream spec still provides dashboard panel schemas under the new key names. | Removing the function was considered but rejected: panel item refs still need to be fixed for the upstream panel types. |
| Read-time safety net for existing `lens-dashboard-app` panels | The existing unknown-panel fallback (`config_json`) in `dashboardMapPanelFromAPI` covers read-time gracefully for Kibana dashboards that still have `lens-dashboard-app` at the API level. No new fallback code is needed. | Adding a dedicated read-only migration adapter was rejected as unnecessary complexity given the fallback already exists. |
| Test cleanup | Remove acceptance tests and unit tests that specifically cover `lens-dashboard-app` panels (in `acc_lens_dashboard_app_panels_test.go`, `acc_drilldowns_test.go`, and `lens_by_value_embed_wiring_test.go`). No replacement test is required for removed functionality. | Leaving tests in place was rejected: tests for a removed panel type will fail or require ongoing maintenance. |
| Upgrade guide | Add an upgrade guide section documenting the migration from `type = "lens-dashboard-app"` + `lens_dashboard_app_config` to `type = "vis"` + `vis_config`, including the `config_json` attribute relocation note for `by_value.config_json` users. | Silent removal was rejected: this is a breaking change and practitioners need an explicit migration path. |

## Risks / Trade-offs

- **Breaking change for existing configs**: any Terraform configuration using `type = "lens-dashboard-app"` will fail plan after upgrading. Mitigated by the upgrade guide and the clear migration path to `type = "vis"`.
- **Generated type renames**: removing `dashboards.json` causes other dashboard schema keys to adopt `Kibana_HTTP_APIs_` prefixes in the generated Go types (e.g., `KbnDashboardData` → `KibanaHTTPAPIsKbnDashboardData`). The exact set of renamed types is unknown until the generator runs. Implementers must run `make -C generated/kbapi transform` first, inspect `oas-filtered.yaml` for exact names, then update all provider-side references. An incomplete update will result in a compile error, which is a safe failure mode.
- **Transform code complexity**: `fixDashboardPanelItemRefs` operates on schema path strings. An incorrect update to the key names will cause the transformer to silently skip its work, potentially breaking panel item ref resolution at API request time. Mitigated by: (a) running the full `make -C generated/kbapi transform generate` pipeline, (b) inspecting `oas-filtered.yaml` for the actual dashboard schema keys after transformation, and (c) confirming that existing dashboard acceptance tests (for panel types other than `lens-dashboard-app`) pass after the change.

## Migration Plan

Practitioners who have `type = "lens-dashboard-app"` panels in their Terraform configurations must migrate to `type = "vis"` before or alongside upgrading to the provider version that includes this change.

**Attribute mapping:**
- `type = "lens-dashboard-app"` → `type = "vis"`
- `lens_dashboard_app_config.by_value.<chart>_config` → `vis_config.by_value.<chart>_config`
- `lens_dashboard_app_config.by_value.config_json` → panel-level `config_json` (not nested under `vis_config`)
- `lens_dashboard_app_config.by_reference.*` → `vis_config.by_reference.*`

**Example — typed by-value chart:**

```hcl
# Before:
panels = [{
  type = "lens-dashboard-app"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  lens_dashboard_app_config = {
    by_value = {
      metric_chart_config = { ... }
    }
  }
}]

# After:
panels = [{
  type = "vis"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  vis_config = {
    by_value = {
      metric_chart_config = { ... }
    }
  }
}]
```

**Example — `by_value.config_json`:**

```hcl
# Before:
panels = [{
  type = "lens-dashboard-app"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  lens_dashboard_app_config = {
    by_value = {
      config_json = jsonencode({ ... })
    }
  }
}]

# After (config_json moves to panel level):
panels = [{
  type = "vis"
  grid = { x = 0, y = 0, w = 12, h = 15 }
  config_json = jsonencode({ ... })
}]
```

Kibana dashboards that were previously saved with `lens-dashboard-app` panels are readable after the upgrade via the unknown-panel fallback (`config_json`). Practitioners should migrate those panels to `type = "vis"` in their HCL configuration and run `terraform apply` to reconcile state.

## Open Questions

None — all open questions are resolved. The implementation path is confirmed by @tobio in the issue thread.

*(Implementation note: run `make -C generated/kbapi transform` before writing any provider code. Inspect `oas-filtered.yaml` for: (a) exact `Kibana_HTTP_APIs_kbn-dashboard-*` schema key names used by `fixDashboardPanelItemRefs`; (b) whether the dashboard path is `/api/dashboards/{id}` or `/api/dashboards/dashboard/{id}` in the upstream spec; (c) whether all current panel types — vis, discover_session, esql_control, image, markdown, etc. — are present in the upstream spec under new names.)*
