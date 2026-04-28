## Context

The current `elasticstack_kibana_slo` resource only supports the string arm of the Kibana KQL union fields used by `kql_custom_indicator.filter`, `good`, and `total`, even though the generated API models also allow an object form with `kqlQuery` and `filters`. The resource also relies on a mix of runtime validation and schema validation, which leaves several indicator-specific errors until apply time. In addition, the resource does not yet expose `settings.sync_field`, does not manage `artifacts`, and does not yet expose `enabled`, so enabled state is not currently represented in Terraform state.

This change spans the Terraform schema, SLO model conversion, shared conditional validators, the Kibana SLO client helpers, acceptance tests, and the canonical OpenSpec requirements. It is therefore worth documenting the design choices before implementation.

## Goals / Non-Goals

**Goals:**
- Add support for the object-form KQL union fields without breaking existing string-based configurations.
- Move indicator-specific validation to provider plan time where the provider can express the rule.
- Align simple schema validators with the current generated SLO API model where the repo has no evidence for a broader historical contract.
- Expose `settings.sync_field`, `artifacts`, and `enabled` in the Terraform resource.
- Implement write support for `enabled` using the actual SLO API surface exposed by the generated client.

**Non-Goals:**
- Redesign the entire SLO schema around dynamic JSON-like blocks.
- Change the existing indicator block model or the one-indicator-only resource shape.
- Replace the current string KQL attributes with a breaking schema migration in this change.
- Expand support beyond the current `artifacts.dashboards[].id` shape modeled in the filtered spec.

## Decisions

### 1. Additive `_kql` attributes for object-form KQL unions

The resource will keep the current string attributes:
- `kql_custom_indicator.filter`
- `kql_custom_indicator.good`
- `kql_custom_indicator.total`

It will add parallel object-form attributes with a `_kql` suffix:
- `filter_kql`
- `good_kql`
- `total_kql`

Each `_kql` attribute will model the object union arm with:
- `kql_query`
- `filters`

This preserves backward compatibility for existing configurations while allowing the provider to round-trip the full Kibana API shape. A more invasive replacement with a single unified object would be cleaner conceptually, but it would either break existing configuration or require a more complex state and schema migration than this change needs.

### 2. Enforce one representation per KQL field

For each of `filter`, `good`, and `total`, the provider will enforce that only one of the string or `_kql` forms is configured. This keeps plans unambiguous and avoids precedence rules that would be hard to explain. The read path will populate whichever representation was used in configuration when possible; when the API returns the object form with filters, the provider should retain that richer form in state rather than degrading it to the string-only representation.

### 3. Reuse conditional validators for indicator-specific rules

The provider already has reusable conditional validators in `internal/utils/validators/conditional.go`, including relative-path helpers that work well for nested list elements. This change will use those validators in the SLO schema to enforce rules such as:
- metric `field` required except for `doc_count`
- metric `field` forbidden for `doc_count`
- timeslice `percentile` required only for `percentile`
- timeslice `percentile` forbidden for non-`percentile`
- indicator-specific required nested blocks that are currently only rejected in model conversion

If the current helper error text is awkward for multi-value conditions, the utility may be tightened as part of this change, but the main plan is to reuse the existing validator pattern rather than invent a SLO-only validation mechanism.

### 4. Align schema validation to the generated contract unless proven otherwise

The repo currently shows a mismatch between the Terraform `slo_id` validator (`8..48`) and the generated SLO create model comment (`8..36`). The workspace does not contain evidence that earlier stack versions allowed a longer SLO id, so this change will align the provider to `8..36` unless implementation uncovers stronger version-specific evidence.

The same principle applies to other simple validators:
- custom metric `aggregation` will be restricted to `sum` and `doc_count`
- metric variable names used in SLO equations will be validated against `^[A-Z]$`, matching the generated API contract that documents valid options as `A-Z`
- `time_window.type` will be restricted to `rolling` and `calendarAligned`

### 5. Manage `enabled` through dedicated APIs, not update payloads

The generated `SLOsUpdateSloRequest` does not include an `enabled` field, but the generated client exposes `EnableSloOpWithResponse` and `DisableSloOpWithResponse`. The resource will therefore manage `enabled` as a first-class Terraform attribute, but write reconciliation will be implemented by:
1. creating or updating the SLO definition with the standard create/update APIs
2. calling enable or disable if the post-write server state does not match the planned `enabled` value
3. reading back the SLO again to settle state

This matches the actual API shape and avoids baking assumptions about hidden update-body support into the resource.

### 6. Expose `artifacts` as modeled metadata

The filtered spec currently models `artifacts` as dashboard references under `artifacts.dashboards[].id`. This change will expose that exact shape rather than inventing a more generic artifact model. The provider should treat it as managed metadata attached to the SLO definition, not as derived or computed server-only state.

### 7. Keep `settings.sync_field` in the existing `settings` object

`settings.sync_field` belongs in the existing `settings` block alongside `sync_delay`, `frequency`, and `prevent_initial_backfill`. This avoids another top-level field and keeps the Terraform schema aligned with the generated `SLOsSettings` structure. The settings object handling in the model layer and any state-upgrade logic will be updated accordingly.

## Risks / Trade-offs

- **Schema growth in `kql_custom_indicator`** → Mitigation: use a consistent `_kql` suffix and keep the object shape minimal so it reads as a parallel representation, not a second subsystem.
- **State round-tripping for KQL unions may be subtle** → Mitigation: add focused unit tests for all string/object combinations and read-back behavior involving `filters`.
- **`enabled` writes may require extra API calls and read-backs** → Mitigation: keep the sequence explicit and only invoke enable/disable when the desired state differs from the server state.
- **`slo_id` alignment to 36 could reject configs that previously planned locally** → Mitigation: confirm no stronger in-repo evidence exists before implementation and document the narrowed contract in the change artifacts.
- **Conditional validation can become noisy if misapplied to nested blocks** → Mitigation: mirror existing validator usage patterns from other resources and keep runtime validation only for cases the framework cannot express cleanly.

## Migration Plan

The preferred rollout is additive and low risk:
1. Add new schema fields and validators.
2. Extend model conversion and client helpers.
3. Add or update tests to lock in round-trip and validation behavior.
4. Regenerate resource docs.

No external data migration is expected. Existing configurations using string KQL inputs will continue to work. If a user adopts `_kql` inputs, state should converge through normal read-after-create and read-after-update behavior. If `enabled` is added with a defaulted value, implementation should ensure the default matches current server behavior to avoid surprise diffs.

## Open Questions

- Whether `filter` on timeslice `doc_count` metrics should remain accepted unless the API proves otherwise, or be forbidden up front for stricter alignment.
- Whether `artifacts` should be exposed immediately for both create and update, or staged behind read support first if acceptance testing shows server-side normalization.
- Whether state-upgrade logic needs an explicit schema version bump for the expanded `settings` object and new `_kql` fields, or whether the plugin framework can absorb the additive changes without a dedicated upgrader.
