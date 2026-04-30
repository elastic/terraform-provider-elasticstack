## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate uptime-monitor-status-validation-fix --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 In `internal/kibana/alertingrule/validate.go`, in `ruleTypeParamsSpecs`, change the `mustNewParamsSchemaSpecFromContainer` call for `"xpack.uptime.alerts.monitorStatus"` from `&kbapi.KibanaHTTPAPIsXpackSyntheticsAlertsMonitorstatusCreateRuleBodyAlerting{}` to `&kbapi.KibanaHTTPAPIsXpackUptimeAlertsMonitorstatusCreateRuleBodyAlerting{}`.
- [x] 2.2 In `internal/kibana/alertingrule/validate.go`, expand `legacyMonitorStatusParams.Filters` from its current single-field shape to include all four Kibana-accepted sub-fields:
  - `Tags *[]string \`json:"tags,omitempty"\`` (already present — keep)
  - `MonitorType *[]string \`json:"monitor.type,omitempty"\`` (add)
  - `ObserverGeoName *[]string \`json:"observer.geo.name,omitempty"\`` (add)
  - `URLPort *[]string \`json:"url.port,omitempty"\`` (add)

## 3. Testing

- [x] 3.1 Add or extend a unit test in `internal/kibana/alertingrule/` covering `validateRuleParams("xpack.uptime.alerts.monitorStatus", ...)` for a params object that includes both `availability` and `filters` with all four sub-fields — assert that validation returns no errors.
- [x] 3.2 Add a unit test asserting that params matching the existing Synthetics struct (e.g. containing `condition` or `monitorIds`) still fail validation for `"xpack.uptime.alerts.monitorStatus"` (regression guard to confirm we did not accidentally broaden acceptance to the wrong namespace).
- [x] 3.3 (Optional) If a `TestAcc` acceptance test exists for `xpack.uptime.alerts.monitorStatus`, add a config step that exercises `availability` + multi-field `filters` to confirm end-to-end plan/apply success.
  - **N/A**: No existing `TestAcc` acceptance test for this rule type.
