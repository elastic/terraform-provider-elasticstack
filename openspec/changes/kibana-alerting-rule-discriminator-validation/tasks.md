## 1. Discriminator validation core

- [ ] 1.1 Add `validateParamsViaDiscriminator(ruleTypeID string, params map[string]any) []string` in `internal/kibana/alertingrule/validate.go` that builds a stub `AlertingRuleAPIBody`, calls `ValueByDiscriminator()`, extracts `Params` via reflection, and re-decodes params with `DisallowUnknownFields()`
- [ ] 1.2 Wire required-keys heuristic (`computeRequiredKeys`, `missingRequiredKeys`) into the default discriminator path after successful strict decode
- [ ] 1.3 Refactor `validateRuleParams` to consult override table first, then fall back to discriminator path; preserve pass-through for unknown discriminators

## 2. Override table migration

- [ ] 2.1 Rename `ruleTypeParamsSpecs` to an override-focused map containing only: `logs.alert.document.count`, `xpack.uptime.alerts.monitorStatus`, `.es-query`, `.index-threshold`
- [ ] 2.2 Remove redundant override entries now covered by default path (APM, metrics, SLO, synthetics TLS where kbapi ID applies, etc.)
- [ ] 2.3 Remove incorrect `apm.rules.anomaly` entry; confirm `apm.anomaly` validates via default path
- [ ] 2.4 Resolve `xpack.uptime.alerts.tls` — remove override or remap to `xpack.synthetics.alerts.tls` / `xpack.uptime.alerts.tlsCertificate` per design open question

## 3. Tests

- [ ] 3.1 Add `TestDiscriminatorValidationCoversAllKbapiRuleTypes` that derives the `rule_type_id` list from `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` (for example by parsing `generated/kbapi/kibana.gen.go`) and asserts each is either default-validated or in the override table
- [ ] 3.2 Add fixture tests for `observability.rules.custom_threshold` (valid params pass, unknown key fails)
- [ ] 3.3 Add fixture tests for at least one stack monitoring rule type and `apm.anomaly`
- [ ] 3.4 Run existing `validate_test.go` fixtures and fix any regressions from stricter default validation
- [ ] 3.5 Add acceptance test for `observability.rules.custom_threshold` if CI Kibana supports it; otherwise document skip reason in test

## 4. Validation and docs

- [ ] 4.1 Run `make build` and targeted `go test ./internal/kibana/alertingrule/...`
- [ ] 4.2 Run `make check-openspec` (or `openspec validate kibana-alerting-rule-discriminator-validation --strict`)
- [ ] 4.3 Update resource docs or embedded description if supported rule types list should reference discriminator coverage (optional, only if maintainers want docs touch)
