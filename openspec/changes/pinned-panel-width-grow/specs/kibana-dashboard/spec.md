# Delta spec: `kibana-dashboard` — pinned panel `width`/`grow`

Extends the main spec at `openspec/specs/kibana-dashboard/spec.md`.

## MODIFIED Requirements

### Requirement: Dashboard-level pinned controls round-trip (REQ-038)

The resource SHALL expose dashboard-level pinned controls at the dashboard root as `pinned_panels =
list(object({ type, options_list_control_config, range_slider_control_config,
time_slider_control_config, esql_control_config, width, grow }))`. Each entry SHALL declare exactly
one of the four typed `*_control_config` blocks, and the chosen block SHALL match the entry's `type`
value (for example `type = "options_list_control"` requires `options_list_control_config` and
forbids the other three). `pinned_panels` entries SHALL NOT include a `grid` attribute, and the
resource SHALL NOT populate one on read.

The four `*_control_config` block schemas SHALL be identical to the schemas used for the same
control types under `panels[]`; any future change to those typed schemas applies to both
placements.

Each `pinned_panels` entry SHALL additionally accept two optional sizing attributes, independent of
which `*_control_config` block is set:

- `width` (string): one of `"small"`, `"medium"`, or `"large"`. The resource SHALL reject any other
  value with a validation error diagnostic.
- `grow` (bool): when `true`, the control expands to fill available horizontal space.

When `width` and/or `grow` are set in configuration, the resource SHALL include them in the
corresponding pinned-control-group API payload on create and update. When `width` and/or `grow` are
null in configuration, the resource SHALL omit them from the API payload and SHALL let Kibana apply
its own defaults (`medium` for `width`, `false` for `grow`).

On read, when a prior Terraform value exists for a `pinned_panels` entry matched by `type`, the
resource SHALL preserve a null/unknown `width` or `grow` value rather than overwriting it with the
value Kibana reports (including a Kibana-applied default), so that omitting `width`/`grow` from
configuration produces no diff. When there is no prior Terraform value for the entry (for example
during `terraform import`, or when no prior entry of the same `type` exists at that list position),
the resource SHALL populate `width` and `grow` directly from the API response.

On create and update, the resource SHALL include `pinned_panels` in the dashboard API request body
in the order given in configuration. On read, the resource SHALL repopulate `pinned_panels` from the
API `pinned_panels` array in the order returned. When `pinned_panels` is unset in configuration and
the API returns an empty list (Kibana's default), the resource SHALL keep the Terraform attribute
unset.

#### Scenario: Pinned options-list control round-trip

- GIVEN `pinned_panels = [{ type = "options_list_control", options_list_control_config = { ... } }]`
- WHEN create runs and the post-apply read returns the same control
- THEN state SHALL contain a single `pinned_panels` entry with the matching typed config and no
  `grid`

#### Scenario: Mismatched type and config block

- GIVEN a `pinned_panels` entry with `type = "range_slider_control"` and only
  `options_list_control_config` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the typed block does not match
  `type`

#### Scenario: Multiple typed config blocks set on one entry

- GIVEN a `pinned_panels` entry with both `options_list_control_config` and
  `range_slider_control_config` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one typed block must be set

#### Scenario: Empty pinned_panels preserved as unset

- GIVEN `pinned_panels` is unset in configuration
- WHEN read runs and the API returns `pinned_panels: []`
- THEN the Terraform `pinned_panels` attribute SHALL remain unset

#### Scenario: Configured width and grow are applied

- GIVEN `pinned_panels = [{ type = "options_list_control", width = "small", grow = false,
  options_list_control_config = { ... } }]`
- WHEN create runs
- THEN the resource SHALL include `width = "small"` and `grow = false` in the pinned-control-group
  API payload for that entry
- AND the post-apply read SHALL show `width = "small"` and `grow = false` in state

#### Scenario: Updating width and grow

- GIVEN state contains a `pinned_panels` entry with `width = "small"` and `grow = false`
- AND configuration changes that entry to `width = "large"` and `grow = true`
- WHEN Terraform applies the update
- THEN the plan SHALL show a non-empty diff for `width` and `grow`
- AND the resource SHALL include `width = "large"` and `grow = true` in the update API payload
- AND the post-apply read SHALL show `width = "large"` and `grow = true` in state

#### Scenario: Omitted width and grow produce no diff after refresh

- GIVEN configuration for a `pinned_panels` entry omits both `width` and `grow`
- AND the entry previously applied successfully, so a prior Terraform value exists for it
- WHEN refresh runs and the API reports `width = "medium"` and `grow = false` (Kibana's defaults)
- THEN the Terraform `width` and `grow` attributes SHALL remain null
- AND a subsequent plan-only run SHALL show no diff for `width` or `grow`

#### Scenario: Removing a previously-configured width and grow value

- GIVEN state contains a `pinned_panels` entry with `width = "large"` and `grow = true`
- AND configuration is changed to omit both attributes from that entry
- WHEN Terraform applies the update
- THEN the resource SHALL omit `width` and `grow` from the update API payload
- AND the post-apply read SHALL show `width` and `grow` as null in state, not the removed values

#### Scenario: Width and grow import directly with no prior state

- GIVEN a dashboard exists in Kibana with a pinned options-list control whose API response reports
  `width = "small"` and `grow = false`
- WHEN `terraform import` reads that dashboard
- THEN the imported state SHALL contain `width = "small"` and `grow = false` for that
  `pinned_panels` entry, read directly from the API response rather than left null

#### Scenario: Invalid width value is rejected

- GIVEN a `pinned_panels` entry with `width = "huge"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return a validation error diagnostic indicating `width` must be one of
  `"small"`, `"medium"`, or `"large"`

#### Scenario: Width and grow are consistent across all four pinned control types

- GIVEN `pinned_panels` entries of each of the four supported types (`options_list_control`,
  `range_slider_control`, `esql_control`, `time_slider_control`), each with `width` and `grow` set
- WHEN create runs and the post-apply read returns the same controls
- THEN state SHALL contain the configured `width` and `grow` values for every entry, regardless of
  control type
