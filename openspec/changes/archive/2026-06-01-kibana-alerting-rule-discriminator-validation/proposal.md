## Why

`elasticstack_kibana_alerting_rule` only validates `params` against a hand-maintained map of 12 rule types, while the generated `kbapi` client models 35 typed rule bodies via `AlertingRuleAPIBody.ValueByDiscriminator()`. The gap leaves 25 OpenAPI-known types (including `observability.rules.custom_threshold` from [terraform-provider-elasticstack#940](https://github.com/elastic/terraform-provider-elasticstack/issues/940)) without plan-time structural checks, and the manual map has already drifted (e.g. `apm.rules.anomaly` vs `apm.anomaly`). Replacing the map with runtime discriminator dispatch plus strict params re-decode keeps validation in sync with `kbapi` regeneration without maintaining parallel rule-type lists.

## What Changes

- Replace the static `ruleTypeParamsSpecs` map as the **default** params validation path with runtime dispatch through `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()`, followed by strict re-decode of the `Params` field using `json.Decoder.DisallowUnknownFields()`.
- Automatically cover all rule types present in the generated `ValueByDiscriminator()` switch (currently 35), including `observability.rules.custom_threshold`, stack monitoring rules, ML rules, and synthetics/uptime variants.
- Retain a **small override table** for rule types where OpenAPI ≠ Kibana runtime behavior or where params are nested unions requiring multi-variant tries (e.g. `logs.alert.document.count`, `xpack.uptime.alerts.monitorStatus` legacy shape, `.es-query` extra required keys, `.index-threshold` post-decode checks).
- Fix known ID mismatches in overrides (`apm.rules.anomaly` → `apm.anomaly`; reconcile `xpack.uptime.alerts.tls` with kbapi IDs).
- Add unit tests proving discriminator validation covers all `ValueByDiscriminator()` cases and that overrides still behave as today.
- Add at least one acceptance test for `observability.rules.custom_threshold` (or extend existing fixture-based validation tests if acc env lacks the feature).

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Extend REQ-018 plan-time `params` validation so all rule types known to `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` receive structural validation by default; document override behavior for union/legacy/Kibana-mismatch cases; unknown discriminators continue pass-through per existing unsupported-type semantics.

## Impact

- **Specs**: Delta under `openspec/changes/kibana-alerting-rule-discriminator-validation/specs/kibana-alerting-rule/spec.md`.
- **Implementation**: `internal/kibana/alertingrule/validate.go` (primary), `validate_test.go`, possibly small helper in `internal/kibana/alertingrule/` or `internal/clients/kibanaoapi/`.
- **Generated client**: Read-only use of existing `kbapi.AlertingRuleAPIBody`; no regeneration required.
- **Practitioner impact**: Stricter plan-time errors for previously pass-through rule types (e.g. typo'd params keys on `observability.rules.custom_threshold`); **not breaking** for valid configs that Kibana already accepts. Unsupported rule types outside the discriminator switch remain pass-through.
