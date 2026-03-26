## Context

The `elasticstack_kibana_dashboard` resource supports a growing list of typed panel config blocks (e.g. `markdown_config`, `xy_chart_config`, `metric_chart_config`). Each typed block is valid for a specific `type` value and is mutually exclusive with all other typed config blocks on a given panel entry. The `slo_error_budget` panel type is not a Lens visualization: it carries its own standalone inline config (SLO reference, optional instance selection, optional drilldowns, and display preferences) rather than a Lens expression or `config_json` blob.

The Kibana dashboard API exposes SLO error budget panels through the `slo-error-budget-embeddable` schema. The panel shares structural similarities with the `slo_burn_rate` (REQ-032) and `slo_overview` panels: all three reference an SLO by `slo_id`, optionally scope to a specific instance with `slo_instance_id`, and support the same `drilldowns` array shape.

## Goals / Non-Goals

**Goals:**

- Introduce a typed `slo_error_budget_config` block that lets practitioners declare all attributes of an SLO error budget panel in HCL, with `slo_id` enforced as required at plan time.
- Model `slo_instance_id` as optional; preserve Terraform's null in state when the practitioner did not configure it, rather than letting the API's default value of `"*"` bleed into state and cause spurious drift.
- Map `drilldowns` as a list of typed objects sharing the same drilldown model used by `slo_overview` and `slo_burn_rate` panels, with `url`, `label`, `trigger`, and `type` required, and `encode_url` (default `true`) and `open_in_new_tab` (default `true`) optional.
- Add schema-level validation that `slo_error_budget_config` is only used with `type = "slo_error_budget"` and conflicts with all other typed config blocks and `config_json`.
- Cover the full panel lifecycle with acceptance tests and the converter logic with unit tests.

**Non-Goals:**

- Supporting `slo_error_budget` panels through the raw `config_json` path (that path remains available for unsupported types; this proposal adds a typed alternative only).
- Managing the referenced SLO lifecycle (the `slo_id` is a reference attribute, not a managed dependency).
- Extending the `config_json` type allowlist to include `slo_error_budget`.

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|-------------------------|
| Config block name | `slo_error_budget_config`, following the `<panel_type>_config` naming convention used by all other typed blocks. | Shorter names were rejected as less discoverable and inconsistent with the convention. |
| Required fields | `slo_id` is required inside the config block; all other attributes are optional. | Making `slo_id` optional was rejected because the API schema marks it required and omitting it produces an invalid panel. |
| `slo_instance_id` default handling | The attribute is optional in the Terraform schema. The provider SHALL NOT write the API default `"*"` into state when the practitioner omits `slo_instance_id`. On read-back, if `slo_instance_id` was null in prior state/plan, the provider SHALL preserve null rather than replacing it with the API-returned `"*"`. | Writing `"*"` unconditionally into state was rejected because it would produce spurious drift for practitioners who leave the field unset, which is the common case for single-instance SLOs. |
| `drilldowns` representation | A list of typed objects with `url` (string, required), `label` (string, required), `trigger` (string, required, const `"on_open_panel_menu"`), `type` (string, required, const `"url_drilldown"`), `encode_url` (bool, optional, default `true`), `open_in_new_tab` (bool, optional, default `true`). | A `drilldowns_json` string blob was considered but rejected in favour of typed HCL for practitioner usability and plan-time validation. |
| Shared drilldown model | The drilldown object shape is identical across `slo_overview`, `slo_burn_rate`, and `slo_error_budget` panels. The implementation SHOULD reuse a shared drilldown converter function to avoid duplication. | Separate per-panel drilldown converters were rejected as unnecessarily duplicating identical logic. |
| `trigger` and `type` validation | Enum validation at schema level: `trigger` must be `"on_open_panel_menu"`, `type` must be `"url_drilldown"`. | Free-form strings were rejected to catch typos at plan time. |
| `encode_url` / `open_in_new_tab` default handling | Both are optional booleans. The provider SHALL normalize the API defaults (`true` for each) on read so that omitting them in configuration does not cause drift when Kibana returns `true`. | Treating these as purely optional with no default normalization was rejected because Kibana always materializes the default values on read-back. |
| Panel type exclusivity | `slo_error_budget_config` SHALL conflict with all other typed config blocks and `config_json`; enforced by schema validators consistent with existing typed blocks. | A runtime-only check was rejected; plan-time validation gives earlier, cleaner error messages. |

## Risks / Trade-offs

- [Risk] Kibana may emit additional optional fields in the `slo-error-budget-embeddable` config on read-back that are not yet modelled. -> Mitigation: the converter ignores unmapped fields on read, consistent with the pattern used by other typed panel converters.
- [Risk] The `slo_instance_id` null-preservation logic must be applied consistently in the read path; if the prior-state seed is not applied, a plan-refresh after initial apply would show a diff for practitioners who left the field unset. -> Mitigation: the converter MUST check prior state/plan before deciding whether to write the API-returned value, following the same pattern as other fields with API-injected defaults in the dashboard resource.
- [Risk] The shared drilldown model introduces a coupling point: if `slo_burn_rate` or `slo_overview` panels evolve their drilldown schema, the shared converter may need updating for all three panels simultaneously. -> Mitigation: document the shared dependency; any extension to the drilldown model should be applied uniformly across all SLO panel types.

## Migration Plan

- No user-facing migration is required. Existing dashboards that include `slo_error_budget` panels managed through `config_json` continue to work unchanged.
- Practitioners who wish to migrate from `config_json` to the typed block must replace `config_json` with `slo_error_budget_config` in their configuration; the underlying Kibana object is unchanged and no resource replacement is triggered.
- No schema version bump or state upgrade is needed.

## Open Questions

- None blocking the proposal. The exact `slo_instance_id` read-back behavior (whether Kibana always returns `"*"` or omits the field when no instance is configured) should be confirmed against a real Kibana instance during acceptance testing, to validate that the null-preservation strategy is sufficient.
