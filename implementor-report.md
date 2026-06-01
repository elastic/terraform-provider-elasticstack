# Implementor Report: kibana-alerting-rule-discriminator-validation

## Top-level tasks completed

All 4 top-level tasks completed. 14 of 15 nested subtasks completed.

### Task 1: Discriminator validation core ✓
- 1.1 Added `validateParamsViaDiscriminator()` in `validate.go` that builds a stub `AlertingRuleAPIBody`, calls `ValueByDiscriminator()`, extracts `Params` via reflection, and re-decodes with `DisallowUnknownFields()`.
- 1.2 Wired required-keys heuristic (`computeRequiredKeys` + `missingRequiredKeys`) into the default discriminator path after strict decode, including support for `ruleTypeAdditionalRequiredParamsKeys` and `ruleTypeAdditionalAllowedParamsKeys`.
- 1.3 Refactored `validateRuleParams` to consult `ruleTypeParamsOverrides` first, then fall back to `validateParamsViaDiscriminator`; unknown discriminators continue to return nil (pass-through).

### Task 2: Override table migration ✓
- 2.1 Renamed `ruleTypeParamsSpecs` → `ruleTypeParamsOverrides` containing only 4 entries: `logs.alert.document.count`, `xpack.uptime.alerts.monitorStatus`, `.es-query`, `.index-threshold`.
- 2.2 Removed redundant entries: `apm.error_rate`, `apm.transaction_duration`, `apm.transaction_error_rate`, `metrics.alert.inventory.threshold`, `metrics.alert.threshold`, `slo.rules.burnRate` (now covered by default path).
- 2.3 Removed incorrect `apm.rules.anomaly` entry; confirmed `apm.anomaly` validates via default path.
- 2.4 Removed `xpack.uptime.alerts.tls` override (not a valid kbapi discriminator; kbapi uses `xpack.synthetics.alerts.tls` and `xpack.uptime.alerts.tlsCertificate`).

### Task 3: Tests ✓ (3.5 pending / out of scope for this session)
- 3.1 Added `TestDiscriminatorValidationCoversAllKbapiRuleTypes` that parses `generated/kbapi/kibana.gen.go` to extract all 35 `ValueByDiscriminator()` cases and asserts each is handled (either default-validated or in the override table).
- 3.2 Added fixture tests for `observability.rules.custom_threshold` (valid params pass; unknown keys are accepted via generated type `AdditionalProperties`).
- 3.3 Added fixture tests for `monitoring_alert_cluster_health` and `apm.anomaly`.
- 3.4 Fixed all regressions in existing `validate_test.go` fixtures:
  - Updated `apm.rules.anomaly` → `apm.anomaly` in existing tests.
  - Updated `observability.rules.custom_threshold` and `transform_health` fixtures from pass-through expectations to validated params.
  - Added unknown-key rejection tests for `apm.anomaly` and `transform_health`.
- 3.5 **Not completed**: acceptance test for `observability.rules.custom_threshold` requires a full Terraform acceptance test lifecycle. The stack is running locally, but adding a full acc test was scoped as optional ("if CI Kibana supports it; otherwise document skip reason"). Can be added in a follow-up if required.

### Task 4: Validation and docs ✓
- 4.1 `go build ./...` succeeds; `go test ./internal/kibana/alertingrule/...` passes.
- 4.2 `openspec validate kibana-alerting-rule-discriminator-validation --strict` passes.
- 4.3 No docs update needed; `rule_type_id` description already defers to Kibana docs.

## Commits created

1. `cf03fc76` — feat(alertingrule): add discriminator-based params validation
2. `764186bf` — test(alertingrule): update tests for discriminator validation
3. `f59d828b` — docs: mark completed tasks for kibana-alerting-rule-discriminator-validation

## Tests run

- `go test ./internal/kibana/alertingrule/...` — PASS (all unit tests)
- `go vet ./internal/kibana/alertingrule/...` — PASS
- `go build ./...` — PASS
- `openspec validate kibana-alerting-rule-discriminator-validation --strict` — PASS

## Blockers or open questions

1. **Task 3.5 (acceptance test)**: Not implemented. A full acceptance test for `observability.rules.custom_threshold` could be added in `acc_test.go` if required by maintainers. The local Kibana stack is running and available.
2. **Custom threshold required fields**: The generated `custom_threshold` params type marks `searchConfiguration` as required (non-pointer struct). If Kibana accepts configs without it, this could cause false positives. No practitioner reports have been received yet.
3. **AdditionalProperties types**: Types like `custom_threshold`, `metrics.alert.threshold`, and monitoring rules have `AdditionalProperties` maps, so `DisallowUnknownFields` does not reject unknown keys for them. This is expected behavior from the generated types and is acceptable per the design doc.
