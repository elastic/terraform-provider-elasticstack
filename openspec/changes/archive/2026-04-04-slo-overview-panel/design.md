## Context

The Kibana dashboard API exposes SLO overview panels through two distinct embeddable types: `slo-single-overview-embeddable` (single mode, one SLO by id) and `slo-group-overview-embeddable` (groups mode, aggregated SLO view). The two variants share common display properties (`title`, `description`, `hide_title`, `hide_border`) and a `drilldowns` array, but differ in their identifying and filtering fields. The correct API embeddable type is determined at write time by the value of the `overview_mode` discriminant field (`"single"` or `"groups"`).

The existing dashboard provider already has a pattern for mutually exclusive sub-blocks within a single typed config block: `datatable_config` carries a `no_esql` and an `esql` sub-block, and schema validators enforce that exactly one is present. The `slo_overview_config` design follows this same pattern.

The `group_filters.filters` field in the groups-mode API schema accepts an array of AS-code filter objects (conditions, groups, DSL queries, spatial filters). This schema is deeply polymorphic and difficult to represent ergonomically as typed Terraform attributes. A `filters_json` escape-hatch attribute is used instead.

## Goals / Non-Goals

**Goals:**

- Support full lifecycle management of `slo_overview` panels in both single and groups modes.
- Expose all documented fields for each mode as typed schema attributes where feasible.
- Provide validated enum constraints for `group_filters.group_by`.
- Model `drilldowns` as a typed list of objects with required and optional fields.
- Preserve `slo_instance_id = null` in state when not configured by the practitioner, even though the API default is `"*"`.
- Follow the same sub-block mutual exclusion pattern as `datatable_config`.

**Non-Goals:**

- Supporting `slo_overview` through `config_json`; the complexity of the two-mode discriminated union is better served by the typed block.
- Typed representation of the complex AS-code `group_filters.filters` schema; `filters_json` is used instead.
- Supporting read-back drift normalization for `drilldowns` beyond what the API naturally round-trips.
- Exposing internal SLO configuration fields beyond the panel embedding layer.

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|-------------------------|
| Sub-block structure | Use `single` and `groups` nested blocks inside `slo_overview_config`, mutually exclusive, following the `datatable_config` pattern. | A flat `slo_overview_config` with all fields at once was rejected because it would require complex cross-field validations and obscure the discriminated-union nature of the API. |
| `overview_mode` in schema | Derive `overview_mode` from which sub-block is present at write time; do not expose it as a direct practitioner attribute. | Exposing `overview_mode` as a required string would duplicate the sub-block selection and create an extra validation burden. |
| `group_filters.filters` | Use `filters_json` (a JSON string, normalized) to represent the polymorphic AS-code filter array. | Typed filter objects were rejected because the AS-code filter schema has four discriminated variants (condition, group, DSL, spatial) with deep nested structure, making a Terraform-native representation fragile and hard to maintain. |
| `drilldowns` modeling | Typed `list(object(...))` with `url`, `label`, `trigger`, `type` required and `encode_url`, `open_in_new_tab` optional. | Using `drilldowns_json` was considered and rejected; the drilldown schema is well-defined and a typed list offers plan-time validation and better ergonomics. |
| `slo_instance_id` default | Preserve `null` in state when not configured; do not force `"*"` into state. | Defaulting to `"*"` would cause drift for users who omit the field and expect null semantics, and would complicate import round-trips if Kibana omits it on read. |
| `config_json` support | Not extended to `slo_overview`; typed block only. | Extending `config_json` was rejected because the two-mode API payload would require `config_json` consumers to manually track the discriminant, offering worse UX than the typed approach with no benefit. |
| `group_by` validation | Enum validator on `group_filters.group_by` restricted to `"slo.tags"`, `"status"`, `"slo.indicator.type"`, `"_index"`. | Accepting any string was rejected; the Kibana API rejects unknown values and surfacing invalid values at plan time is more helpful. |

## Risks / Trade-offs

- [Risk] The AS-code `filters` schema may evolve independently of the provider. -> Mitigation: `filters_json` isolates the provider from schema changes; practitioners who use complex filters accept responsibility for valid JSON.
- [Risk] Kibana may inject additional defaults into `drilldowns` on read that are not present on write, causing spurious drift. -> Mitigation: implement read-back normalization for known drilldown defaults during development; document any residual behavior in the requirement.
- [Risk] If `slo_instance_id` is preserved as `null` but Kibana always returns `"*"`, every refresh would show a diff for users who set it explicitly to `"*"`. -> Mitigation: on read, if `slo_instance_id` was null in prior state and Kibana returns `"*"`, preserve null; if it was explicitly set by the practitioner, reconcile from the API value.
- [Risk] The two-embeddable-type API shape means write payloads must send different top-level type keys for the two modes. -> Mitigation: the converter selects the embeddable type from the sub-block presence at write time; this is straightforward and testable.

## Migration Plan

- No user-facing migration is required for existing state.
- The new `slo_overview_config` block is additive; no existing panel type or config block is modified.
- Rollback is a normal code revert; no schema version or state upgrade is needed.
- Existing dashboards that contain `slo_overview` panels managed through Kibana UI can be imported once the implementation ships; import round-trip correctness should be covered by acceptance tests.

## Open Questions

- Confirm whether the Kibana read payload for `slo_overview` panels always returns `slo_instance_id` or omits it when it equals `"*"`. The null-preservation decision above should be re-evaluated against real API behavior during implementation.
- Confirm whether `drilldowns` in the Kibana read response includes additional injected fields beyond those in the write payload that would require normalization.
- Determine whether `hide_border` is included in the actual API response or only in the write path; if it is omitted on read, null preservation or a default will be needed.
