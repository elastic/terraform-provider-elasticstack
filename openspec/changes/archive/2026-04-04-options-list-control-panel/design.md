## Context

The `elasticstack_kibana_dashboard` resource supports a growing list of typed panel config blocks (e.g. `markdown_config`, `xy_chart_config`, `metric_chart_config`). Each typed block is valid for a specific `type` value and is mutually exclusive with all other typed config blocks on a given panel entry. The `options_list_control` panel type is not a Lens visualization: it carries its own standalone inline config (data view reference, field selection, display preferences, filtering behavior) rather than a Lens expression or `config_json` blob.

The Kibana dashboard API exposes `options_list_control` panels through the `kbn-dashboard-panel-options_list_control` schema, which has a required `config` object containing `data_view_id` and `field_name` as required fields plus a set of optional filtering and display attributes.

## Goals / Non-Goals

**Goals:**

- Introduce a typed `options_list_control_config` block that lets practitioners declare all attributes of an options list control panel in HCL, with required fields enforced at plan time.
- Validate `search_technique` as one of `prefix`, `wildcard`, or `exact`.
- Map `display_settings` as a nested block with individual boolean and string attributes.
- Map `sort` as a nested block with `by` and `direction` string attributes.
- Map `selected_options` as a list of strings (the API accepts string or number; the Terraform model unifies to string).
- Add schema-level validation that `options_list_control_config` is only used with `type = "options_list_control"` and conflicts with all other typed config blocks.
- Cover the full panel lifecycle with acceptance tests and the converter logic with unit tests.

**Non-Goals:**

- Supporting `options_list_control` panels through the raw `config_json` path (that path remains available for unsupported types; this proposal adds a typed alternative).
- Managing the referenced data view lifecycle (the `data_view_id` is a reference attribute, not a managed dependency).
- Extending `config_json` type allowlist to include `options_list_control`.

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|-------------------------|
| Config block name | `options_list_control_config`, following the `<panel_type>_config` naming convention used by all other typed blocks. | Shorter names like `options_list_config` or `control_config` were rejected as less discoverable. |
| Required fields | `data_view_id` and `field_name` are required inside the block; all other attributes are optional. | Making `data_view_id` optional was rejected because the API schema marks it required and omitting it produces an invalid panel. |
| `display_settings` representation | A nested block with individual typed attributes (`placeholder`, `hide_action_bar`, `hide_exclude`, `hide_exists`, `hide_sort`), not a `display_settings_json` string. | A JSON string blob was considered but rejected in favour of typed HCL for practitioner usability. |
| `sort` representation | A nested block with `by` (string) and `direction` (string) attributes. | Flattening to `sort_by` / `sort_direction` top-level attributes was considered but rejected to match the API structure and allow future extension. |
| `search_technique` validation | Enum validation at schema level: values must be one of `prefix`, `wildcard`, or `exact`. | Free-form string was rejected to catch typos at plan time. |
| `selected_options` type | `list(string)`. The API accepts `string \| number` values in the array; the provider coerces numbers to strings on read so the Terraform model is uniform. | `list(any)` was rejected because Terraform plugin framework requires a concrete element type. |
| Panel type exclusivity | `options_list_control_config` SHALL conflict with all other typed config blocks and `config_json`; this is enforced by schema validators consistent with existing typed blocks. | A runtime-only check was rejected; plan-time validation gives earlier, cleaner error messages. |

## Risks / Trade-offs

- [Risk] The `selected_options` string coercion may not round-trip numeric options without loss of type fidelity in the API. -> Mitigation: document that the provider stores selected options as strings; practitioners who need numeric-typed options should be aware that Kibana may persist them as numbers while the provider treats them as strings.
- [Risk] Kibana may emit additional optional fields in the `options_list_control` config on read-back that are not yet modelled. -> Mitigation: the converter ignores unmapped fields on read, consistent with the pattern used by other typed panel converters.
- [Risk] The `display_settings` nested block means that omitting the block entirely and omitting all its individual attributes are two different states, which could cause plan noise if Kibana materializes an empty display_settings object. -> Mitigation: treat a nil or empty `display_settings` object from Kibana as equivalent to an omitted block in state.

## Migration Plan

- No user-facing migration is required. Existing dashboards that include `options_list_control` panels managed through `config_json` continue to work unchanged.
- Practitioners who wish to migrate from `config_json` to the typed block must replace `config_json` with `options_list_control_config` in their configuration; the underlying Kibana object is unchanged and no resource replacement is triggered.
- No schema version bump or state upgrade is needed.

## Open Questions

- None blocking the proposal. The exact coercion behavior for numeric `selected_options` values should be validated against a real Kibana instance during acceptance testing.
