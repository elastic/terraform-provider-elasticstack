## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-security-detection-rule-alerts-filter --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, sync delta into `openspec/specs/kibana-security-detection-rule/spec.md` and archive the change per project workflow.

## 2. Schema and models

- [ ] 2.1 In `internal/kibana/security_detection_rule/schema.go`: replace `schema.MapAttribute{ElementType: types.StringType}` for `alerts_filter` with a `schema.SingleNestedBlock` containing a required `query` sub-block (`kql` optional string, `filters_json` optional `jsontypes.Normalized`) and an optional `timeframe` sub-block (`days` list int64, `timezone` string, `hours_start` string, `hours_end` string). Add `objectvalidator.AlsoRequires` validators on `timeframe` so all four attributes are required when the block is present.
- [ ] 2.2 In `internal/kibana/security_detection_rule/models.go`: replace `AlertsFilter types.Map` on `ActionModel` with `AlertsFilter types.Object`. Add `ActionAlertsFilterModel` struct (`Query types.Object`, `Timeframe types.Object`) and `ActionAlertsFilterQueryModel` struct (`Kql types.String`, `FiltersJSON jsontypes.Normalized`). Add `ActionAlertsFilterTimeframeModel` struct (`Days types.List`, `Timezone types.String`, `HoursStart types.String`, `HoursEnd types.String`). Add `getAlertsFilterAttrTypes()`, `getAlertsFilterQueryAttrTypes()`, and `getAlertsFilterTimeframeAttrTypes()` helper functions that derive types from the schema (follow the `sync.Once` caching pattern used in `alertingrule/schema.go`).
- [ ] 2.3 Bump `Schema.Version` from 1 to 2 in `internal/kibana/security_detection_rule/resource.go` (or wherever the schema version is defined). Add a `StateUpgraders` entry (version 1 → 2) with a no-op function that returns state unchanged (the old `MapAttribute` state was non-functional; no data survives migration).

## 3. Read path

- [ ] 3.1 In `internal/kibana/security_detection_rule/models_from_api_type_utils.go`: replace the `alerts_filter` read logic (lines 194–205) with a new `flattenActionAlertsFilter` helper. The helper SHALL:
  - Accept `*kbapi.SecurityDetectionsAPIRuleActionAlertsFilter` (a `map[string]interface{}`).
  - Extract the `"query"` key as a `map[string]interface{}`, then extract `"kql"` (string) and `"filters"` (array), marshal `filters` back to a normalized JSON string for `filters_json`.
  - Extract the `"timeframe"` key as a `map[string]interface{}` and map to `ActionAlertsFilterTimeframeModel`.
  - Return `types.ObjectNull(getAlertsFilterAttrTypes())` when the input is nil.

## 4. Write path

- [ ] 4.1 In `internal/kibana/security_detection_rule/models_to_api_type_utils.go`: replace the `alerts_filter` write logic (lines 559–570) with a new `expandActionAlertsFilter` helper. The helper SHALL:
  - Accept the `types.Object` value from `ActionModel.AlertsFilter`.
  - Unmarshal `filters_json` into `[]interface{}`.
  - Build `map[string]interface{}{"query": map[string]interface{}{"kql": kqlValue, "filters": filtersSlice}}`.
  - When `timeframe` is present, add `"timeframe": map[string]interface{}{"days": daysSlice, "timezone": tz, "hours": map[string]interface{}{"start": start, "end": end}}` — using the Kibana API's `hours.start`/`hours.end` nested key shape.
  - Cast the assembled map to `kbapi.SecurityDetectionsAPIRuleActionAlertsFilter` and return a pointer.

## 5. Documentation and descriptions

- [ ] 5.1 Update the embedded resource description (`internal/kibana/security_detection_rule/resource_description.md` or equivalent) to document the new `alerts_filter` nested block and its sub-attributes, including the `filters_json` attribute name and a `jsonencode([])` example.
- [ ] 5.2 Run `make generate-docs` (or equivalent) and commit the regenerated documentation under `docs/`.

## 6. Acceptance tests

- [ ] 6.1 Add or update an acceptance test in `internal/kibana/security_detection_rule/` that:
  - Creates a detection rule with an `actions` block that includes `alerts_filter.query.kql` and `alerts_filter.query.filters_json = jsonencode([])`.
  - Asserts that `terraform plan` after apply shows no diff (round-trip correctness).
  - Updates the `kql` value and asserts the update succeeds.
- [ ] 6.2 Add an acceptance test step that includes an `alerts_filter.timeframe` block and asserts all four timeframe fields round-trip correctly.
- [ ] 6.3 Ensure that rules without `alerts_filter` continue to create/update/read without errors (regression coverage).
