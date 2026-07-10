## ADDED Requirements

### Requirement: `links_config` panel block (REQ-LINKS-001)

The `panels` list in `elasticstack_kibana_dashboard` SHALL accept entries with `type = "links"` by providing a `links_config` block. The block SHALL follow the same mutual-exclusion, `AllowedIf`/`RequiredIf`, and sibling `ConflictsWith` patterns as all other typed panel blocks. When a panel carries `type = "links"`, the `links_config` block SHALL be required; omitting it SHALL produce a plan-time error.

The block accepts exactly one of two branches:

**`by_value`** — inline link configuration:
- `layout` (required string, enum `"horizontal"` | `"vertical"`)
- `links` (required list, at least 1 item)
- `title`, `description`, `hide_title`, `hide_border` (all optional)

**`by_reference`** — references a Kibana Links library saved object:
- `ref_id` (required string, non-empty)
- `title`, `description`, `hide_title`, `hide_border` (all optional)

Setting both `by_value` and `by_reference`, or neither, SHALL produce a plan-time error.

Each item in `by_value.links[]` is a flat object with:
- `type` (required string, enum `"dashboard"` | `"external"`)
- `destination` (required string — dashboard saved-object id or URL)
- `label` (optional string)
- `open_in_new_tab` (optional bool — all types)
- `use_filters` (optional bool — `type = "dashboard"` only)
- `use_time_range` (optional bool — `type = "dashboard"` only)
- `encode_url` (optional bool — `type = "external"` only)

The `"dashboard"` type maps to the API discriminator `"dashboardLink"`; `"external"` maps to `"externalLink"`.

Optional display fields (`title`, `description`, `hide_title`, `hide_border`) on both branches and optional link item fields SHALL follow REQ-009 null-preservation on refresh/read after a user-managed apply: they remain null in state when omitted by the user, even if Kibana echoes server-side defaults. On import, these fields SHALL be left null in state when Kibana returns only server-side defaults, so practitioners are not forced to manage those defaults in HCL.

#### Scenario: `by_value` panel with dashboard and external links

- GIVEN a `links` panel with `links_config.by_value` containing a `"dashboard"` link (`destination`, `label`, `open_in_new_tab`, `use_filters`, `use_time_range`) and an `"external"` link (`destination`, `label`, `open_in_new_tab`, `encode_url`)
- WHEN create runs and the post-apply read returns the panel
- THEN state SHALL reflect all configured fields and a subsequent plan SHALL show no changes

#### Scenario: `by_reference` panel referencing a library saved object

- GIVEN a `links` panel with `links_config.by_reference` carrying a `ref_id` and optional `title`
- WHEN create runs and the post-apply read returns the panel
- THEN state SHALL contain `ref_id` and `title` and a subsequent plan SHALL show no changes

#### Scenario: Mutual exclusion — both branches set

- GIVEN a `links_config` block with both `by_value` and `by_reference` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one must be set

#### Scenario: Mutual exclusion — neither branch set

- GIVEN a `links_config` block with neither `by_value` nor `by_reference` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating exactly one must be set

#### Scenario: `links_config` required for `type = "links"`

- GIVEN a panel with `type = "links"` and no `links_config` block
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating `links_config` is required

#### Scenario: `layout` validation

- GIVEN a `by_value` block with `layout = "diagonal"`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating the value must be `horizontal` or `vertical`

#### Scenario: `links` minimum length

- GIVEN a `by_value` block with an empty `links = []`
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating at least one item is required

#### Scenario: Link item type-specific field isolation — `encode_url` on `dashboard` link

- GIVEN a link item with `type = "dashboard"` and `encode_url` set to a concrete value
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating `encode_url` is not valid for `type = "dashboard"`

#### Scenario: Link item type-specific field isolation — `use_filters` on `external` link

- GIVEN a link item with `type = "external"` and `use_filters` set to a concrete value
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic indicating `use_filters` is not valid for `type = "external"`

#### Scenario: Null-preservation of optional display fields

- GIVEN a `by_value` links panel whose `title` and `hide_border` are not set in configuration
- WHEN Kibana returns `title = "Links"` and `hide_border = false` in the API response
- THEN state SHALL keep `title` and `hide_border` null and a subsequent plan SHALL show no changes

#### Scenario: Import — `by_value` panel

- GIVEN an existing Kibana dashboard with a `links` panel in `by_value` configuration
- WHEN the resource imports the dashboard
- THEN the `links_config.by_value` block SHALL be populated from the API response — including `layout`, all `links[]` items, and any optional display fields returned by Kibana — and a subsequent plan against a matching configuration SHALL show no changes

#### Scenario: Import — `by_reference` panel

- GIVEN an existing Kibana dashboard with a `links` panel in `by_reference` configuration
- WHEN the resource imports the dashboard
- THEN the `links_config.by_reference` block SHALL be populated from the API response — including `ref_id` and any optional display fields returned by Kibana — and a subsequent plan against a matching configuration SHALL show no changes

#### Scenario: Optional display fields null on import

- GIVEN an existing Kibana dashboard whose `links` panel has server-side defaults for `hide_title` (`false`) and `hide_border` (`false`)
- WHEN the resource imports the dashboard and the user's configuration omits those attributes
- THEN `hide_title` and `hide_border` SHALL remain null in state and a subsequent plan against a matching configuration SHALL show no changes
