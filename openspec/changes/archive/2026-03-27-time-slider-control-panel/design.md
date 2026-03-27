## Context

The `elasticstack_kibana_dashboard` resource supports a growing list of typed panel config blocks (e.g. `markdown_config`, `xy_chart_config`, `options_list_control_config`). Each typed block is valid for a specific `type` value and is mutually exclusive with all other typed config blocks on a given panel entry. The `time_slider_control` panel type is not a Lens visualization: it carries its own standalone inline config rather than a Lens expression or `config_json` blob.

The Kibana dashboard API exposes `time_slider_control` panels through a schema with a `config` object that has **no required fields**. The entire `config` block is optional, and each of its three attributes — `start_percentage_of_time_range`, `end_percentage_of_time_range`, and `is_anchored` — may be independently omitted. When all config fields are omitted, Kibana uses its defaults for the time slider position and anchoring behavior.

## Goals / Non-Goals

**Goals:**

- Introduce a typed `time_slider_control_config` block that lets practitioners declare the time window position and anchoring state of a time slider control panel in HCL.
- Validate `start_percentage_of_time_range` and `end_percentage_of_time_range` as float32 values in the range 0.0 to 1.0 inclusive at plan time, aligned with the Kibana API type.
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
| Percentage field types | `float32` for `start_percentage_of_time_range` and `end_percentage_of_time_range`. Terraform plugin framework validators enforce 0.0 ≤ value ≤ 1.0. This matches the generated Kibana client (`*float32`) and prevents refresh drift when HCL decimals are widened to float64 in state. | Integer percentage (0–100) was rejected because the API uses fractions. Keeping `float64` in schema while normalizing read-back to float32 bits was a smaller-scope alternative but leaves a wider state type than the API. Rounding to two decimal places on read was rejected as a primary approach because it is lossy. |
| `is_anchored` type | `bool`, optional. When absent from config, the provider omits it from the write payload and does not import the API default into state. | Defaulting to `false` was rejected because it would cause drift when Kibana's default differs or when the practitioner intentionally leaves the field unset. |
| Read-back when config fields are absent | When a config field is not set in Terraform state, the provider SHALL NOT populate it from Kibana's read response even if Kibana returns a value. The null intent SHALL be preserved. | Always importing all Kibana-returned fields was rejected because it would cause permanent plan diffs when the practitioner intentionally omits a field. |
| Panel type exclusivity | `time_slider_control_config` SHALL conflict with all other typed config blocks and with practitioner-authored `config_json` on the same panel. Schema validators enforce the allowlist on `config_json` by panel `type`. The panel object validator documents `time_slider_control` + `config_json` rules in its description only, avoiding a second diagnostic for the same error. | A runtime-only check was rejected; plan-time validation gives earlier, cleaner error messages. |

## Risks / Trade-offs

- [Risk] Drift when user sets percentage fields: if the practitioner configures `start_percentage_of_time_range` or `end_percentage_of_time_range`, Kibana returns those values on read and the provider imports them into state. If the practitioner then removes the fields from config without removing the `time_slider_control_config` block, the provider will see null in config but a non-null value in state, producing a plan diff. -> Mitigation: document this behavior; practitioners who want to reset to defaults should omit the entire block or explicitly set the fields to `null`. **API/type alignment**: modeling percentages as float32 mitigates a separate class of drift where float64 state disagrees with float32 API round-trip for the same literal.
- [Risk] Drift when config is omitted entirely: Kibana may or may not include `config` in its read response when no fields are configured. If Kibana materializes an empty config object, the converter must not interpret that as a user-configured empty block. -> Mitigation: treat a nil or empty config object from Kibana as equivalent to an omitted `time_slider_control_config` block in state, consistent with the null-preservation decision above.
- [Risk] Percentage validation at plan time catches only syntactically valid values. If the practitioner passes a value outside 0.0–1.0, the plan fails before reaching Kibana, which is the desired behavior but requires clear error messaging. -> Mitigation: use Terraform framework validators with descriptive messages indicating the valid range.

## Migration Plan

- Practitioner-authored `config_json` for `time_slider_control` is not supported and is rejected at plan time; there is no migration path from that pattern because it was never valid for this resource. Express time slider settings with the optional `time_slider_control_config` block (or omit config for Kibana defaults).
- Imported dashboards and API-created panels continue to map through the typed read path; computed `config_json` may appear in state after read and must not be copied back into configuration as an authoring shortcut.
- **State upgrades**: the provider does not implement a resource `StateUpgrader` for dashboard panel attribute renames or type changes. The shipped percentage fields are `float32` from the start of this feature in-tree; no automated state migration is required for that path. If an older experimental build stored the same logical fields as `float64`, upgrading to this schema is a type change—expect Terraform to normalize values to float32 on read/plan; do not depend on float64 precision in configuration.

## Open Questions

- None blocking the proposal. The exact behavior of Kibana when it returns a `config` object with no fields set (versus omitting `config` entirely) should be validated against a real Kibana instance during acceptance testing to confirm whether the null-preservation strategy covers all read-back variants.
