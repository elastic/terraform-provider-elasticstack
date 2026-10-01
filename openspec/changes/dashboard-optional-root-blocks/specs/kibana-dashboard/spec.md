## MODIFIED Requirements

### Requirement: Dashboard root schema API naming (REQ-036)

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

### Requirement: Create and update request mapping (REQ-007)

On create and update, the resource SHALL map Terraform state to the dashboard API request body using `title`, `description`, nested `time_range`, nested `refresh_interval`, nested `query`, tags, options, panels, and sections when those values are known. `access_control` SHALL be sent on create when known. The current regenerated Kibana `PUT /dashboards/{id}` request body does not expose `access_control`, so updates SHALL preserve prior `access_control` state but SHALL NOT claim to mutate it through the dashboard update request until the API surface supports that field. Query mapping SHALL send `query.text` as the string form of the API query expression and `query.json` as the JSON-object form of the same expression field. On create, the provider SHALL call the `POST /dashboards` API and let Kibana assign the dashboard id. If conversion of query or panel data fails, the operation SHALL return diagnostics and SHALL NOT proceed with the dashboard API call. When `time_range`, `refresh_interval`, or `query` is null in the plan or state, the provider SHALL omit that field from the request body on both create (`POST /dashboards`) and update (`PUT /dashboards/{id}`). Because the Kibana `PUT /dashboards/{id}` operation is a full replace, omitting a block on update SHALL clear it in Kibana rather than preserve its previous value. The provider SHALL NOT send zero-value `time_range`, `refresh_interval`, `query`, or `options` objects for a null block, since Kibana rejects them (for example HTTP 400 on `query.language`). To make this possible, the hand-maintained `PUT /api/dashboards/{id}` overlay in `generated/kbapi/dashboard-paths.json` SHALL list only `title` as required, and the generated `PutDashboardsIdJSONBody` SHALL expose `time_range`, `refresh_interval`, `query`, and `options` as optional (pointer) fields.

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

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009)

When Kibana omits or defaults fields on read, the resource SHALL preserve prior Terraform intent to avoid inconsistent results and spurious drift where the implementation supports that behavior. The resource preserves the prior `time_range.mode` value already held in state or plan instead of overwriting it from read-back when the GET response does not supply a usable mode. When the GET dashboard API does not supply a usable `access_control.access_mode` value, the resource SHALL clear `access_control` in Terraform state rather than leaving a stale prior value behind. When the options block was omitted in Terraform and Kibana materializes only the default dashboard options matching the implementation's `isDashboardOptionsDefaultSet` helper (including `auto_apply_filters` and `hide_panel_borders` at their API defaults when applicable), the resource SHALL keep the `options` block null in state. When a section's prior `collapsed` value was null and Kibana returns `false`, the resource SHALL preserve null rather than forcing `false` into state.

The root `time_range`, `refresh_interval`, and `query` blocks are not subject to null-preservation or default detection. Kibana applies no defaults to these blocks and returns them as absent when they were not sent (verified against Kibana 9.4.2 and 9.6.0-SNAPSHOT), so the resource SHALL map them from the API response as-is: a block present in the response is populated in state, and a block absent from the response is null in state. The existing `time_range.mode` preservation continues to apply when `time_range` is present. If a future Kibana version begins returning defaults for these blocks, the title-only acceptance test is the intended detection mechanism and a follow-up change would add null-preservation.

For panel reads, the provider SHALL seed each panel from prior practitioner intent before finalizing state: from the prior plan on the post-create and post-update read-back, and from prior state on refresh. After that seed, it SHALL apply panel-type-specific alignment so Kibana-injected defaults or omitted optional values do not overwrite practitioner intent. This alignment includes preserving configured titles and descriptions when the API returns blank values, preserving ES|QL control `esql_query`, `title`, and `available_options` when the API omits them, preserving raw `config_json` when the read-back only differs by omitted optional `filters` or `query` keys, and preserving semantically equivalent optional JSON defaults such as `rank_by` in metric and tagcloud configurations.

For typed panel config blocks whose `PopulateFromAPI` receives both `pm` (the panel model being built, which callers always pass zero-valued to avoid aliasing plan pointers) and `prior` (the prior plan or state panel at the same index), the null-preservation decision for that block SHALL key on `prior.<Type>Config`, not on `pm`'s own field: `prior.<Type>Config != nil` SHALL be treated as a same-type update (honor the practitioner's null intent for optional fields), and `prior.<Type>Config == nil` SHALL be treated as creation, import, or a genuine type change: there is no prior null intent to honor, so the block SHALL be rebuilt from the API. For config blocks that are optional even when the panel type matches (`synthetics_monitors_config`, `synthetics_stats_overview_config`), the block SHALL instead be left null when the API response carries no content for it, rather than materializing an empty block. `pm`'s own field state SHALL NOT be used for this decision, since it never carries prior intent into `PopulateFromAPI`.

As of Kibana 9.5.0 GA, several typed panel config blocks' optional enum-shaped fields (for example `aiops_pattern_analysis_config.minimum_time_range` and `.random_sampler_mode`) receive concrete server-side default values on read where earlier Kibana versions returned no value at all. The provider SHALL apply REQ-009 null-preservation to these fields exactly as it does for any other Kibana-injected default: when the practitioner left the field unset (null in prior state/plan), the resource SHALL keep it null in state even though the API now returns a concrete default, rather than materializing that default and producing "Provider produced inconsistent result after apply".

The resource models only the currently supported Terraform subset of dashboard fields. Fields present in the Kibana dashboard API but not modeled by this resource — for example top-level `project_routing` — are outside this resource contract (see REQ-037 for `filters` and REQ-038 for `pinned_panels`).

The provider SHALL treat an API-returned `""` for `description` as semantically equivalent to an omitted field when prior plan/state had `description` null, restoring null in state rather than propagating the API-echoed empty string. This is an instance of REQ-009 null-preservation applied to the dashboard root `description`. This SHALL be consistent with the null/empty-string normalization already applied to XY chart `fitting.type`, `fitting.end_value`, and panel-level `time_range`.

#### Scenario: Empty-string description treated as null for null-intent practitioners

- GIVEN a practitioner has never set `description` on a dashboard (prior state: null)
- AND Kibana 9.5 returns `description: ""` on a subsequent read or post-apply read-back
- WHEN the provider applies REQ-009 null-preservation to `description`
- THEN state SHALL contain `description = null` and no drift SHALL be reported on the next plan

#### Scenario: Root blocks omitted in config stay null

- GIVEN a configuration with only `title` set, applied and read back
- WHEN Kibana returns the dashboard without `time_range`, `refresh_interval`, or `query`
- THEN state SHALL contain null for all three blocks
- AND a subsequent plan SHALL show no changes
- AND importing the dashboard SHALL produce the same state without drift

#### Scenario: Options omitted in config

- GIVEN Terraform configuration omitted the `options` block
- WHEN Kibana read-back contains only the dashboard option defaults
- THEN the resource SHALL keep `options` unset in state

#### Scenario: Typed panel config null-preservation keys on prior, not on pm

- GIVEN a typed panel config block (e.g. `aiops_pattern_analysis_config`) whose practitioner-authored value left an optional enum field (e.g. `minimum_time_range`) unset, applied and read back at least once (so `prior.<Type>Config` is non-nil on the next read)
- WHEN a subsequent refresh or post-update read runs against Kibana 9.5.0 GA or later, which now returns a concrete default for that field instead of omitting it
- THEN the resource SHALL keep the field null in state because `prior.<Type>Config` is non-nil (same-type update), regardless of what `pm`'s own (zero-valued) field state would otherwise suggest
- AND the apply SHALL NOT report "Provider produced inconsistent result after apply"
- AND a subsequent plan SHALL show no changes
