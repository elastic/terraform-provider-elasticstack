## ADDED Requirements

### Requirement: Dashboard-level pinned controls round-trip (REQ-038)

The resource SHALL expose dashboard-level pinned controls at the dashboard root as `pinned_panels = list(object({ type, options_list_control_config, range_slider_control_config, time_slider_control_config, esql_control_config }))`. Each entry SHALL declare exactly one of the four typed `*_control_config` blocks, and the chosen block SHALL match the entry's `type` value (for example `type = "options_list_control"` requires `options_list_control_config` and forbids the other three). `pinned_panels` entries SHALL NOT include a `grid` attribute, and the resource SHALL NOT populate one on read.

The four `*_control_config` block schemas SHALL be identical to the schemas used for the same control types under `panels[]`; any future change to those typed schemas applies to both placements.

On create and update, the resource SHALL include `pinned_panels` in the dashboard API request body in the order given in configuration. On read, the resource SHALL repopulate `pinned_panels` from the API `pinned_panels` array in the order returned. When `pinned_panels` is unset in configuration and the API returns an empty list (Kibana's default), the resource SHALL keep the Terraform attribute unset.

#### Scenario: Pinned options-list control round-trip

- GIVEN `pinned_panels = [{ type = "options_list_control", options_list_control_config = { ... } }]`
- WHEN create runs and the post-apply read returns the same control
- THEN state SHALL contain a single `pinned_panels` entry with the matching typed config and no `grid`

#### Scenario: Mismatched type and config block

- GIVEN a `pinned_panels` entry with `type = "range_slider_control"` and only `options_list_control_config` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the typed block does not match `type`

#### Scenario: Multiple typed config blocks set on one entry

- GIVEN a `pinned_panels` entry with both `options_list_control_config` and `range_slider_control_config` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one typed block must be set

#### Scenario: Empty pinned_panels preserved as unset

- GIVEN `pinned_panels` is unset in configuration
- WHEN read runs and the API returns `pinned_panels: []`
- THEN the Terraform `pinned_panels` attribute SHALL remain unset
