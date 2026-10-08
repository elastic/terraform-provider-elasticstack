## MODIFIED Requirements

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009)

When Kibana omits or defaults fields on read, the resource SHALL preserve prior Terraform intent to avoid inconsistent results and spurious drift where the implementation supports that behavior. The resource preserves the prior `time_range.mode` value already held in state or plan instead of overwriting it from read-back when the GET response does not supply a usable mode. When the GET dashboard API does not supply a usable `access_control.access_mode` value, the resource SHALL clear `access_control` in Terraform state rather than leaving a stale prior value behind. When the options block was omitted in Terraform and Kibana materializes only the default dashboard options matching the implementation's `isDashboardOptionsDefaultSet` helper (including `auto_apply_filters` and `hide_panel_borders` at their API defaults when applicable), the resource SHALL keep the `options` block null in state. When a section's prior `collapsed` value was null and Kibana returns `false`, the resource SHALL preserve null rather than forcing `false` into state.

The root `time_range`, `refresh_interval`, and `query` blocks are not subject to null-preservation or default detection. Kibana applies no defaults to these blocks and returns them as absent when they were not sent (verified against Kibana 9.4.2 and 9.6.0-SNAPSHOT), so the resource SHALL map them from the API response as-is: a block present in the response is populated in state, and a block absent from the response is null in state. The existing `time_range.mode` preservation continues to apply when `time_range` is present. If a future Kibana version begins returning defaults for these blocks, the title-only acceptance test is the intended detection mechanism and a follow-up change would add null-preservation.

For panel reads, the provider SHALL seed each panel from prior practitioner intent before finalizing state: from the prior plan on the post-create and post-update read-back, and from prior state on refresh. After that seed, it SHALL apply panel-type-specific alignment so Kibana-injected defaults or omitted optional values do not overwrite practitioner intent. This alignment includes preserving configured titles and descriptions when the API returns blank values, preserving ES|QL control `esql_query`, `title`, and `available_options` when the API omits them, preserving raw `config_json` when the read-back only differs by omitted optional `filters` or `query` keys, and preserving semantically equivalent optional JSON defaults such as `rank_by` in metric and tagcloud configurations.

For typed panel config blocks whose `PopulateFromAPI` receives both `pm` (the panel model being built, which callers always pass zero-valued to avoid aliasing plan pointers) and `prior` (the prior plan or state panel at the same index), the null-preservation decision for that block SHALL key on `prior.<Type>Config`, not on `pm`'s own field: `prior.<Type>Config != nil` SHALL be treated as a same-type update (honor the practitioner's null intent for optional fields), and `prior.<Type>Config == nil` SHALL be treated as creation, import, or a genuine type change: there is no prior null intent to honor, so the block SHALL be rebuilt from the API. For config blocks that are optional even when the panel type matches (`synthetics_monitors_config`, `synthetics_stats_overview_config`), the block SHALL instead be left null when the API response carries no content for it, rather than materializing an empty block. `pm`'s own field state SHALL NOT be used for this decision, since it never carries prior intent into `PopulateFromAPI`.

As of Kibana 9.5.0 GA, several typed panel config blocks' optional enum-shaped fields (for example `aiops_pattern_analysis_config.minimum_time_range` and `.random_sampler_mode`) receive concrete server-side default values on read where earlier Kibana versions returned no value at all. The provider SHALL apply REQ-009 null-preservation to these fields exactly as it does for any other Kibana-injected default: when the practitioner left the field unset (null in prior state/plan), the resource SHALL keep it null in state even though the API now returns a concrete default, rather than materializing that default and producing "Provider produced inconsistent result after apply".

The resource models only the currently supported Terraform subset of dashboard fields. Fields present in the Kibana dashboard API but not modeled by this resource — for example top-level `project_routing` — are outside this resource contract (see REQ-037 for `filters` and REQ-038 for `pinned_panels`).

The provider SHALL treat an API-returned `""` for `description` as semantically equivalent to an omitted field when prior plan/state had `description` null, restoring null in state rather than propagating the API-echoed empty string. This is an instance of REQ-009 null-preservation applied to the dashboard root `description`. This SHALL be consistent with the null/empty-string normalization already applied to XY chart `fitting.type`, `fitting.end_value`, and panel-level `time_range`.

The provider SHALL apply the same intent-preserving treatment to the root-level `tags` attribute. When the Kibana API returns a nil or empty `tags` value and the prior Terraform plan/state had `tags` as a known value — including a known-empty list (`tags = []`) — the provider SHALL preserve that prior known value in state rather than forcing `null`. When the prior plan/state `tags` value is itself null (the practitioner did not set `tags`), the provider SHALL preserve `null`. When the API returns a non-empty `tags` array, the provider SHALL store that value from the API response, overwriting any prior value. When there is no prior plan/state value to consult (for example on `terraform import`, where `tags` is `Unknown`), the provider SHALL map a nil-or-empty API `tags` value to `null`. This fixes the error reported when a practitioner sets `tags = []`: without this normalization, the provider produces an unexpected new value, `.tags: was cty.ListValEmpty(cty.String), but now null`.

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

#### Scenario: Known-empty tags normalizes to empty list on read, not null

- GIVEN a dashboard configured with `tags = []` (known-empty list in Terraform plan/state)
- AND the Kibana API returns a nil or empty `tags` value
- WHEN the provider reads the dashboard
- THEN state SHALL contain `tags = []`
- AND a subsequent plan SHALL show no changes

#### Scenario: Omitted tags stays null on read

- GIVEN a dashboard configured without `tags` (null in Terraform plan/state)
- AND the Kibana API returns a nil or empty `tags` value
- WHEN the provider reads the dashboard
- THEN state SHALL contain `tags = null`

#### Scenario: Non-empty tags preserved unchanged

- GIVEN a dashboard configured with `tags = ["a", "b"]`
- AND the Kibana API returns `tags: ["a", "b"]`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `tags = ["a", "b"]`

#### Scenario: Import with no prior tags intent maps nil/empty API tags to null

- GIVEN a dashboard is imported via `terraform import` (no prior plan/state `tags` value; `tags` is Unknown)
- AND the Kibana API returns a nil or empty `tags` value
- WHEN the provider reads the dashboard
- THEN state SHALL contain `tags = null`
