# Kibana Dashboard Guides

Guide implementation: `templates/guides/kibana-dashboard-getting-started.md.tmpl`, `templates/guides/kibana-dashboard-operations.md.tmpl`, `templates/guides/kibana-dashboard-advanced.md.tmpl`, rendered to `docs/guides/kibana-dashboard-getting-started.md`, `docs/guides/kibana-dashboard-operations.md`, and `docs/guides/kibana-dashboard-advanced.md`

## ADDED Requirements

### Requirement: Three guide templates exist, render to guide files, and are linked from the dashboard resource docs

Three standalone provider guide templates SHALL exist as sources consumed by `make docs-generate`:
- `templates/guides/kibana-dashboard-getting-started.md.tmpl`
- `templates/guides/kibana-dashboard-operations.md.tmpl`
- `templates/guides/kibana-dashboard-advanced.md.tmpl`

Running `make docs-generate` SHALL render those templates to these guide files:
- `docs/guides/kibana-dashboard-getting-started.md`
- `docs/guides/kibana-dashboard-operations.md`
- `docs/guides/kibana-dashboard-advanced.md`

The `elasticstack_kibana_dashboard` resource documentation SHALL include a "See also" section linking to all three guides.

Each guide template SHALL declare `subcategory: ""` and an appropriate `page_title` and `description` in its frontmatter so the rendered guide preserves those values.

#### Scenario: Guides render without error

- **WHEN** the three template files under `templates/guides/` are present and `make docs-generate` is run
- **THEN** all three guide files are created under `docs/guides/` with no generation errors

#### Scenario: make check-docs passes

- **WHEN** all three guide template files, the rendered guide files, and their referenced example files are committed
- **THEN** `make check-docs` exits with code 0

---

### Requirement: Guide 1 — Getting started guide builds a dashboard step by step

Guide 1 (`kibana-dashboard-getting-started.md`) SHALL demonstrate building a web server log monitoring dashboard using `kibana_sample_data_logs`, adding panels incrementally so each step is independently reproducible.

The guide SHALL cover these topics in order:
1. Prerequisites — provider setup, Kibana 9.4+ version requirement, sample data installation
2. Dashboard shell — `title`, `time_range`, `refresh_interval`, `query`, and the grid coordinate system
3. Markdown panel — `markdown_config.by_value` with `content`, `title`, and `settings.open_links_in_new_tab`
4. Lens Metric panels — `vis_config.by_value.metric_chart_config` for count and cardinality metrics
5. Lens Line chart — `vis_config.by_value.xy_chart_config` for time-series request volume
6. Lens Bar chart (horizontal) — `vis_config.by_value.xy_chart_config` for top-N URL breakdown
7. Lens Donut chart — `vis_config.by_value.pie_chart_config` for response code distribution

The guide SHALL include an explicit section explaining how to obtain Lens visualization JSON from the Kibana UI using Inspect → Request → Response.

Each panel addition step SHALL include:
- A complete Terraform snippet showing the full `panels` list up to that step
- An embedded screenshot of the dashboard after applying that step

#### Scenario: Guide 1 Terraform config applies cleanly

- **WHEN** `terraform apply` is run with `examples/guides/guide1-getting-started/main.tf`
- **THEN** the apply completes without errors
- **THEN** the dashboard is visible in Kibana at `http://localhost:5601` with all 6 panels rendered

#### Scenario: Guide 1 covers at least 4 distinct Lens chart subtypes

- **WHEN** the guide is read
- **THEN** it demonstrates at least 4 distinct Lens visualization types: metric, line/area, bar, and pie/donut

#### Scenario: Guide 1 explains the grid coordinate system

- **WHEN** the grid section is read
- **THEN** it explains that Kibana uses a 48-column grid and defines `x`, `y`, `w`, `h` fields

---

### Requirement: Guide 2 — Operations guide demonstrates interactive controls and Discover sessions

Guide 2 (`kibana-dashboard-operations.md`) SHALL demonstrate an interactive eCommerce monitoring dashboard using `kibana_sample_data_ecommerce` with controls and an embedded Discover session.

The guide SHALL cover:
- `pinned_panels` with `options_list_control` wired to a data field, filtering all panels simultaneously
- At least 3 Lens Metric panels forming a KPI row
- Lens Stacked Area chart for trend analysis
- Lens Data Table panel
- `discover_session_config.by_value` with `tab.dsl`, including `data_view_reference` via `ref_id`, `column_order`, and `view_mode`
- Dashboard `options` block: `use_margins`, `sync_colors`, `sync_tooltips`

The guide SHALL include a before/after screenshot pair showing the dashboard with no filter applied and with an options_list_control filter active.

#### Scenario: Guide 2 Terraform config applies cleanly

- **WHEN** `terraform apply` is run with `examples/guides/guide2-operations/main.tf`
- **THEN** the apply completes without errors
- **THEN** the dashboard is visible in Kibana with controls panel, KPI row, area chart, data table, and Discover session

#### Scenario: Guide 2 demonstrates pinned_panels

- **WHEN** the guide is read
- **THEN** it includes a Terraform snippet using `pinned_panels` with at least one control panel
- **THEN** the guide explains the difference between `pinned_panels` (controls) and `panels` (content panels)

#### Scenario: Guide 2 demonstrates discover_session panel

- **WHEN** the guide is read
- **THEN** it includes a `discover_session` panel using `by_value.tab.dsl` with `data_view_reference`
- **THEN** it explains the `ref_id` field and how it links to a Kibana data view

---

### Requirement: Guide 3 — Advanced guide covers sections, ES|QL controls, and production patterns

Guide 3 (`kibana-dashboard-advanced.md`) SHALL demonstrate advanced dashboard features targeting users building production-grade dashboards.

The guide SHALL cover:
- Collapsible `sections` grouping panels, with a tech-preview callout
- `image_config.by_value` with `url` source for branding
- Lens Gauge chart (`vis_config.by_value.gauge_chart_config`) with a goal value
- Lens Heatmap chart (`vis_config.by_value.heatmap_chart_config`) showing activity by day-of-week and hour
- `esql_control` in `pinned_panels` with a named variable, and a panel query referencing that variable
- `access_control` with `access_mode = "write_restricted"` and its replacement-on-change implication
- `tags` field for dashboard organisation

The guide SHALL include a screenshot showing one section collapsed and another expanded.

#### Scenario: Guide 3 Terraform config applies cleanly

- **WHEN** `terraform apply` is run with `examples/guides/guide3-advanced/main.tf`
- **THEN** the apply completes without errors
- **THEN** the dashboard is visible in Kibana with collapsible sections, a gauge, a heatmap, and an image panel

#### Scenario: Guide 3 covers sections with a tech-preview callout

- **WHEN** the sections portion of the guide is read
- **THEN** it includes a callout stating that `sections` is a Kibana technical preview feature
- **THEN** it links to the Kibana release notes or docs for current feature status

#### Scenario: Guide 3 covers access_control replacement behavior

- **WHEN** the access_control portion of the guide is read
- **THEN** it notes that changing `access_mode` forces replacement of the dashboard resource

---

### Requirement: Runnable example configs exist for all three guides

Each guide SHALL have a corresponding self-contained Terraform configuration under `examples/guides/`:
- `examples/guides/guide1-getting-started/main.tf`
- `examples/guides/guide2-operations/main.tf`
- `examples/guides/guide3-advanced/main.tf`

Each config SHALL include a provider block, all required resource arguments, and use `kibana_sample_data_logs` or `kibana_sample_data_ecommerce` as the data source.

All `vis_config.by_value.*_chart_config` blocks SHALL reference `kibana_sample_data_logs` or `kibana_sample_data_ecommerce` index patterns with correct field names for those datasets.

#### Scenario: Example configs are syntactically valid Terraform

- **WHEN** `terraform validate` is run in each example directory
- **THEN** validation succeeds with no errors

---

### Requirement: Playwright screenshot scripts are committed and documented

A `scripts/screenshots/` directory SHALL exist containing:
- `package.json` declaring `playwright` as a dependency
- `README.md` explaining prerequisites and how to run the scripts
- `guide1.mjs`, `guide2.mjs`, `guide3.mjs` — one script per guide

Each script SHALL:
- Accept `KIBANA_URL`, `KIBANA_USER`, and `KIBANA_PASS` environment variables with sensible defaults
- Handle the Kibana login flow before navigating to the dashboard
- Wait for panel rendering to complete (`networkidle` + panel selector) before capturing
- Write output PNGs to `docs/guides/images/`

Screenshot PNG files SHALL be committed to the repository at `docs/guides/images/`.

#### Scenario: Screenshot scripts produce output PNGs

- **WHEN** a local Kibana 9.4+ instance is running with sample data installed and the guide configs applied
- **WHEN** `node scripts/screenshots/guide1.mjs` is run
- **THEN** PNG files are written to `docs/guides/images/` with no errors

#### Scenario: README documents the full screenshot workflow

- **WHEN** `scripts/screenshots/README.md` is read
- **THEN** it lists the prerequisites: running Kibana, sample data installed, terraform applied, `npm install`
- **THEN** it documents the environment variables and an example invocation
