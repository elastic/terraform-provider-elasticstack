## 1. Client layer

- [ ] 1.1 Add `GetAgentPolicyByName(ctx, client, name, spaceID string) (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, diag.Diagnostics)` to `internal/clients/fleet/agent_policy.go`, alongside the existing `GetAgentPolicy`.
- [ ] 1.2 Implement the escaped, quoted Kuery term: `fmt.Sprintf("ingest-agent-policies.name:%q", escape(name))` where `escape` replaces `\` with `\\` and `"` with `\"` before the `%q`-driven quoting (or build the quoted literal manually — confirm whether Go's `%q` escaping matches KQL's expected escaping for the characters that appear in practice; if not, write the escape explicitly).
- [ ] 1.3 Implement pagination of the *filtered* result set using `GetFleetAgentPoliciesWithResponse` with `Kuery` set and an explicit `SortField: "name"` (never `"id"` — confirmed rejected by the API).
- [ ] 1.4 Implement the anomaly-detecting pagination loop: track seen `Id`s (error on duplicate), error if a page is shorter than expected before reaching `Total`, error if `Total` changes between page requests.
- [ ] 1.5 Implement the client-side exact-match filter: from the filtered (and fully paginated) result set, keep only items where `item.Name == name`.
- [ ] 1.6 Return a sentinel/empty result (not an error) when zero exact matches are found, and all matches (not just the first) when more than one is found, so the caller can build the "no match" vs. "ambiguous match" diagnostics with the conflicting `policy_id`s.
- [ ] 1.7 Add unit tests in `internal/clients/fleet/agent_policy_test.go` against a mocked/httptest Kibana server:
  - [ ] 1.7.1 A single exact match on the first page.
  - [ ] 1.7.2 Two exact matches, one of which only appears on a later page (proves pagination is exhaustive before the ambiguity check runs).
  - [ ] 1.7.3 A filtered response containing similar-but-not-identical names, proving the client-side exact-match filter drops all non-exact matches.
  - [ ] 1.7.4 Zero items in the filtered response → sentinel "no match" result, not an error.
  - [ ] 1.7.5 An `Items` page that comes back short, or a `Total` that changes, before pagination completes → error, not a possibly-incomplete match.
  - [ ] 1.7.6 By-ID 404 (existing `GetAgentPolicy`, for parity) → `found=false`.
  - [ ] 1.7.7 A non-2xx/network error from either endpoint propagates as a diagnostic, not a silent empty result.
  - [ ] 1.7.8 The Kuery string sent on the wire is asserted against a known name containing a double quote and a backslash, confirming the escaping is applied before the request is sent.

## 2. Data source package `internal/fleet/agentpolicyds`

- [ ] 2.1 Create the package skeleton: `schema.go`, `models.go`, `read.go`, `data_source.go`, mirroring `internal/fleet/enrollmenttokens`.
- [ ] 2.2 `models.go`: define the model struct with `entitycore.KibanaConnectionField`, `ID`, `PolicyID` (Optional+Computed), `Name` (Optional+Computed), `SpaceID`, and every Computed attribute listed in the delta spec's schema table (`namespace`, `description`, `data_output_id`, `monitoring_output_id`, `fleet_server_host_id`, `download_source_id`, `monitor_logs`, `monitor_metrics`, `is_protected`, `host_name_format`, `supports_agentless`, `inactivity_timeout`, `unenrollment_timeout`, `global_data_tags`, `space_ids`, `required_versions`, `advanced_settings`, `advanced_monitoring_options`).
- [ ] 2.3 Implement `GetID() types.String` and `GetResourceID() types.String` (return `PolicyID` when set and non-empty, else the literal `"_"` placeholder, per `enrollmenttokens/models.go:40-45`).
- [ ] 2.4 Implement `GetSpaceID() types.String`.
- [ ] 2.5 `schema.go`: declare `policy_id` and `name` as `String`, `Optional: true`, `Computed: true`, each with `stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName(<other>))`. Declare `space_id` via the shared `kbschema.DataSourceSpaceIDAttribute()` helper (as `enrollmenttokens`/`outputds` already do). Declare every other attribute as `Computed: true`, using the same Terraform types as the resource's schema (`internal/fleet/agentpolicy/schema.go`) for `global_data_tags`, `advanced_settings`, `advanced_monitoring_options`, and the `customtypes.DurationType` fields, so the data source's output shape matches the resource's for users who reference both.
- [ ] 2.6 `read.go`: implement the selector branch — `config.PolicyID` set → call `fleet.GetAgentPolicy`; `config.Name` set → call `fleet.GetAgentPolicyByName`. Read the filter value from `config`, not the envelope-resolved resource ID (mirroring `enrollmenttokens/read.go:31-33`'s reasoning, since `"_"` is itself a valid `policy_id`).
- [ ] 2.7 `read.go`: on a by-name zero-match result, return `found=false` so the envelope's standard "not found" diagnostic fires (consistent with the by-ID 404 path), but with a message naming the `name` and `space_id` that were searched (`no agent policy found matching name %q in space %q`) rather than the generic `<component>_<name> "<resourceID>" not found` the envelope would otherwise produce with the `"_"` placeholder.
- [ ] 2.8 `read.go`: on a by-name multi-match result, return an error diagnostic (not `found=false`) listing every conflicting `policy_id`, e.g. `multiple agent policies match name %q in space %q: policy_id "a", policy_id "b"`.
- [ ] 2.9 `read.go`/`models.go`: implement the data-source-specific `populateFromAPI` mapping (new function, not a call into the resource's `agentPolicyModel.populateFromAPI`) — explicitly set `ID` and the unused selector (`PolicyID` or `Name`) from the API response, and map every other attribute directly from the raw API value with no null-preservation (see design.md Decision 5). Compute `monitor_logs`/`monitor_metrics` unconditionally from `MonitoringEnabled` containment.
- [ ] 2.10 `data_source.go`: wire the schema, model, and read function into `entitycore.NewKibanaDataSource[Model]`, following `internal/fleet/enrollmenttokens/data_source.go:26-35`.
- [ ] 2.11 Implement `GetVersionRequirements` returning no entries for the core lookup (per design.md Decision 6); confirm no version gate is needed for any `Computed`-only attribute exported here.

## 3. Registration and docs

- [ ] 3.1 Register `agentpolicyds.NewDataSource` in `provider/plugin_framework.go`'s `dataSources()` list, next to `enrollmenttokens.NewDataSource`, `outputds.NewDataSource`, `integrationds.NewDataSource`.
- [ ] 3.2 Add `examples/data-sources/elasticstack_fleet_agent_policy/data-source.tf` demonstrating lookup by `name` and by `policy_id`, including the `elasticstack_fleet_enrollment_tokens` pairing from the issue.
- [ ] 3.3 Generate/add `docs/data-sources/fleet_agent_policy.md` per the project's documentation generation workflow (`dev-docs/high-level/documentation.md`).

## 4. Acceptance tests (`internal/fleet/agentpolicyds/acc_test.go`)

- [ ] 4.1 Baseline fixture at the resource's agent-policy floor (`minVersionAgentPolicy = 8.6.0`, `internal/fleet/agentpolicy/acc_test.go:41`) with only ungated fields set; look it up by `policy_id`; assert each data-source attribute equals the raw API value.
- [ ] 4.2 Same baseline fixture, look it up by `name`; assert `policy_id` is populated and matches the fixture's resource ID.
- [ ] 4.3 A fixture with optional fields left unset at create time; assert the data source reports `description` as whatever the API actually returns (commonly empty string, not null) and reports `data_output_id`/`monitoring_output_id`/`fleet_server_host_id` as null only if the API itself returns null.
- [ ] 4.4 Neither `policy_id` nor `name` set → plan-time validation error from `ExactlyOneOf`.
- [ ] 4.5 Both `policy_id` and `name` set → plan-time validation error from `ExactlyOneOf`.
- [ ] 4.6 `name` matching zero policies → apply-time "not found" error.
- [ ] 4.7 `name` matching two fixtures sharing a name in the same space → apply-time error naming both conflicting `policy_id`s.
- [ ] 4.8 `space_id` omitted vs. explicit default vs. a non-default space, gated at 9.1.0 (mirroring `TestAccDataSourceEnrollmentTokensSpaceID`, `internal/fleet/enrollmenttokens/data_source_test.go:40,102`) — assert the lookup only matches within the given space, and that two fixtures with the same `name` in *different* spaces do not trigger the ambiguous-match error for a space-scoped `name` lookup.
- [ ] 4.9 Multi-step update: start with `monitor_logs`/`monitor_metrics` both `true` on the fixture resource, update the resource to `false` in a second step, re-read the data source, assert it now reports `false` (regression guard against a stale `true`).
- [ ] 4.10 Multi-step selector switch: look up the same fixture by `policy_id` in step one and by `name` in step two, confirming the now-populated computed `name`/`policy_id` values from step one don't trip `ExactlyOneOf` or cause a planning error in step two.
- [ ] 4.11 Version-gated steps for the advanced attributes (`global_data_tags` 8.15.0, `advanced_settings` 8.17.0/`advanced_settings.internal` 9.3.0, `advanced_monitoring_options` 8.16.0, `supports_agentless` 8.15.0, `inactivity_timeout`/`host_name_format=fqdn` 8.7.0, `unenrollment_timeout` 8.15.0, `required_versions`/`space_ids` 9.1.0), each skipped below its own floor via `versionutils.SkipIfUnsupported`, asserting the data source surfaces each fixture's configured value correctly. The `is_protected=true` fixture needs create-disabled → attach Elastic Defend integration (8.14.0 floor) → update-enabled steps, since `internal/fleet/agentpolicy/modify_plan.go:28-54` rejects creating a policy with `is_protected=true` directly.

## 5. OpenSpec

- [ ] 5.1 Confirm `openspec validate fleet-agent-policy-data-source --type change` passes with the delta spec under `openspec/changes/fleet-agent-policy-data-source/specs/fleet-agent-policy/spec.md`.
- [ ] 5.2 On archive, merge the delta `## ADDED Requirements` into `openspec/specs/fleet-agent-policy/spec.md`, including a `## Schema` addition documenting the new `data "elasticstack_fleet_agent_policy"` block alongside the existing resource schema.
