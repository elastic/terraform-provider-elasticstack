## MODIFIED Requirements

### Requirement: Plan-time `params` validation (REQ-018–REQ-021)

When both `params` and `rule_type_id` are known during configuration validation (before apply), the resource SHALL parse `params` as JSON. If parsing fails, it SHALL report an attribute error on `params` with summary `Invalid params JSON`.

For **supported** rule types, the resource SHALL verify that `params` is a JSON object that matches the encoded params shape for that rule type: required fields for that rule type MUST be present, property names outside the allowed set MUST be rejected, and where a rule type allows more than one payload shape (for example DSL vs KQL vs ESQL for the same `rule_type_id`, or log-threshold params variants), the value MUST match exactly one of those shapes. The provider MAY allow specific extra property names when Kibana accepts them but the encoded shape does not yet list them. If validation fails, the resource SHALL report an attribute error on `params` whose summary is `Invalid params for rule_type_id "<id>"` and whose detail explains what was wrong.

A rule type SHALL be treated as **supported** when either:

1. Its `rule_type_id` is handled by `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` (the generated OpenAPI discriminated union), in which case the default validator SHALL:
   - construct a minimal stub alerting rule body containing the practitioner's `rule_type_id` and `params` plus stub values for other required rule fields (`name`, `consumer`, `schedule.interval`) sufficient for union dispatch;
   - dispatch via `ValueByDiscriminator()` to the typed rule body;
   - extract the typed `Params` value and re-decode the practitioner's params JSON into that params type using `json.Decoder.DisallowUnknownFields()`;
   - apply the required-keys heuristic (including any per-type additional required or allowed key patches registered for that rule type); and
   - run any registered post-decode checks for that rule type;
   or

2. Its `rule_type_id` is listed in the provider's explicit **params validation override** table for cases where OpenAPI does not match Kibana runtime behavior or where params are nested unions requiring multiple variant attempts (for example `logs.alert.document.count`, `xpack.uptime.alerts.monitorStatus`, `.es-query`, `.index-threshold`).

For **unsupported** rule types (any `rule_type_id` that is neither handled by `ValueByDiscriminator()` nor listed in the override table), the resource SHALL perform no structural check of `params` beyond JSON syntax when known; compatibility with Kibana is left to the API.

At apply time, if `params` cannot be decoded as JSON for the request, the resource SHALL surface a diagnostic rather than calling the API with invalid JSON. Structural rules checked at plan time for supported types MUST NOT need to be re-stated as duplicate diagnostics on a normal successful plan.

#### Scenario: Invalid JSON

- GIVEN known `params` that are not valid JSON
- WHEN configuration is validated before apply
- THEN the provider SHALL report an error on `params` with `Invalid params JSON`

#### Scenario: Unsupported rule type passes structural checks

- GIVEN a `rule_type_id` that is not handled by `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` and is not listed in the params validation override table
- WHEN configuration is validated before apply
- THEN the provider SHALL not reject `params` for unknown property names or missing keys solely because the rule type is outside the provider's supported-params set

#### Scenario: Supported rule type with wrong shape

- GIVEN a supported `rule_type_id` and `params` that are valid JSON but omit a required key or include a disallowed key
- WHEN configuration is validated before apply
- THEN the provider SHALL report an error on `params` referencing that `rule_type_id` and describing the validation failure

#### Scenario: Discriminator-known rule type validates via default path

- GIVEN a `rule_type_id` handled by `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` that is not in the override table (for example `observability.rules.custom_threshold`)
- AND `params` that are valid JSON matching the generated params type for that rule type
- WHEN configuration is validated before apply
- THEN the provider SHALL NOT report a params validation error

#### Scenario: Discriminator-known rule type rejects unknown params keys

- GIVEN a `rule_type_id` handled by `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` that is not in the override table
- AND `params` that are valid JSON but include a property name not defined on the generated params type for that rule type
- WHEN configuration is validated before apply
- THEN the provider SHALL report an error on `params` with summary `Invalid params for rule_type_id "<id>"`

## ADDED Requirements

### Requirement: Discriminator validation coverage guard (REQ-051)

The provider SHALL maintain automated tests ensuring every `rule_type_id` case in `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` is either validated by the default discriminator params path or explicitly listed in the params validation override table with documented rationale.

#### Scenario: New kbapi rule type without validation decision

- GIVEN `kbapi` regeneration adds a new case to `ValueByDiscriminator()`
- AND that `rule_type_id` is neither covered by the default path in tests nor listed in the override table
- WHEN unit tests run
- THEN the coverage guard test SHALL fail until the new rule type is handled

### Requirement: Params validation override table (REQ-052)

The provider SHALL keep an explicit override table for `rule_type_id` values whose params validation cannot be satisfied by the default discriminator path alone. Overrides SHALL preserve existing behavior for:

- `logs.alert.document.count` — multiple params union variants;
- `xpack.uptime.alerts.monitorStatus` — Uptime generated struct plus legacy fallback shape (REQ-043, REQ-044);
- `.es-query` — additional required `size` key beyond OpenAPI optional marking;
- `.index-threshold` — post-decode check that `index` is an array of strings when present.

The override table SHALL NOT include incorrect rule type IDs that do not exist in `kbapi` (for example `apm.rules.anomaly`).

#### Scenario: Override takes precedence over default path

- GIVEN `rule_type_id = "logs.alert.document.count"`
- WHEN configuration is validated before apply
- THEN the provider SHALL use the override multi-variant validation logic rather than only the default discriminator params re-decode

#### Scenario: APM anomaly uses kbapi rule type id

- GIVEN `rule_type_id = "apm.anomaly"`
- AND params matching the generated APM anomaly params shape
- WHEN configuration is validated before apply
- THEN the provider SHALL validate via the default discriminator path and SHALL NOT require the obsolete id `apm.rules.anomaly`
