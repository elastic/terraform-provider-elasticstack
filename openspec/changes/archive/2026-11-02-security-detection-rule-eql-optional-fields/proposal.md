## Why

The Kibana detection rules API accepts three optional EQL settings: `event_category_override`, `tiebreaker_field`, and `timestamp_field`. The Kibana rule editor offers all three under EQL settings. `elasticstack_kibana_security_detection_rule` supports only `tiebreaker_field`.

Without `event_category_override`, an EQL query whose `<category> where` head is not an `event.category` value matches nothing. For example, `info where event.action == "ExampleAction"` is evaluated against `event.category`, and `info` is an `event.type` value, not an ECS category. Kibana accepts the rule, every execution succeeds, and the rule never alerts. Plan, apply, and the rule's execution status all look healthy. Fifteen Elastic prebuilt rules in `elastic/detection-rules` (AWS CloudTrail and Okta integrations) set `event_category_override`, so they cannot be reproduced with this resource today.

Rule update is a full replace that deletes every field the request does not send. A value for either field set outside Terraform (in the Kibana UI, or on an imported rule) is therefore removed on the next apply of that rule, and the plan does not show it.

## What Changes

- `elasticstack_kibana_security_detection_rule` gains two optional EQL attributes:
  - `event_category_override`: the field the EQL query's event category head is matched against. Kibana uses `event.category` when it is unset.
  - `timestamp_field`: the field used to sort events in a sequence. Kibana uses `@timestamp` when it is unset. It is not the same as `timestamp_override`.
- Both attributes behave like the existing `tiebreaker_field`:
  - They are sent on create and update of EQL rules when set.
  - Removing one from configuration clears it in Kibana.
  - Read and import populate them from Kibana; an unset value is null.
- A value for either field set outside Terraform shows up as drift in the plan instead of being silently removed.

Non-goals:

- Rejecting EQL-only attributes on other rule types. That would also cover `tiebreaker_field` and is a separate change.
- Changes to `tiebreaker_field` or `timestamp_override`.
- Defaults, plan modifiers, or forced replacement for the new attributes.
- A minimum stack version gate for the new attributes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-security-detection-rule`: the schema gains `event_category_override` and `timestamp_field` among the EQL-specific attributes, and a requirement defines how `event_category_override`, `tiebreaker_field`, and `timestamp_field` map between Terraform and the API on create, update, read, and import.

## Impact

- `internal/kibana/security_detection_rule/`: schema, EQL create/update/read mapping, type-specific defaults, unit and acceptance tests, acceptance testdata.
- Generated resource documentation for `elasticstack_kibana_security_detection_rule`.
- `openspec/specs/kibana-security-detection-rule/spec.md`: EQL schema block and the new mapping requirement.
- Users: no migration. Additive optional attributes; existing configurations and state are unaffected. Rules that carry either value set outside Terraform will show it as drift on the next plan.
