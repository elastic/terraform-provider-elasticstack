# Proposal: Remove `lens-dashboard-app` Panel Type from `elasticstack_kibana_dashboard`

## Why

The Kibana Dashboard API included `lens-dashboard-app` as a panel type in its public OpenAPI spec by mistake. Upstream Kibana has since removed it — the type is intended to be Kibana-internal only and was never meant to be exposed as a public panel type. The provider currently carries a full implementation for `type = "lens-dashboard-app"` panels (the `lensdashboardapp` Go package, the `lens_dashboard_app_config` schema block, and the `kbn-dashboard-panel-type-lens-dashboard-app` schema in the supplementary `generated/kbapi/dashboards.json` file). Retaining this implementation diverges from the upstream API contract, creates maintenance surface for a dead code path, and makes the provider spec inconsistent with Kibana's published API. The `vis` panel type is a complete, 1:1 migration target for all configurations that used `lens-dashboard-app`. The dashboard resource is in technical preview, so breaking changes are expected and documented.

## What Changes

- **Remove `dashboards.json` and its schema-merge infrastructure** from the kbapi generator: delete `generated/kbapi/dashboards.json`, remove `mergeDashboardsSchema` from the transformer pipeline in `generated/kbapi/transform_schema.go`, add a paths-only `generated/kbapi/dashboard-paths.json` overlay injected via `injectDashboardAPIPaths`, retain `fixVisualizationIdParam` (upstream redirect stubs still lack `{id}` parameters), and update `fixDashboardPanelItemRefs` to use the upstream-native `Kibana_HTTP_APIs_` schema key prefix.
- **Regenerate `generated/kbapi/kibana.gen.go`** so that `KbnDashboardPanelTypeLensDashboardApp` and any other dashboard types renamed by the upstream key-prefix change are emitted correctly by the generator.
- **Delete `internal/kibana/dashboard/panel/lensdashboardapp/`** — the full panel handler implementation.
- **Remove `lens_dashboard_app_config`** from the panel schema (`schema.go`), from `PanelModel` (`models/panel.go`), and from the model types in `models/lens.go`.
- **Deregister `lensdashboardapp.Handler{}`** from `registry.go` and remove its `panelTypeAliases` entry.
- **Update provider references** to any generated Go types that are renamed after regeneration (e.g., `KbnDashboardData` → `KibanaHTTPAPIsKbnDashboardData`).
- **Remove REQ-035** from `openspec/specs/kibana-dashboard/spec.md` and all remaining references to `lens-dashboard-app` in that file.
- **Remove lens-dashboard-app examples** from `examples/resources/elasticstack_kibana_dashboard/resource.tf`.
- **Write an upgrade guide section** directing practitioners to migrate `type = "lens-dashboard-app"` to `type = "vis"`.

## Capabilities

After this change:

- The provider schema no longer accepts `type = "lens-dashboard-app"` panels or the `lens_dashboard_app_config` block. Existing Terraform configurations that use this type will fail plan until migrated to `type = "vis"`.
- Kibana dashboards that were previously saved with `lens-dashboard-app` panels at the API level are handled gracefully at read time by the existing unknown-panel fallback (`config_json`). No data is lost; Terraform will surface those panels as `config_json`-typed unknowns, which practitioners can migrate to `vis` panels.
- The kbapi generator no longer requires the supplementary `dashboards.json` overlay; the generated client derives dashboard types exclusively from the upstream Kibana spec.

## Implementation Notes

- **Run `make -C generated/kbapi transform` before writing provider code.** The transform step produces `oas-filtered.yaml`, which reveals the exact post-removal key names under `schemas.Kibana_HTTP_APIs_kbn-dashboard-*`. These names must be confirmed before updating `fixDashboardPanelItemRefs` and any provider-side references to generated structs.
- **`fixVisualizationIdParam` is retained**: upstream `oas.yaml` still exposes `/api/visualizations/{id}` as redirect stubs without `{id}` path parameters; the transformer injects them so `oapi-codegen` succeeds. Dashboard HTTP routes use a separate paths-only `dashboard-paths.json` overlay (not schema merging).
- **Unknown-panel fallback is the safety net**: the read path for unrecognized panel types already falls back to `config_json`. No new fallback code is needed.
- **Shared Lens infrastructure stays**: `lenscommon/` and `vis_config` packages remain in use for `type = "vis"` panels and are not in scope for removal.

## Impact

- **Breaking change**: any Terraform configuration that uses `type = "lens-dashboard-app"` will receive a plan error after upgrading. The upgrade guide provides the migration path.
- **kbapi layer**: removing `dashboards.json` causes other dashboard schema keys to adopt `Kibana_HTTP_APIs_` prefixes in the generated client. Provider code that references renamed types (e.g., `KbnDashboardData`, `KbnDashboardSection`) must be updated. The exact set of affected call sites is determined after running the generator.
- **Test cleanup**: acceptance tests and unit tests for `lens-dashboard-app` panels are removed; no replacement test is needed for removed functionality.
- **Spec**: REQ-035 and all `lens-dashboard-app` references are removed from the canonical requirements.
