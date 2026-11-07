# Delta spec: `kibana-security-detection-rule` — EQL optional fields

Extends the main spec at `openspec/specs/kibana-security-detection-rule/spec.md`.

## Schema refinements

Add two attributes to the EQL-specific part of the canonical schema, next to `tiebreaker_field`:

```hcl
  # EQL-specific
  tiebreaker_field        = <optional, string>
  event_category_override = <optional, string> # field the EQL event category head is matched against; Kibana uses event.category when unset
  timestamp_field         = <optional, string> # field used to sort events in a sequence; Kibana uses @timestamp when unset; not timestamp_override
```

## ADDED Requirements

### Requirement: Mapping — EQL optional fields (REQ-037)

For rules with `type = "eql"`, the resource SHALL map the EQL optional fields `event_category_override`, `tiebreaker_field`, and `timestamp_field` between Terraform and the Kibana detection rules API in the same way. Each is an optional string attribute with no default; none forces replacement.

- On create and update, the resource SHALL send each field when its configured value is known, and SHALL omit it when the value is null.
- Because update replaces the whole rule, removing one of these attributes from configuration SHALL clear the field in Kibana, and the rule SHALL be updated in place.
- On read and import, the resource SHALL set each attribute from the API response, and SHALL set it to null when the response does not include the field.
- A field that Kibana holds but the configuration does not set SHALL appear in state after refresh, so that the next plan shows the change instead of silently removing the field.
- For rule types other than `eql`, the resource SHALL NOT send these fields to the API. State keeps the configured value, as it does for `tiebreaker_field` today; rejecting these attributes on other rule types is out of scope.

#### Scenario: Create an EQL rule with event_category_override and timestamp_field

- GIVEN an EQL rule configuration that sets `event_category_override = "event.type"` and `timestamp_field = "event.ingested"`
- WHEN the resource is created
- THEN Kibana SHALL store both values on the rule
- AND state SHALL hold both values
- AND a subsequent plan on the unchanged configuration SHALL show no changes

#### Scenario: Change the EQL optional fields

- GIVEN an existing EQL rule with `event_category_override = "event.type"` and `timestamp_field = "event.ingested"`
- WHEN the configuration changes them to `event_category_override = "event.kind"` and `timestamp_field = "@timestamp"` and is applied
- THEN Kibana SHALL store the new values
- AND the rule SHALL be updated in place, not replaced

#### Scenario: Remove the EQL optional fields

- GIVEN an existing EQL rule that sets `event_category_override` and `timestamp_field`
- WHEN both attributes are removed from the configuration and it is applied
- THEN Kibana SHALL no longer hold either field on the rule
- AND state SHALL hold null for both attributes
- AND the rule SHALL be updated in place, not replaced

#### Scenario: EQL rule without the optional fields

- GIVEN an EQL rule configuration that sets neither `event_category_override` nor `timestamp_field`
- WHEN the resource is created
- THEN the create request SHALL NOT include either field
- AND state SHALL hold null for both attributes
- AND a subsequent plan on the unchanged configuration SHALL show no changes

#### Scenario: Import an EQL rule with the optional fields

- GIVEN an EQL rule in Kibana that has `event_category_override` and `timestamp_field` set
- WHEN the rule is imported
- THEN state SHALL hold both values from Kibana

#### Scenario: Field set outside Terraform appears as drift

- GIVEN an EQL rule managed by Terraform whose configuration does not set `event_category_override`
- AND `event_category_override` was set on the rule outside Terraform
- WHEN the resource is refreshed
- THEN state SHALL hold the value from Kibana
- AND the next plan SHALL show an update that removes it

#### Scenario: Non-EQL rule type does not send EQL optional fields

- GIVEN a rule configuration with `type = "query"` that sets `event_category_override`
- WHEN the resource builds its create or update request
- THEN the request SHALL NOT include `event_category_override`
