## Context

The Kibana Dashboard API formalized five flat-sibling fields on every Lens chart root in the recent OpenAPI spec bump (`generated/kbapi/dashboards.json`, all 24 `*NoESQL` and `*ESQL` schemas):

- `time_range` (**required**) — `kbn-es-query-server-timeRangeSchema` with `from`, `to`, optional `mode` enum (`absolute`|`relative`).
- `hide_title` (optional bool).
- `hide_border` (optional bool).
- `drilldowns` (optional array, max 100) — discriminated union of three variants: `dashboard_drilldown`, `discover_drilldown`, `url_drilldown`. Discriminator is `type`. Each variant has a `trigger` enum whose allowed values differ.
- `references` (optional array of `kbn-content-management-utils-referenceSchema`).

The Terraform resource currently:

- Hardcodes every chart panel's `time_range` to `now-15m..now` via `lensPanelTimeRange()` in `internal/kibana/dashboard/models_panels.go`.
- Provides no schema attribute for any of the five fields on the twelve typed chart blocks (`xy_chart_config` … `legacy_metric_config`).

The dashboard-level `time_range` is also required by the API (`kbn-dashboard-data.required = [time_range, …]`), so it is guaranteed to exist as an inheritance source.

The resource is unreleased; breaking changes to the typed chart blocks are explicitly in scope.

## Goals / Non-Goals

**Goals:**

- Make `time_range` first-class on every typed Lens chart block, with default behavior preserving wire-level correctness by inheriting the dashboard-level `time_range` when the chart-level value is null.
- Add `hide_title`, `hide_border`, `references_json`, and a structured `drilldowns` list to every typed Lens chart block.
- Mirror the API shape: flat siblings on each chart root, structured drilldown variants matching the API discriminated union.
- Reuse the existing null-preservation mechanism that the dashboard-level `time_range.mode` already implements (REQ-009).
- Reuse `internal/utils/validators/conditional.go` for inter-variant exclusivity within each drilldown list item.

**Non-Goals:**

- Surfacing chart-level fields that already have typed coverage (`title`, `description`, `filters`, `query`, `data_source`, etc.) — out of scope.
- Touching panels outside the twelve typed Lens chart blocks (markdown, SLO, synthetics, controls, `lens_dashboard_app_config`) — those have their own change proposals.
- Migration paths for users with existing state — the resource is unreleased.
- Modeling `references` as a typed list — chosen JSON for the same reasons existing Lens internals use JSON escape hatches (saved-object refs are rarely hand-authored and the shape can churn).
- Extending the new structured `drilldowns` model to `lens_dashboard_app_config` or SLO panels — out of scope for this change.

## Decisions

### 1. Flat siblings, not a shared `presentation` sub-block

Each chart root in the API places `time_range`, `hide_title`, `hide_border`, `drilldowns`, and `references` as **flat sibling properties** alongside `metrics`, `legend`, `query`, etc. — not nested under any wrapper.

The TF schema follows the same shape exactly. Each of the twelve chart configs gets five new top-level optional attributes.

**Alternative rejected**: shared `presentation { ... }` sub-block. Hurts symmetry with existing flat fields (`title`, `description`) and adds an extra indentation level for no API correspondence.

### 2. `time_range` inheritance from dashboard-level

The API requires `time_range` on every chart root, but the dashboard-level `time_range` is also required and therefore always available as a fallback.

**Write path**: when `panel.<chart>_config.time_range` is null, `panelsToAPI()` copies the dashboard-level `time_range` into the API payload, preserving the API-required invariant.

**Read path**: null-preservation, identical to existing REQ-009 (`time_range.mode`):

- If prior state has `panel.<chart>_config.time_range` as null AND the API-returned value equals the dashboard-level `time_range` → keep state null.
- Otherwise → populate state with the API-returned value.

This gives the natural UX: chart-level `time_range` is opt-in for overrides, otherwise inherits silently.

**Alternative rejected**: continue hardcoding `now-15m..now`. Users would never be able to override or align panels with dashboard time. Already a footgun.

**Alternative rejected**: chart-level `time_range` always required from user. Verbose and redundant when the dashboard-level value is right 95% of the time.

### 3. `drilldowns` is a structured typed list with per-variant sub-blocks

Each list item is an object containing three mutually-exclusive optional sub-blocks:

```hcl
drilldowns = [
  {
    dashboard_drilldown = {
      dashboard_id     = required string
      label            = required string
      trigger          = computed string  # API enum: ["on_apply_filter"] (single value)
      use_filters      = optional bool    # default true
      use_time_range   = optional bool    # default true
      open_in_new_tab  = optional bool    # default false
    }
    discover_drilldown = {
      label            = required string
      trigger          = computed string  # API enum: ["on_apply_filter"] (single value)
      open_in_new_tab  = optional bool    # default true
    }
    url_drilldown = {
      url              = required string
      label            = required string
      trigger          = required string  # stringvalidator.OneOf(...)
      encode_url       = optional bool    # default true
      open_in_new_tab  = optional bool    # default true
    }
  },
]
```

**Per-variant trigger validation:**

- `dashboard_drilldown.trigger` and `discover_drilldown.trigger` are **computed-only** — the API enum has a single allowed value (`on_apply_filter`). The user has no choice; surfacing this as a user-settable attribute is needless noise. The model writes the constant on read; the schema marks it `Computed` with a `UseStateForUnknown` plan modifier.
- `url_drilldown.trigger` is **required** with `stringvalidator.OneOf("on_click_row", "on_click_value", "on_open_panel_menu", "on_select_range")`.

**Inter-variant exclusivity** within each drilldown list item is enforced by attaching object validators built from `validators.AllowedIfDependentPathExpressionOneOf` to each variant sub-block. Concretely, each of the three variant blocks is decorated with two sibling-relative `ForbiddenIfDependentPathExpressionOneOf` validators ensuring the other two variants are unset when this one is set; plus a list-item-level validator that at least one variant must be set per item.

**Alternative rejected**: `drilldowns_json` (single normalized JSON string). Inconsistent with the existing typed SLO drilldown pattern in `slo_burn_rate_config` / `slo_overview_config`. Loses plan-time enum validation and field-level diffs.

**Alternative rejected**: top-level `type` discriminator field (e.g., `drilldowns = [{ type = "url", url = "..." }]`). Less aligned with the existing TF schema pattern where mutually-exclusive variants are modeled as sibling sub-blocks (see `lens_dashboard_app_config.by_value` / `by_reference`).

### 4. `references` as `references_json` (normalized JSON string)

The API shape (`kbn-content-management-utils-referenceSchema` array of `{ name, type, id }`) is small but rarely hand-authored — saved-object references are emitted by Kibana when a panel uses by-reference data sources.

**Decision**: model as `references_json` (normalized JSON string) on each chart block, consistent with existing JSON escape hatches in `models_lens_dashboard_app_converters.go` and other Lens internals.

**Alternative rejected**: typed list of objects. Adds boilerplate to twelve chart blocks for a field users rarely author by hand.

### 5. Reuse existing null-preservation machinery

The `time_range.mode` null-preservation logic (REQ-009 in `openspec/specs/kibana-dashboard/spec.md`) is the model. Apply the same recipe to:

- `panel.<chart>_config.time_range` as a whole (vs dashboard-level equality check).
- `panel.<chart>_config.time_range.mode` (independent of the above).

No new validator package code is needed for null preservation — it's a `ModifyPlan` / model-read pattern already used by the resource.

### 6. Default removal: `lensPanelTimeRange()` retires

The helper in `internal/kibana/dashboard/models_panels.go` returning `{ from: "now-15m", to: "now" }` is deleted. Callers in every `models_*_panel.go` switch to a shared helper that resolves chart-level `time_range` → dashboard-level `time_range`. This helper lives in `models_panels.go` and takes the dashboard model as input so it can read the dashboard-level value during write.

## Risks / Trade-offs

- **State drift on first apply for existing unreleased users.** Anyone with state created against the current code has implicit `now-15m..now` on every panel. After this change, the wire payload starts using the dashboard-level `time_range` instead, causing a one-time diff on the next plan for panels where the dashboard time differs.
  - **Mitigation**: resource is unreleased; documented in proposal. No further action.

- **Import verbosity.** Importing a dashboard populates every panel's `time_range` in state from the API response. If many panels happen to equal the dashboard-level value, the imported config will be verbose.
  - **Mitigation**: users can null-out chart-level `time_range` post-import to opt into inheritance. Acceptable, and consistent with how Terraform import handles other optional-with-defaults fields.

- **Drilldown variant exclusivity is enforced via conditional validators.** If a user sets multiple variant sub-blocks on the same drilldown list item, the error surfaces at plan time. The error message must be clear and point to the offending list item index.
  - **Mitigation**: explicit unit tests for the conflict cases; structured error messages via `AddAttributeError` with the variant path.

- **Computed trigger field on dashboard/discover variants.** A user attempting to set `dashboard_drilldown.trigger` in config will see a plan-time error (`Computed` attributes cannot be set). Acceptable since the API has only one valid value.
  - **Mitigation**: clear schema description noting it is computed and fixed.

- **API spec churn risk for `url_drilldown.trigger` enum.** If Kibana adds a new trigger value, the strict `stringvalidator.OneOf` will reject it at plan time, requiring a provider release.
  - **Mitigation**: accepted cost (consistent with existing enum-strict patterns elsewhere in the resource). Future-proofing via "any string" is explicitly rejected per the design discussion.

- **Read-back time_range equality check assumes Kibana echoes input verbatim.** If Kibana normalizes `now-15m` to a canonical form, the inheritance equality check could fail and unexpectedly populate state when the user intended inheritance.
  - **Mitigation**: acceptance tests in `acc_test.go` verify equality holds across a real Kibana round-trip; if Kibana normalizes, the equality check falls back to literal-string comparison (no semantic time-range parsing).

## Migration Plan

None required. Resource is unreleased; first applies after the change land with the new behavior. Documented in the proposal.
