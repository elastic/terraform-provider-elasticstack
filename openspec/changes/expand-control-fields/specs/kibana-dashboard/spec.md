## ADDED Requirements

### Requirement: Control panel width and grow attributes (REQ-039)

The four typed control panel config blocks (`options_list_control_config`, `range_slider_control_config`, `time_slider_control_config`, `esql_control_config`) SHALL each accept the optional panel-level attributes `width` (string, one of `small`, `medium`, or `large`) and `grow` (bool). On create and update, when set, the resource SHALL include these attributes in the API panel body. On read, the resource SHALL apply REQ-009 null-preservation: when prior state had either attribute null, the resource SHALL keep it null even if Kibana returns its server-side default (`width = "medium"`, `grow = false`). On import (no prior state), the resource SHALL leave these attributes null so practitioners are not forced to manage server-side defaults in HCL. Invalid values for `width` (not in the enum) SHALL produce an error diagnostic at plan time.

This requirement applies equally to the four control config blocks when used inside `panels[]` and inside dashboard-level `pinned_panels` (see REQ-038).

#### Scenario: Width and grow round-trip on an in-grid control

- GIVEN a panel with `type = "options_list_control"` whose `options_list_control_config` sets `width = "large"` and `grow = true`
- WHEN create runs and the post-apply read returns the same control
- THEN state SHALL contain `width = "large"` and `grow = true` and a subsequent plan SHALL show no changes

#### Scenario: Width and grow null-preserved on import

- GIVEN an existing dashboard whose options-list control has `width = "medium"` and `grow = false` from Kibana defaults
- WHEN the resource imports the dashboard
- THEN `width` and `grow` SHALL remain null in state and a subsequent plan against a configuration that omits them SHALL show no changes

#### Scenario: Invalid width rejected

- GIVEN any control config block with `width = "huge"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be `small`, `medium`, or `large`

#### Scenario: Width and grow apply to pinned controls

- GIVEN a `pinned_panels` entry with `range_slider_control_config = { ..., width = "small", grow = true }`
- WHEN create runs and the post-apply read returns the same control
- THEN state SHALL contain those attributes on the pinned control entry
