## Why

The `elasticstack_kibana_dashboard` resource has extensive generated reference documentation but no use-case-focused guides. Users must read several hundred lines of API reference to accomplish common tasks — issue [#1722](https://github.com/elastic/terraform-provider-elasticstack/issues/1722) specifically requests simplified guides that solve everyday problems concisely, with screenshots of the resulting dashboards.

## What Changes

- **New guide**: `docs/guides/kibana-dashboard-getting-started.md` — step-by-step construction of a web server log monitoring dashboard using Kibana sample data, introducing every core concept one panel at a time
- **New guide**: `docs/guides/kibana-dashboard-operations.md` — an interactive eCommerce monitoring dashboard demonstrating controls panels, a Discover session, and dashboard-level options
- **New guide**: `docs/guides/kibana-dashboard-advanced.md` — advanced patterns covering collapsible sections, ES|QL controls, image panels, gauge and heatmap chart types, and access control
- **New example configs**: `examples/guides/guide1-getting-started/main.tf`, `examples/guides/guide2-operations/main.tf`, `examples/guides/guide3-advanced/main.tf` — fully runnable Terraform configurations for each guide
- **New Playwright scripts**: `scripts/screenshots/guide{1,2,3}.mjs` — reproducible screenshot scripts for capturing dashboard images from a local Kibana instance
- **New screenshots**: `templates/guides/images/g{1,2,3}-*.png` (copied to `docs/guides/images/` by `make docs-generate`) — embedded screenshots across the three guides
- **Provider fixes surfaced while authoring the guides**:
  - Guide 1: align `vis_config.by_value.xy_chart_config.fitting` plan state when Kibana omits the block (e.g. `bar_horizontal` with terms).
  - Guide 3 — panel read and alignment: stop shallow-copying plan into read state; clone plan-side typed `*_config` blocks when Kibana leaves them nil (`alignPlanTypedPanelConfigsWhenStateNil`); preserve nested section panels (`image_config`, `xy_chart_config.legend`, `heatmap_config.legend`, gauge `styling.shape_json`); keep top-level `panels = []` when explicitly set (omit vs empty list).
  - Guide 3 — Kibana API routing: route `metric_chart_config` with `type = "esql"` to the ES|QL API variant; set `type` on ES|QL control panels (including pinned controls).
  - Guide 3 — operator note: `access_control = "write_restricted"` requires a non-bootstrap Kibana user (the committed example leaves this block commented out).

Apart from the targeted dashboard alignment fixes above, no other changes to provider resource code or existing documentation.

## Capabilities

### New Capabilities

- `kibana-dashboard-guides`: Three progressive guides covering the `elasticstack_kibana_dashboard` resource from first panel to advanced features, with runnable Terraform examples, embedded screenshots, and Playwright automation scripts for screenshot generation. Requires Kibana 9.4+.

### Modified Capabilities

<!-- No existing requirement specs are changing. -->

## Impact

- **Mostly new files** — three new guide templates and examples; generated docs change because templates render into `docs/guides/`
- **Targeted `elasticstack_kibana_dashboard` provider fixes** (no schema changes): alignment in `lensxy`, `lensgauge`, `lensheatmap`, `lensmetric`, `image`, `esqlcontrol`, and shared panel read/plan-state flow (`models_panels.go`, `models_plan_state_alignment.go`); unit tests per area; CHANGELOG entries under Unreleased
- **Guide 3 constraint**: `access_control = "write_restricted"` needs a Kibana user with a profile ID (not the default `elastic` bootstrap user)
- **Kibana version baseline**: Kibana 9.4+ (minimum required by the dashboard API and resource)
- **Dependencies**: Kibana sample data (logs + eCommerce datasets, installable from Kibana home); Playwright for screenshot generation
- **Guide format**: Follows the established pattern of existing guides (`security-roles.md`, `elasticstack-kibana-rule.md`) — frontmatter, prerequisites, prose with inline Terraform code blocks, embedded images
