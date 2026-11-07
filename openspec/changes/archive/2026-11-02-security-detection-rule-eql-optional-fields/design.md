## Context

`elasticstack_kibana_security_detection_rule` is one resource for all rule types. Type-specific attributes live at the top level of the schema and are dispatched per type through rule processors. EQL rules go through `EqlRuleProcessor` in `internal/kibana/security_detection_rule/models_eql.go`, which builds the create and update request bodies and maps the API response back into the model.

The generated client already models all three EQL optional fields: `SecurityDetectionsAPIEqlOptionalFields` in `generated/kbapi/kibana.gen.go` carries `EventCategoryOverride`, `TiebreakerField`, and `TimestampField` as optional string pointers, and they are present on the EQL create props, update props, and rule response types. The resource maps only `TiebreakerField`:

- `Data.TiebreakerField` in `models.go`, listed under the EQL-specific fields
- an optional `tiebreaker_field` string attribute in `schema.go`, inline `MarkdownDescription`, no validators, no plan modifiers
- `toEqlRuleCreateProps` and `toEqlRuleUpdateProps` set it when the value is known
- `updateFromEqlRule` reads it with `typeutils.StringishPointerValue`
- `initializeTypeSpecificFieldsToDefaults` sets it to null when unknown

Rule update uses `PUT /api/detection_engine/rules`, which replaces the whole rule. Any field the request omits is deleted.

## Goals / Non-Goals

**Goals:**

- Expose `event_category_override` and `timestamp_field` on EQL rules with the same behavior as `tiebreaker_field`.
- Make a value for either field set outside Terraform visible as drift instead of silently removing it on the next update.

**Non-Goals:**

- Validation that rejects EQL-only attributes on other rule types.
- Changes to `tiebreaker_field` or `timestamp_override`.
- Defaults, plan modifiers, `RequiresReplace`, or a version gate for the new attributes.
- Grouping the EQL settings into a nested block.

## Decisions

**Two top-level attributes that mirror `tiebreaker_field` exactly.** `event_category_override` and `timestamp_field` are added as optional, non-computed string attributes in the EQL-specific section of `schema.go`, next to `tiebreaker_field`, with inline `MarkdownDescription`s that name Kibana's default (`event.category`, `@timestamp`) and, for `timestamp_field`, the difference from `timestamp_override`. Two matching `types.String` fields join `TiebreakerField` in the EQL-specific section of `Data`. Mirroring the existing attribute keeps one consistent shape for all three EQL settings that Kibana groups together, and needs no new abstraction. A nested EQL settings block was not chosen because `tiebreaker_field` already sits at the top level and moving it would break existing configurations.

**Mapping lives in the EQL processor only.**

- `toEqlRuleCreateProps` and `toEqlRuleUpdateProps` set `EventCategoryOverride` and `TimestampField` when the value is known, next to `TiebreakerField`.
- `updateFromEqlRule` reads both with `typeutils.StringishPointerValue`, so an absent field becomes null and a field set outside Terraform lands in state.
- `initializeTypeSpecificFieldsToDefaults` nulls both when unknown, next to `TiebreakerField`.

Other rule processors are untouched. Their request types have no such fields, so non-EQL rules never send them.

**Clearing relies on the full-replace update.** Removing an attribute from configuration leaves it null in the plan. The update request then omits the field, and Kibana's full replace deletes it. No explicit "unset" handling is needed. Because neither attribute forces replacement, the change is an in-place update.

**No version gate.** Both fields belong to the same `EqlOptionalFields` schema as `tiebreaker_field`, which the provider already sends without a gate. The acceptance steps that send the new fields run in a test gated at the resource's minimum supported version (8.11), so the matrix exercises them on every supported version.

**Testing approach.**

- Unit (`models_test.go`): create and update props carry both fields when set and omit them when null; the read mapping sets both from an API response and maps absent fields to null, including a value present in Kibana but unset in prior state; a non-EQL request does not carry them; `tiebreaker_field` follows the same mapping.
- Acceptance against the standard stack, placed by version gate:
  - `TestAccResourceSecurityDetectionRule_EQLMinimal` (runs from 8.11): create without the fields and assert both absent with an empty re-plan; an in-place update that sets both; an update back to the original configuration that removes both and asserts they are cleared; then a drift step that sets `event_category_override` on the rule directly through the Kibana API in `PreConfig` and expects refresh plus a non-empty plan.
  - `TestAccResourceSecurityDetectionRule_EQL` (runs from 8.16, gated by its response actions): create with both fields set, an in-place update that changes both values, and import verification.
- The drift check uses a `PreConfig` call through the acceptance-testing Kibana client and a `RefreshState` step that asserts the value lands in state. terraform-plugin-testing does not allow a configuration on a `RefreshState` step, so that step's own plan is a destroy and cannot show the update. A following plan-only step with the unchanged configuration therefore asserts an in-place update that plans `event_category_override` as null.

**Documentation.** The resource docs are regenerated from the schema descriptions with the repository's docs-generation target. The canonical spec's schema block gains both attributes during spec sync.

## Risks / Trade-offs

- [Rules that already carry either value set in the Kibana UI show a plan diff after upgrading] → Intended. Before this change the next update silently removed the value; now the plan shows it, and practitioners either add it to configuration or let the apply remove it. Called out in the proposal's Impact.
- [Kibana could return a server-filled default (`event.category`, `@timestamp`) when the field is unset, which would make null-on-absent produce drift] → The minimal acceptance test asserts both attributes stay absent with an empty re-plan, so such behavior would fail the build rather than ship.
- [Setting the attributes on a non-EQL rule is silently ignored] → Same as `tiebreaker_field` today. Type-scoped validation for all EQL-only attributes is a separate change.
- [An empty string is sent as-is] → Same as `tiebreaker_field`; Kibana's API decides whether to accept it. No new validation, to keep the three attributes consistent.

## Migration Plan

None. The attributes are additive and optional, so existing configurations and state need no change and no state upgrade.

## Open Questions

None.
