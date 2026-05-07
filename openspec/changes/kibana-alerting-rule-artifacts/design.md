## Context

The Kibana alerting rule API accepts an `artifacts` object in the create (`POST /api/alerting/rule/{id}`) and update (`PUT /api/alerting/rule/{id}`) request bodies, and returns it in the GET response. The generated `kbapi` client already models this in `AlertingRuleAPIBodyGeneric.Artifacts` (dashboards list + investigation_guide.blob) and in `GetAlertingRuleIdResponse.JSON200.Artifacts`. Neither `internal/models.AlertingRule` nor the Terraform resource schema currently surface these fields.

## Goals

- Expose `artifacts` on `elasticstack_kibana_alerting_rule` so practitioners can link Kibana dashboards and attach investigation guides to rules as code.
- Support **inline content** (`content`) and **file-path-based content** (`content_path`) for the investigation guide; the latter uses a SHA-256 checksum (computed at plan time) to detect when the source file changes outside Terraform.
- Map `artifacts` correctly on create, update, and read; when `artifacts` is absent from the Terraform config, omit it from the PUT body so Kibana does not wipe existing rule artifacts.

## Non-Goals

- Changing Kibana alerting rule behavior beyond what the API's `artifacts` object controls.
- Adding new Kibana API behavior (this tracks existing API semantics).
- Supporting dashboard creation from within this resource.

## Decisions

| Topic | Decision |
|-------|-----------|
| Schema shape | `artifacts` as a `schema.SingleNestedBlock` (needs nested blocks for the `dashboards` list). `dashboards` as `schema.ListNestedBlock`. `investigation_guide` as `schema.SingleNestedBlock`. |
| Dashboards | Each `dashboards {}` entry holds a single required `id` string attribute (Kibana dashboard saved-object id). Maps 1-to-1 to `artifacts.dashboards[].id` in the API. |
| Investigation guide — inline | `content` (optional string): passed directly to API as `blob`. On read, API-returned `blob` is stored back as `content` in state. |
| Investigation guide — file-based | `content_path` (optional string): provider reads the file at plan time, computes SHA-256, and at apply time sends file content as `blob`. `checksum` (computed string) is written to state after create/update. On the next plan the provider re-reads the file and compares the new checksum to state; if different, computed fields are marked unknown (same pattern as `elasticstack_fleet_custom_integration`). |
| Mutual exclusion | When `investigation_guide` is present, exactly one of `content` or `content_path` MUST be set. Enforced via `objectvalidator.ExactlyOneOf` or equivalent plan-time validation. |
| `checksum` in state | Only meaningful when `content_path` is used. Written to state at apply time (create or update). On read, `checksum` is not overwritten from the API (the API returns `blob`, not a checksum); the plan modifier manages drift. |
| Update when absent | When `artifacts` is not in the Terraform config, the provider SHALL omit `artifacts` from the PUT body so Kibana leaves existing rule artifacts unchanged. |
| Read path — content vs content_path | If prior state has `content` set (and `content_path` null): populate `content` from API `blob` after read. If prior state has `content_path` set (and `content` null): do not update `content` or `content_path` from the API; the plan modifier handles drift detection via `checksum`. |
| Plan modifier | A `ModifyPlan` hook on the resource (mirroring `elasticstack_fleet_custom_integration`) reads `content_path` during plan, computes SHA-256, and if it differs from the stored `checksum`, marks `checksum` (and potentially `id`) as unknown so Terraform shows a non-empty plan. |

## Non-Goals (implementation)

- Do not add `content_path` support to the `investigation_guide.content` attribute; keep them as separate, mutually exclusive attributes.

## Risks / Trade-offs

- **Removing `artifacts` from Terraform config** still results in an update that omits `artifacts`; Kibana keeps any previously stored artifacts. After refresh, state may again show `artifacts` from the API while configuration omits the block, producing a **plan diff** until the practitioner aligns config with the API or clears rule artifacts outside Terraform. Document in resource docs.
- **External file changes**: When `content_path` is used, an external change to the file (without running `terraform plan`) will not be detected until the next `terraform plan` (at which point the plan modifier detects the checksum mismatch). This is consistent with the custom integration behavior and acceptable.
- **Content drift from Kibana**: Kibana may normalise the blob (e.g. trim whitespace). If so, `content`-based state may show perpetual drift. If this is observed during testing, consider trimming both sides before comparison or treating blob as opaque (read-only after first write). This should be investigated at implementation time.

## Open Questions

1. **Minimum Kibana version for `artifacts` on alerting rules**: Not confirmed from release notes. Implementation MUST check whether the field is rejected on older stack versions and add an appropriate version gate (e.g. `>= 8.16.0` or `>= 9.x`) with a diagnostic if confirmed. If no minimum version can be established, rely on the API to reject the request on unsupported versions.
2. **Blob normalisation**: Does Kibana modify the investigation guide blob (e.g. trim whitespace, normalise line endings) before storing? If so, read-path comparison for `content` may produce spurious drift. Investigate during acceptance testing.

## Migration / State

- `artifacts` is a new optional block; no state upgrade is required. Existing resources without `artifacts` configured will have it absent from state after upgrade; if Kibana returns artifacts for those rules, a drift plan will appear (expected behaviour — practitioner must configure or ignore).
