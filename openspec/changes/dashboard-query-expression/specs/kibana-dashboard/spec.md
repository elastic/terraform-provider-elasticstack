## REMOVED Requirements

### Requirement: Dashboard root schema API naming — text/json query union (REQ-036)

The resource SHALL expose dashboard-level time selection, refresh, and query using nested attribute objects whose names mirror the Kibana Dashboard API JSON: `time_range` (`from`, `to`, optional `mode`), `refresh_interval` (`pause`, `value`), and `query` (`language`, restricted to `kql` or `lucene`, with exactly one of `text` or `json` for the query expression). The root `time_range`, `refresh_interval`, and `query` attributes SHALL be **optional**, matching the Kibana Dashboards API where only `title` is required on create and update. Their nested attributes SHALL remain required when the parent block is configured: `time_range.from` / `time_range.to`, `refresh_interval.pause` / `refresh_interval.value`, and `query.language`. The resource SHALL NOT set a schema default, plan modifier default, or `Computed` flag on any of the three root blocks, and the schema version SHALL NOT change.

The resource SHALL expose dashboard `options` with the API-aligned flags `auto_apply_filters` and `hide_panel_borders` in addition to the existing option fields.

#### Scenario: Query union uses text branch

- GIVEN `query = { language = "kql" text = "http.response.status_code:200" }`
- WHEN the provider builds the create or update request body
- THEN it SHALL set the API query expression from `query.text` and SHALL set `query.language` from `query.language`

#### Scenario: Query union uses json branch

- GIVEN `query = { language = "kql" json = jsonencode({ ... }) }`
- WHEN the provider builds the create or update request
- THEN it SHALL set the API query expression from `query.json` and SHALL reject configurations where both `text` and `json` are set, or where neither is set

#### Scenario: Options include new flags

- GIVEN `options { hide_panel_borders = true auto_apply_filters = false }`
- WHEN create or update runs
- THEN the provider SHALL include those fields in the API `options` object when known

#### Scenario: Title-only configuration is valid

- GIVEN a configuration containing only `title` (no `time_range`, `refresh_interval`, or `query`)
- WHEN Terraform validates and plans the resource
- THEN validation SHALL succeed with no diagnostics about missing required root attributes

#### Scenario: Nested attributes remain required when the block is set

- GIVEN a configuration with `time_range = { from = "now-7d" }` (missing `to`)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic for the missing nested `to` attribute

**Reason:** Replaced by "Dashboard root schema API naming — expression query (REQ-036)" (below, under ADDED) that aligns `query` with the single-field Kibana API shape (`expression` instead of `text`/`json`) and bumps the schema version. See issue #5081.

## MODIFIED Requirements

### Requirement: Create and update request mapping (REQ-007)

On create and update, the resource SHALL map Terraform state to the dashboard API request body using `title`, `description`, nested `time_range`, nested `refresh_interval`, nested `query`, tags, options, panels, and sections when those values are known. `access_control` SHALL be sent on create when known. The current regenerated Kibana `PUT /dashboards/{id}` request body does not expose `access_control`, so updates SHALL preserve prior `access_control` state but SHALL NOT claim to mutate it through the dashboard update request until the API surface supports that field. Query mapping SHALL send `query.expression` as the API query expression and `query.language` as the API query language, a direct one-to-one copy. On create without a practitioner-supplied `dashboard_id`, the provider SHALL call the `POST /dashboards` API and let Kibana assign the dashboard id; when `dashboard_id` is configured, it SHALL create via `PUT /dashboards/{id}` as described in REQ-003 and REQ-003a. If conversion of query or panel data fails, the operation SHALL return diagnostics and SHALL NOT proceed with the dashboard API call. When `time_range`, `refresh_interval`, or `query` is null in the plan or state, the provider SHALL omit that field from the request body on both create (`POST /dashboards`) and update (`PUT /dashboards/{id}`). Because the Kibana `PUT /dashboards/{id}` operation is a full replace, omitting a block on update SHALL clear it in Kibana rather than preserve its previous value. The provider SHALL NOT send zero-value `time_range`, `refresh_interval`, `query`, or `options` objects for a null block, since Kibana rejects them (for example HTTP 400 on `query.language`). To make this possible, the hand-maintained `PUT /api/dashboards/{id}` overlay in `generated/kbapi/dashboard-paths.json` SHALL list only `title` as required, and the generated `PutDashboardsIdJSONBody` SHALL expose `time_range`, `refresh_interval`, `query`, and `options` as optional (pointer) fields.

After a successful create or update, the resource SHALL read the dashboard back and use that read as the authoritative final state; if the dashboard cannot be read back, the operation SHALL fail.

#### Scenario: Create omits unset root blocks

- GIVEN a configuration with only `title` set
- WHEN create runs
- THEN the `POST /dashboards` request body SHALL contain `title` and SHALL NOT contain `time_range`, `refresh_interval`, or `query`

#### Scenario: Update omits unset root blocks

- GIVEN a configuration with only `title` set and prior state with the same configuration
- WHEN update runs
- THEN the `PUT /dashboards/{id}` request body SHALL NOT contain `time_range`, `refresh_interval`, `query`, or `options`
- AND the request SHALL NOT be rejected by Kibana for zero-value blocks

#### Scenario: Removing a block clears it

- GIVEN prior state with `time_range`, `refresh_interval`, and `query` set
- AND the configuration is changed to remove all three blocks
- WHEN update runs
- THEN the `PUT /dashboards/{id}` request body SHALL omit all three
- AND after the post-apply read the three blocks SHALL be null in state (not reset to a Kibana default)

#### Scenario: Post-apply authoritative read

- GIVEN create or update succeeds
- WHEN the provider finalizes state
- THEN it SHALL re-read the dashboard and SHALL fail if the dashboard cannot be retrieved

## ADDED Requirements

### Requirement: Dashboard root schema API naming — expression query (REQ-036)

The resource SHALL expose dashboard-level time selection, refresh, and query using nested attribute objects whose names mirror the Kibana Dashboard API JSON: `time_range` (`from`, `to`, optional `mode`), `refresh_interval` (`pause`, `value`), and `query` (`language`, restricted to `kql` or `lucene`, and `expression`, the KQL/Lucene query string). The root `time_range`, `refresh_interval`, and `query` attributes SHALL be **optional**, matching the Kibana Dashboards API where only `title` is required on create and update. Their nested attributes SHALL remain required when the parent block is configured: `time_range.from` / `time_range.to`, `refresh_interval.pause` / `refresh_interval.value`, and `query.language` / `query.expression`. The resource SHALL NOT set a schema default or plan modifier default on any of the three root blocks.

The resource SHALL expose dashboard `options` with the API-aligned flags `auto_apply_filters` and `hide_panel_borders` in addition to the existing option fields.

#### Scenario: Query expression maps directly to the API field

- GIVEN `query = { language = "kql" expression = "http.response.status_code:200" }`
- WHEN the provider builds the create or update request body
- THEN it SHALL set the API query expression from `query.expression` and SHALL set `query.language` from `query.language`

#### Scenario: Empty-string expression is valid

- GIVEN `query = { language = "kql" expression = "" }`
- WHEN Terraform validates and plans the resource
- THEN validation SHALL succeed and the provider SHALL send `expression = ""` in the request body without rejecting it as missing

#### Scenario: Missing expression is rejected

- GIVEN a configuration with `query = { language = "kql" }` and no `expression` set
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic for the missing required `expression` attribute

#### Scenario: Options include new flags

- GIVEN `options { hide_panel_borders = true auto_apply_filters = false }`
- WHEN create or update runs
- THEN the provider SHALL include those fields in the API `options` object when known

#### Scenario: Title-only configuration is valid

- GIVEN a configuration containing only `title` (no `time_range`, `refresh_interval`, or `query`)
- WHEN Terraform validates and plans the resource
- THEN validation SHALL succeed with no diagnostics about missing required root attributes

#### Scenario: Nested attributes remain required when the block is set

- GIVEN a configuration with `time_range = { from = "now-7d" }` (missing `to`)
- WHEN Terraform validates the configuration
- THEN the resource SHALL return an error diagnostic for the missing nested `to` attribute

### Requirement: Dashboard root query state upgrade to `expression` (REQ-055)

The `elasticstack_kibana_dashboard` resource SHALL implement a state upgrade from schema version 1 to version 2 that migrates the root `query` block from the `text` / `json` union shape to the single `expression` attribute described in REQ-036. The resource schema version SHALL be incremented from 1 to 2.

The v1 → v2 upgrader SHALL:

1. Leave a null or absent `query` block untouched.
2. For a non-null `query` block, set `query.expression` from `query.text` when `text` is a non-null string; otherwise set it from `query.json` when `json` is present, copying the stored string value unchanged.
3. Remove `query.text` and `query.json` from the upgraded state.

The resource's existing v0 → v1 upgrader (REQ-040) SHALL be composed with this v1 → v2 transform so that state starting at schema version 0 upgrades directly to version 2 in a single `UpgradeState` call, without requiring two separate `terraform apply` runs. No data SHALL be lost during either upgrade path: the resulting `query.expression` value SHALL be functionally equivalent to the query expression Kibana was already receiving before the upgrade.

#### Scenario: v1 state with `query.text` upgrades to `expression`

- GIVEN a v1 state containing `query = { language = "kql", text = "http.response.status_code:200" }`
- WHEN the state upgrader runs
- THEN the v2 state SHALL contain `query = { language = "kql", expression = "http.response.status_code:200" }`
- AND `query.text` SHALL NOT be present in the upgraded state

#### Scenario: v1 state with `query.json` upgrades to `expression` unchanged

- GIVEN a v1 state containing `query = { language = "kql", json = "{\"match_all\":{}}" }`
- WHEN the state upgrader runs
- THEN the v2 state SHALL contain `query.expression` equal to the exact string `{"match_all":{}}`
- AND `query.json` SHALL NOT be present in the upgraded state

#### Scenario: Null query block is left untouched

- GIVEN a v1 state with no `query` block set (null)
- WHEN the state upgrader runs
- THEN the v2 state SHALL also have `query` null

#### Scenario: v0 state upgrades directly to v2

- GIVEN a v0 state containing an `options_list_control` panel with flat attributes and a root `query.text` value
- WHEN the state upgrader runs
- THEN the resulting v2 state SHALL have the panel's flat attributes relocated under `by_field {}` per REQ-040
- AND the root `query` block SHALL have `expression` populated from the prior `text` value, with `text` and `json` absent
