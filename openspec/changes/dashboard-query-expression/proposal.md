## Why

The Kibana Dashboard API's root `query` object has a single string field, `expression` (`kbn-as-code-query` in the pinned OpenAPI spec: `additionalProperties: false`, `required: [expression, language]`). The `elasticstack_kibana_dashboard` resource instead exposes `text` and `json`, both of which map onto that same `Expression` field on write (`dashboardQueryToAPI`, `internal/kibana/dashboard/models.go`), while read sniffs the returned string for a leading `{` to decide which of `text` / `json` to populate. `json` therefore has no real API counterpart: Kibana always treats the value as a KQL/Lucene query string, never as structured JSON.

This is reported in issue #5081: a dashboard exported from Kibana as JSON has `query.expression`, but the equivalent HCL needs `query.text`, so a practitioner (or a conversion tool) has to translate the field name. Every other query block in this resource — Lens chart queries (`lenscommon/filter_simple.go`), discover-session queries, and all the typed `vis` panel `query` blocks — already uses `expression`. The root `query` block is the one place that doesn't, and it is also the only place where a lossy heuristic (JSON-object sniffing) decides what the practitioner meant.

## What Changes

- Root `query` becomes `{ language (required), expression (required) }`; `text` and `json` are removed.
- **Breaking configuration change.** Existing `.tf` files using `query.text` or `query.json` must be updated to `query.expression`. This is intentional: the dashboard resource is Kibana "Technical Preview" (9.4+).
- **State migrates automatically.** The resource schema version moves from 1 to 2, and existing state (from v0 or v1) is upgraded in one step. No manual `terraform state` surgery is required.
- The JSON-object sniffing heuristic on read is removed; read and write become a direct one-to-one mapping of `language` and `expression`.
- Spec, tests, docs, examples and the CHANGELOG are updated to match. Implementation detail lives in `design.md`; the work breakdown in `tasks.md`.

### Release note (CHANGELOG draft)

```markdown
### Breaking changes

`elasticstack_kibana_dashboard`: the root `query` block now uses `expression` instead of `text` / `json`, matching the Kibana Dashboard API and the `expression` attribute already used by Lens chart queries. The API exposes a single string field, `expression`; `text` and `json` both mapped onto it, and `json` was only ever a JSON document sent as a query string. Replace `text` (or `json`) with `expression`:

    # Before
    query = {
      language = "kql"
      text     = "http.response.status_code:200"
    }

    # After
    query = {
      language   = "kql"
      expression = "http.response.status_code:200"
    }

A Plugin Framework state upgrader (schema v1 -> v2) automatically migrates existing state on the next `terraform apply`; no manual state surgery is required, but `.tf` files must be updated to use `expression`. Because `query.json` was always sent to Kibana as a string, its value is carried over to `expression` unchanged. ([#5081](https://github.com/elastic/terraform-provider-elasticstack/issues/5081))
```

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: root `query` attribute shape changes from `{ language, text?, json? }` to `{ language, expression }`; schema version 1 → 2 with an automatic state upgrade.

## Impact

- `internal/kibana/dashboard/schema.go` (query attribute definitions, schema `Version`).
- `internal/kibana/dashboard/models/dashboard.go` (`DashboardQueryModel`) and `internal/kibana/dashboard/models.go` (`dashboardQueryToAPI`, read-side query mapping).
- `internal/kibana/dashboard/state_upgrade.go` (new `migrateV1ToV2`, updated `UpgradeState` map).
- `openspec/specs/kibana-dashboard/spec.md` (REQ-036, REQ-007, REQ-008, REQ-040, new REQ-055, plus the schema sketch and the "no schema version / upgrader" note).
- Unit tests: `models_dashboard_root_test.go`, `models_optional_root_blocks_test.go`, `create_test.go`, `pinned_panels_mapping_test.go`, `state_upgrade_test.go`.
- Acceptance tests: `TestAccResourceDashboardRootQueryJSON` deleted; `TestAccResourceDashboardQueryTransition` rewritten; new upgrade-path acceptance test; ~20 `testdata/**/main.tf` files with `text =` renamed to `expression =`.
- `docs/resources/kibana_dashboard.md` (regenerated) and `examples/` and `templates/guides/` using `query.text` or the JSON-query example (`examples/resources/elasticstack_kibana_dashboard/resource.tf`).
- `CHANGELOG` breaking-change entry referencing issue #5081.

## Out of scope

- Other `query` shapes in the dashboard resource (Lens / discover-session / typed `vis` panel blocks already use `expression` and are unaffected).
- Any change to the generated Kibana client; the upstream OpenAPI schema already matches the target shape (`expression`/`language`), so no `generated/kbapi` regeneration is required.
- A future Kibana release reintroducing an object-valued query would be handled as a new attribute, not by overloading `expression` again.
