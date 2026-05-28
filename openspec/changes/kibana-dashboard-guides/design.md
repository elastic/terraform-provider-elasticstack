## Context

The `elasticstack_kibana_dashboard` resource was recently introduced. Its generated reference doc (`docs/resources/kibana_dashboard.md`) covers every field and panel type but is not approachable for first-time users. The project already has a guides pattern — `docs/guides/security-roles.md` and `docs/guides/elasticstack-kibana-rule.md` — that pairs prose with inline Terraform code blocks and demonstrates scenario-based use.

Guides live in `docs/guides/` as Markdown files and are registered in the provider's Terraform Registry template system via `make docs-generate`. Screenshots are static PNGs committed alongside the guide.

The dashboard resource's `vis` panels embed Lens visualization JSON, which is complex and not hand-authorable from scratch. The practical workflow is: create a visualization in the Kibana UI, use Inspect → Request to extract the serialized `attributes` JSON, then embed it in Terraform. Guides must document this workflow explicitly.

## Goals / Non-Goals

**Goals:**
- Three guides covering distinct experience levels: getting-started, operations patterns, advanced features
- All examples use Kibana's built-in sample data (logs + eCommerce) — no custom data pipeline required
- Each guide is self-contained and runnable with `terraform apply`
- Screenshots are reproducible via committed Playwright scripts
- Kibana 9.4+ baseline throughout

**Non-Goals:**
- Covering feature-gated panels (SLO, Synthetics, ML) — out of scope for general guides
- Geographic/Maps panels — geo-specific, narrow audience
- Modifying provider resource code beyond the targeted XY chart `fitting` alignment fix that this work surfaced; no resource schema changes
- Replacing the generated reference documentation

## Decisions

### Decision: By-value panels only in examples

**Choice**: All guide examples use `vis_config.by_value`, `markdown_config.by_value`, `discover_session_config.by_value` — no by-reference patterns.

**Why**: By-reference panels require pre-existing saved objects created outside Terraform, making examples non-runnable in isolation. By-value configs are self-contained and directly applicable. A brief callout in guide 1 mentions that by-reference exists and links to the reference docs.

**Alternative considered**: Show by-reference in guide 3 as an "advanced" pattern. Rejected — the overhead of explaining saved object management outweighs the benefit; by-reference patterns are sufficiently covered by the reference docs.

### Decision: Kibana sample data as the dataset

**Choice**: Guide 1 uses `kibana_sample_data_logs`; guide 2 uses `kibana_sample_data_ecommerce`; guide 3 uses `kibana_sample_data_logs`. Both datasets are installable with one click from the Kibana home page.

**Why**: Every Kibana instance ships with these datasets. Using them avoids any prerequisite data pipeline setup while letting each guide use the sample data best suited to its scenario. The field names and data shapes are stable across Kibana versions. Elastic's own tutorial content already uses these datasets, so users can cross-reference.

**Alternative considered**: Custom synthetic data via `elasticstack_elasticsearch_index` + bulk ingest. Rejected — significantly increases prerequisite complexity and makes guides longer without adding value to the dashboard-specific content.

### Decision: Lens visualization JSON sourced from Kibana UI export

**Choice**: Guide 1 includes an explicit section explaining how to create a Lens visualization in the Kibana UI and extract its JSON via Inspect → Request → Response, then embed it in `vis_config.by_value.*_chart_config`. The TF examples include complete, verbatim Lens JSON.

**Why**: Lens chart specs are not hand-authorable — they contain internal Kibana state fields. Users will always need to start in the UI. Making this workflow explicit prevents confusion about why the JSON is complex.

**Alternative considered**: Provide a minimal hand-crafted Lens spec. Rejected — minimal specs often break across Kibana minor versions as internal field requirements change. Full exported specs are stable.

### Decision: Playwright for screenshots, committed as static PNGs

**Choice**: Screenshot scripts under `scripts/screenshots/` using Playwright. Scripts write PNGs to `templates/guides/images/` (source of truth); `make docs-generate` copies them to `docs/guides/images/` for publication. Both directories are committed so `make check-docs` enforces the published copy stays in sync. Scripts reference dashboards by their Kibana-assigned `dashboard_id`.

**Why**: Playwright handles Kibana's JavaScript-heavy rendering and authentication. Scripts are committed so any contributor can regenerate screenshots when the UI or TF config changes. Static PNGs mean the guides work without running a live stack.

**Alternative considered**: Manual screenshots only (no automation). Rejected — manual screenshots are not reproducible and become stale as Kibana UI evolves.

### Decision: One spec for all three guides

**Choice**: Single capability `kibana-dashboard-guides` with one delta spec covering requirements for all three guides.

**Why**: The three guides are closely related — same resource, same output format, same screenshot mechanism. A single spec avoids duplication of shared requirements (guide format, screenshot reproducibility, `make docs-generate` integration) while keeping requirements organized by guide within the spec.

## Risks / Trade-offs

- **Lens JSON stability** → Kibana may change internal Lens field requirements between minor versions. Mitigation: pin the Kibana version in the prerequisites callout (9.4+); note in the guide that JSON was tested against a specific version and may need re-export for other versions.
- **Screenshot rendering timing** → Kibana panels load asynchronously; a screenshot taken too early shows spinners. Mitigation: Playwright scripts use `waitForSelector('[data-test-subj="embeddablePanel"]')` + `waitForLoadState('networkidle')` before capture, with a fallback timeout.
- **Sample data field name drift** → If Elastic changes the sample data schema, examples break. Mitigation: sample data schemas are stable across all Kibana 8.x/9.x versions; this risk is low.
- **Guide maintenance burden** → Guides with embedded screenshots and Lens JSON can become stale. Mitigation: the Playwright scripts lower the cost of regenerating screenshots; Lens JSON must be re-exported when the resource schema changes (which is infrequent for stable charts).

## Open Questions

- Should guide 3 cover `sections` with a tech-preview callout, or defer until sections GA? **Current plan**: include with a prominent tech-preview callout, as it is already supported by the resource schema.
- Should the Playwright scripts be wired into CI (e.g., run on PR with a Kibana service container)? **Current plan**: scripts are committed but not wired into CI for this change; CI integration is a follow-up.
