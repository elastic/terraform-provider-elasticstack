## Why

The `elasticstack_kibana_slo` resource has drifted from the current Kibana SLO API in several places, which leads to missing functionality, weaker plan-time validation, and at least one known incompatibility around KQL filter object variants. Capturing this now lets the provider support the current API surface more completely while tightening validation before more consumers depend on the current gaps.

## What Changes

- Add additive object-form KQL inputs with a `_kql` suffix for `filter`, `good`, and `total`, while preserving the existing string attributes for backward compatibility.
- Support the Kibana KQL union shape for those fields so the resource can send and read back both string and object forms.
- Improve provider-side validation for indicator-specific required and forbidden fields using conditional validators, moving failures from apply-time to plan-time where possible.
- Align simple schema validators with the API, including `slo_id` length, custom metric aggregation values, metric name format, and `time_window.type`.
- Add support for `settings.sync_field`.
- Add support for SLO `enabled` management, including write behavior if the API requires dedicated enable or disable operations rather than update-body fields.
- Add support for the `artifacts` field exposed by the SLO API.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-slo`: align the SLO resource requirements with the current Kibana SLO API for KQL input shape support, validation behavior, settings coverage, enabled-state management, and supported request fields.

## Impact

- Affected code is centered in `internal/kibana/slo/`, `internal/clients/kibanaoapi/slo.go`, and the conditional validator utilities under `internal/utils/validators/`.
- Affected interfaces include the Terraform schema for `elasticstack_kibana_slo`, SLO CRUD request construction, and acceptance and unit test coverage for SLO behavior.
- The change is intended to be backward compatible for existing string-based KQL configurations by introducing additive `_kql` inputs rather than replacing current attributes.
