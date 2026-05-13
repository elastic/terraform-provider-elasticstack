## MODIFIED Requirements

### Requirement: Markdown panel behavior (REQ-012)

For `type = "markdown"` panels, the resource SHALL accept a `markdown_config` block whose shape mirrors the API's by-value/by-reference union: `markdown_config = object({ by_value = object({ content, settings, description, hide_title, title, hide_border }), by_reference = object({ ref_id, description, hide_title, title, hide_border }) })`. Exactly one of `by_value` or `by_reference` SHALL be set; setting both or neither SHALL produce an error diagnostic at plan time.

The `by_value` block SHALL require `content` (string) and `settings` (nested object). The `settings` block SHALL accept `open_links_in_new_tab` (bool, optional; when unset, Kibana applies its default of `true`). The `by_value` block SHALL also accept optional `description`, `hide_title`, `title`, and `hide_border`.

The `by_reference` block SHALL require `ref_id` (string) — the unique identifier of an existing markdown library item — and SHALL accept optional `description`, `hide_title`, `title`, and `hide_border`. The resource SHALL NOT validate that the library item exists at plan time; runtime errors from Kibana surface as today.

On write, the resource SHALL build either the `KbnDashboardPanelTypeMarkdownConfig0` (by-value) or `KbnDashboardPanelTypeMarkdownConfig1` (by-reference) API payload according to which sub-block is set. On read, the resource SHALL detect which API branch was returned and populate the matching sub-block, leaving the other branch null. As REQ-010 already specifies, when a markdown panel is managed through `config_json` only, the resource SHALL preserve that JSON-only representation instead of populating typed `markdown_config` sub-blocks.

#### Scenario: By-value markdown panel round-trip

- GIVEN a panel with `type = "markdown"` and `markdown_config = { by_value = { content = "# hi", settings = { open_links_in_new_tab = false }, hide_border = true } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `by_value` shape, `by_reference` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: By-reference markdown panel round-trip

- GIVEN an existing markdown library item with id `md-lib-1` and a panel with `markdown_config = { by_reference = { ref_id = "md-lib-1", title = "shared note" } }`
- WHEN create runs and the post-apply read returns the same panel
- THEN state SHALL contain the same `by_reference` shape, `by_value` SHALL be null, and a subsequent plan SHALL show no changes

#### Scenario: Both sub-blocks set

- GIVEN a panel with both `markdown_config.by_value` and `markdown_config.by_reference` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one sub-block must be set

#### Scenario: Neither sub-block set

- GIVEN a panel with `markdown_config = {}` (no `by_value` and no `by_reference`)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one sub-block must be set

#### Scenario: open_links_in_new_tab unset preserves API default

- GIVEN a `by_value` markdown panel whose `settings.open_links_in_new_tab` is unset
- WHEN create runs and the API stores Kibana's default `true`
- THEN state SHALL keep `settings.open_links_in_new_tab` null (REQ-009 null-preservation)

#### Scenario: config_json continues to manage markdown panels

- GIVEN a panel with `type = "markdown"` configured through `config_json` only
- WHEN create, update, or read runs
- THEN the resource SHALL preserve the JSON-only representation per REQ-010 and SHALL NOT populate typed `markdown_config` sub-blocks
