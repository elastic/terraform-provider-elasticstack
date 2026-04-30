## Why

Plan-time validation for `elasticstack_kibana_alerting_rule` with `rule_type_id = "xpack.uptime.alerts.monitorStatus"` rejects valid `params` that Kibana accepts. Rules that use `availability` or `filters` (with any field beyond `tags`) cannot be planned or applied ([#2570](https://github.com/elastic/terraform-provider-elasticstack/issues/2570)). Two bugs in `internal/kibana/alertingrule/validate.go` cause this failure.

## What Changes

Two targeted fixes to `internal/kibana/alertingrule/validate.go`:

1. **Wrong generated struct** — `ruleTypeParamsSpecs` for `"xpack.uptime.alerts.monitorStatus"` references `KibanaHTTPAPIsXpackSyntheticsAlertsMonitorstatusCreateRuleBodyAlerting` (the **Synthetics** namespace struct). The correct struct is `KibanaHTTPAPIsXpackUptimeAlertsMonitorstatusCreateRuleBodyAlerting` (the **Uptime** namespace struct), which already exists in `generated/kbapi/kibana.gen.go` and models both `Availability` and a union-type `Filters` covering all four filter sub-fields.

2. **Incomplete legacy fallback** — `legacyMonitorStatusParams.Filters` only declares a `Tags` field. The Kibana OpenAPI spec and `@kbn/response-ops-rule-params` schema allow four filter sub-fields (`monitor.type`, `observer.geo.name`, `tags`, `url.port`). Because `DisallowUnknownFields()` is used, any of the three missing sub-fields causes the legacy fallback to fail validation.

**Out of scope for this proposal**: changes to the resource schema, the create/update/read API path, state mapping, or any feature gating — this is a bug fix to the plan-time params validator only.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Fix plan-time `params` validation for `xpack.uptime.alerts.monitorStatus` (REQ-043–REQ-044).

## Impact

- **Specs**: Delta under `openspec/changes/uptime-monitor-status-validation-fix/specs/kibana-alerting-rule/spec.md`.
- **Implementation** (future): `internal/kibana/alertingrule/validate.go` — two surgical changes; no schema, no API mapping, no docs changes required.
