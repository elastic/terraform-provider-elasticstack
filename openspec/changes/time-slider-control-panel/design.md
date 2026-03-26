## Context

The `elasticstack_kibana_dashboard` resource supports a growing list of typed panel config blocks (e.g. `markdown_config`, `xy_chart_config`, `options_list_control_config`). Each typed block is valid for a specific `type` value and is mutually exclusive with all other typed config blocks on a given panel entry. The `time_slider_control` panel type is not a Lens visualization: it carries its own standalone inline config rather than a Lens expression or `config_json` blob.

The Kibana dashboard API exposes `time_slider_control` panels through a schema with a `config` object that has **no required fields**. The entire `config` block is optional, and each of its three attributes — `start_percentage_of_time_range`, `end_percentage_of_time_range`, and `is_anchored` — may be independently omitted. When all config fields are omitted, Kibana uses its defaults for the time slider position and anchoring behavior.

## Goals / Non-Goals

**Goals:**

- Introduce a typed `time_slider_control_config` block that lets practitioners declare the time window position and anchoring state of a time slider control panel in HCL.
- Validate `start_percentage_of_time_range` and `end_percentage_of_time_range` as float64 values in the range 0.0 to 1.0 inclusive at plan time.
- Make the `time_slider_control_config` block itself optional on the panel; all fields within it are also optional.
- Add schema-level validation that `time_slider_control_config` is only used with `type = "time_slider_control"` and conflicts with all other typed config blocks.
- Cover the full panel lifecycle with acceptance tests and the converter logic with unit tests.
- Handle the read-back drift case: when the practitioner omits config fields, preserve the null intent from Terraform state rather than importing Kibana's materialized defaults.

**Non-Goals:**

- Supporting `time_slider_control` panels through the raw `config_json` path (that path remains available for unsupported types; this proposal adds a typed alternative).
- Extending `config_json` type allowlist to include `time_slider_control`.
- Managing or exposing the dashboard-level global time range; the time slider operates within whatever global range is configured.

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|-------------------------|
| Config block name | `time_slider_control_config`, following the `<panel_type>_config` naming convention used by all other typed blocks. | Shorter names like `time_slider_config` or `slider_config` were rejected as less discoverable and potentially ambiguous. |
| Block optionality | The `time_slider_control_config` block itself is optional on the panel. A panel with `type = "time_slider_control"` and no config block is valid and uses Kibana defaults. | Requiring the block would force practitioners to write an empty block, which adds noise and provides no value. |
| Percentage field types | `float64` for `start_percentage_of_time_range` and `end_percentage_of_time_range`. Terraform plugin framework validators enforce 0.0 ≤ value ≤ 1.0. | Integer percentage (0–100) was rejected because the API schema uses fractional values; matching the API avoids a hidden scale conversion. |
| `is_anchored` type | `bool`, optional. When absent from config, the provider omits it from the write payload and does not import the API default into state. | Defaulting to `false` was rejected because it would cause drift when Kibana's default differs or when the practitioner intentionally leaves the field unset. |
| Read-back when config fields are absent | When a config field is not set in Terraform state, the provider SHALL NOT populate it from Kibana's read response even if Kibana returns a value. The null intent SHALL be preserved. | Always importing all Kibana-returned fields was rejected because it would cause permanent plan diffs when the practitioner intentionally omits a field. |
| Panel type exclusivity | `time_slider_control_config` SHALL conflict with all other typed config blocks and `config_json`; enforced by schema validators consistent with existing typed blocks. | A runtime-only check was rejected; plan-time validation gives earlier, cleaner error messages. |

## Risks / Trade-offs

- [Risk] Drift when user sets percentage fields: if the practitioner configures `start_percentage_of_time_range` or `end_percentage_of_time_range`, Kibana returns those values on read and the provider imports them into state. If the practitioner then removes the fields from config without removing the `time_slider_control_config` block, the provider will see null in config but a non-null value in state, producing a plan diff. -> Mitigation: document this behavior; practitioners who want to reset to defaults should omit the entire block or explicitly set the fields to `null`.
- [Risk] Drift when config is omitted entirely: Kibana may or may not include `config` in its read response when no fields are configured. If Kibana materializes an empty config object, the converter must not interpret that as a user-configured empty block. -> Mitigation: treat a nil or empty config object from Kibana as equivalent to an omitted `time_slider_control_config` block in state, consistent with the null-preservation decision above.
- [Risk] Percentage validation at plan time catches only syntactically valid values. If the practitioner passes a value outside 0.0–1.0, the plan fails before reaching Kibana, which is the desired behavior but requires clear error messaging. -> Mitigation: use Terraform framework validators with descriptive messages indicating the valid range.

## Migration Plan

- No user-facing migration is required. Existing dashboards that include `time_slider_control` panels managed through `config_json` continue to work unchanged.
- Practitioners who wish to migrate from `config_json` to the typed block must replace `config_json` with `time_slider_control_config` in their configuration; the underlying Kibana object is unchanged and no resource replacement is triggered.
- No schema version bump or state upgrade is needed.

## Open Questions

- None blocking the proposal. The exact behavior of Kibana when it returns a `config` object with no fields set (versus omitting `config` entirely) should be validated against a real Kibana instance during acceptance testing to confirm whether the null-preservation strategy covers all read-back variants.
