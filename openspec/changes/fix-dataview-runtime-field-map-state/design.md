## Context

The `elasticstack_kibana_data_view` resource manages Kibana data views through Kibana's Data Views HTTP API. Kibana's update endpoint follows a partial-update contract: only the properties provided in the request body are modified; omitted properties remain unchanged in Kibana's persisted state.

Terraform's Plugin Framework distinguishes between:
- **Optional only**: Terraform expects the attribute in state to exactly match the config (whether set or null). If the provider returns a different value, Terraform raises a consistency error.
- **Optional + Computed**: Terraform allows the provider to return a value that differs from config, because the real value is determined by the external system.

Currently `data_view.runtime_field_map` is `Optional` but not `Computed`. When a user removes `runtime_field_map` from their Terraform configuration, Kibana preserves the existing runtime fields, the provider reads them back on refresh, and Terraform rejects the state because `null (plan) ≠ non-null (state)`.

This same risk exists for other collection attributes (`source_filters`, `field_formats`), but the issue reporter only hit `runtime_field_map`. The Kibana API docs claim all unspecified fields stay persisted, yet the current acceptance tests pass for those other attributes. That suggests either:
- The test currently masks the issue because `field_attrs` removal forces resource replacement, so the "update" path is never exercised, or
- Kibana's preservation behavior is field-specific despite the generic docs.

## Goals / Non-Goals

**Goals:**
- Fix the Terraform state consistency error for `runtime_field_map` when it is omitted from config but Kibana preserves it
- Update the acceptance test to assert true in-place update (not replacement) behavior for the data view resource
- Document the expected behavior for `runtime_field_map` persistence in requirements

**Non-Goals:**
- Changing `source_filters` or `field_formats` to `Computed` — only if acceptance tests reveal they have the same problem after the test fix
- Adding new provider functionality (e.g., dedicated runtime field sub-resource)
- Changing how Kibana's API works (we work with its partial-update semantics)

## Decisions

### Decision 1: Add `Computed: true` to `runtime_field_map`

**Rationale:** This is the minimal, correct fix. Terraform's semantics say: if an external system can produce a value different from what the user configured, the attribute must be `Computed`. Kibana's partial-update contract means the persisted `runtime_field_map` can diverge from config when omitted, so `Computed` is semantically correct.

**Alternative considered:** Send an explicit empty map `{}` when config omits `runtime_field_map`. Rejected because it would change behavior — users who explicitly omit the field would see their Kibana runtime fields deleted, which is a breaking change and contrary to the current documented/tested behavior.

### Decision 2: Keep `field_attrs` in `basic_updated` testdata

**Rationale:** `field_attrs` has a `RequiresReplace()` plan modifier. Removing it from config in `basic_updated` forces replacement, which masks the `runtime_field_map` bug because the new resource starts with no runtime fields. To test the true update path, the step must avoid any attribute that triggers replacement.

### Decision 3: Assert `runtime_field_map` is still present in the update step

**Rationale:** Because Kibana preserves the field, and now it's `Computed`, Terraform will accept the non-null value into state even when config omits it. The test should assert this preservation rather than asserting absence.

## Risks / Trade-offs

- **[Risk]** Setting `Computed: true` means Terraform will silently accept drift between config and state. If a user genuinely wants to remove all runtime fields, they must explicitly set `runtime_field_map = {}`. `TestCheckNoResourceAttr` is replaced with `TestCheckResourceAttr`.
  - **Mitigation**: This is already how the system behaves (Kibana preserves the field). `Computed` only makes Terraform stop complaining about the existing behavior. The `runtime_field_map = {}` workaround is already documented in issue #2135.

- **[Risk]** The change to `Computed` could cause import verification to fail if `runtime_field_map` is not present in config but present in imported state.
  - **Mitigation**: Add `ImportStateVerifyIgnore` for `data_view.runtime_field_map` in the import test step.

## Open Questions

- Do `source_filters` and `field_formats` also require `Computed: true`? The Kibana docs suggest yes, but current tests don't exercise true update for those fields. This will be resolved by running the fixed acceptance test; if those fields also produce consistency errors, they can be added to a follow-up change.
