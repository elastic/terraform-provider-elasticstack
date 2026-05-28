## 1. Infrastructure & Setup

- [x] 1.1 Create `scripts/screenshots/` directory with `package.json` declaring `playwright` dependency
- [x] 1.2 Write `scripts/screenshots/README.md` documenting prerequisites, environment variables, and example invocations
- [x] 1.3 Create `examples/guides/guide1-getting-started/`, `examples/guides/guide2-operations/`, `examples/guides/guide3-advanced/` directories
- [x] 1.4 Register the three guide templates in the provider's docs generation system (frontmatter + template file wiring) so `make docs-generate` picks them up

## 2. Guide 1 — Terraform Config & Screenshots

- [x] 2.1 Write `examples/guides/guide1-getting-started/main.tf` — dashboard shell with `kibana_sample_data_logs` data view, `time_range`, `refresh_interval`, `query`
- [x] 2.2 Add markdown panel to `main.tf` and verify it renders in Kibana
- [x] 2.3 Add two Lens Metric panels (count + cardinality) to `main.tf` and verify
- [x] 2.4 Add Lens Line chart panel to `main.tf` and verify
- [x] 2.5 Add Lens horizontal Bar chart panel (top-10 URLs) to `main.tf` and verify
- [x] 2.6 Add Lens Donut chart panel (response codes) to `main.tf` and verify
- [x] 2.7 Write `scripts/screenshots/guide1.mjs` — login, navigate, wait for panels, capture 7 PNGs to `templates/guides/images/`
- [x] 2.8 Run Playwright script and commit all `g1-*.png` screenshots to `templates/guides/images/`

## 3. Guide 2 — Terraform Config & Screenshots

- [x] 3.1 Write `examples/guides/guide2-operations/main.tf` — dashboard with `kibana_sample_data_ecommerce`, `options` block, `pinned_panels` with `options_list_control`
- [x] 3.2 Add three Lens Metric panels (revenue, order count, AOV) to `main.tf` and verify
- [x] 3.3 Add Lens Stacked Area chart panel to `main.tf` and verify
- [x] 3.4 Add Lens Data Table panel to `main.tf` and verify
- [x] 3.5 Add Lens Donut panel to `main.tf` and verify
- [x] 3.6 Add `discover_session` panel with `by_value.tab.dsl` to `main.tf` and verify
- [x] 3.7 Write `scripts/screenshots/guide2.mjs` — capture 4 PNGs including before/after filter screenshots
- [x] 3.8 Run Playwright script and commit all `g2-*.png` screenshots to `templates/guides/images/`

## 4. Guide 3 — Terraform Config & Screenshots

- [x] 4.1 Write `examples/guides/guide3-advanced/main.tf` — dashboard with `sections`, `image` panel, `tags` (`access_control` kept as a commented-out block with explanation — see task 7.7 for how the prose covers it)
- [x] 4.2 Add Lens Gauge panel with goal value to `main.tf` and verify
- [x] 4.3 Add Lens Heatmap panel (requests by hour × response code) to `main.tf` and verify
- [x] 4.4 Add multi-layer Lens Area+Line chart to `main.tf` and verify
- [x] 4.5 Add `esql_control` to `pinned_panels` and a panel query referencing its variable — verify filtering works
- [x] 4.6 Wire panels into `sections` blocks and verify collapse/expand in Kibana
- [x] 4.7 Write `scripts/screenshots/guide3.mjs` — capture 4 PNGs including collapsed/expanded section states
- [x] 4.8 Run Playwright script and commit all `g3-*.png` screenshots to `templates/guides/images/`

## 5. Guide 1 — Prose

- [x] 5.1 Write prerequisites section: provider setup, Kibana 9.4+ requirement, sample data installation steps
- [x] 5.2 Write dashboard shell section: explain `time_range`, `refresh_interval`, `query`; introduce the 48-column grid with a visual diagram
- [x] 5.3 Write "How to get Lens visualization JSON" section: Kibana UI → Inspect → Request → copy attributes
- [x] 5.4 Write markdown panel step with snippet and `g1-02-markdown.png`
- [x] 5.5 Write Metric panels step (count + cardinality) with snippet and `g1-04-metric2.png`
- [x] 5.6 Write Line chart step with snippet and `g1-05-line.png`
- [x] 5.7 Write Bar chart step with snippet and `g1-06-bar.png`
- [x] 5.8 Write Donut chart step with snippet and `g1-07-final.png` (final dashboard)
- [x] 5.9 Write "Next steps" section linking to guides 2 and 3 and the reference docs

## 6. Guide 2 — Prose

- [ ] 6.1 Write prerequisites section and overview of what the dashboard will show
- [ ] 6.2 Write `pinned_panels` and `options_list_control` section — explain controls vs. content panels, show before/after screenshots `g2-01-full.png` and `g2-02-filtered.png`
- [ ] 6.3 Write KPI row section (three Metric panels) with snippet
- [ ] 6.4 Write Stacked Area chart section with snippet
- [ ] 6.5 Write Data Table section with `g2-04-table.png`
- [ ] 6.6 Write `discover_session` section — explain `ref_id`, `column_order`, `view_mode`; include `g2-03-discover.png`
- [ ] 6.7 Write dashboard `options` section explaining `use_margins`, `sync_colors`, `sync_tooltips`

## 7. Guide 3 — Prose

- [ ] 7.1 Write prerequisites section and overview of advanced features covered
- [ ] 7.2 Write `sections` section with tech-preview callout; include `g3-01-full.png` and `g3-02-collapsed.png`
- [ ] 7.3 Write `image` panel section (branding use-case)
- [ ] 7.4 Write Gauge chart section with goal value explanation; include `g3-03-gauge.png`
- [ ] 7.5 Write Heatmap chart section; include `g3-04-heatmap.png`
- [ ] 7.6 Write `esql_control` section — explain variable declaration and `?variable` query syntax
- [ ] 7.7 Write `access_control` section — explain `write_restricted`, replacement-on-change behavior
- [ ] 7.8 Write `tags` section — explain tag IDs and dashboard organisation

## 8. Validation & Polish

- [ ] 8.1 Run `terraform validate` in all three example directories; fix any schema errors
- [ ] 8.2 Run `make docs-generate` and verify all three guides appear in `docs/guides/`
- [ ] 8.3 Run `make check-docs` and fix any broken references or missing example files
- [ ] 8.4 Review all screenshots — no loading spinners, no empty panels; re-run scripts if needed
- [ ] 8.5 Verify all Terraform code blocks in the guides match the final `main.tf` configs
- [ ] 8.6 Check frontmatter in all three guides matches the format of existing guides
- [ ] 8.7 Verify `docs/resources/kibana_dashboard.md` includes "See also" links to the three guides
