## Context

`internal/kibana/alertingrule/validate.go` holds `ruleTypeParamsSpecs`, a map from `rule_type_id` to a slice of `paramsSchemaSpec`. Each spec wraps a Go struct used to validate `params` JSON at plan time: the decoder is created with `DisallowUnknownFields()`, so any key in the practitioner's `params` that is not a field in the target struct causes validation to fail.

For `"xpack.uptime.alerts.monitorStatus"` two specs are registered:

1. A primary spec produced by `mustNewParamsSchemaSpecFromContainer`, which extracts the `Params` field from the container struct and validates against it.
2. A legacy fallback produced by `mustNewParamsSchemaSpec` using the handwritten `legacyMonitorStatusParams` type.

Both are broken today.

## Goals

- Fix the primary spec to reference the correct Uptime-namespace generated struct so `availability`, the union-type `filters`, and other uptime-specific fields are accepted.
- Expand the legacy fallback `legacyMonitorStatusParams.Filters` to include all four sub-fields that Kibana accepts.
- Do not alter any other rule-type validation, schema, or API path.

## Non-Goals

- Changing any resource schema attribute.
- Adding new Terraform configuration options.
- Modifying create/update/read API paths or state mapping.
- Adding acceptance tests beyond what the existing `xpack.uptime.alerts.monitorStatus` suite already covers (optional regression test welcome but not required by this fix).

## Decisions

| Topic | Decision |
|-------|----------|
| Primary struct | Change `KibanaHTTPAPIsXpackSyntheticsAlertsMonitorstatusCreateRuleBodyAlerting` → `KibanaHTTPAPIsXpackUptimeAlertsMonitorstatusCreateRuleBodyAlerting` in the `mustNewParamsSchemaSpecFromContainer` call. This struct already exists in `generated/kbapi/kibana.gen.go` and its `Params` field models `availability`, `filters` (union), `numTimes`, `shouldCheckAvailability`, `shouldCheckStatus`, etc. |
| Filters union type | The Uptime struct's `Filters` field is `*KibanaHTTPAPIsXpackUptimeAlertsMonitorstatusCreateRuleBodyAlerting_Params_Filters`, a union type backed by `json.RawMessage`. Union types implement custom `UnmarshalJSON`, so `DisallowUnknownFields()` on the outer struct does not propagate into the union — it accepts both a plain string and an object form, which matches Kibana's behavior. |
| Legacy fallback | Add `MonitorType`, `ObserverGeoName`, and `UrlPort` fields (each `*[]string`) to `legacyMonitorStatusParams.Filters`. Field JSON keys must use the Kibana-format dot-notation keys (`"monitor.type"`, `"observer.geo.name"`, `"url.port"`) with `omitempty`. `Tags` already present; keep it unchanged. |
| Required-keys heuristic | `legacyMonitorStatusParams.Filters` is a pointer field — it remains omitable. The new sub-fields inside `Filters` are also pointer types tagged `omitempty`, so they will not be inferred as "required" by `computeRequiredKeys`. No changes needed to `ruleTypeAdditionalRequiredParamsKeys`. |

## Risks / Trade-offs

- Switching to the Uptime struct means the union `Filters` field no longer rejects unknown keys via `DisallowUnknownFields()` (the union swallows everything). This is intentional and matches Kibana's own permissive handling of the `filters` union — the generated struct already models this design choice.
- The legacy fallback remains in place for callers using older params shapes. After the fix, the primary Uptime struct handles the canonical modern shape; the legacy struct catches remaining legacy payload variants.

## Open Questions

- None. Both fixes are fully specified by the existing generated struct and the Kibana OpenAPI spec.
