## Context

`elasticstack_kibana_slo` already validates `slo_id` in the Terraform schema before any API call is made. The current schema description and `stringvalidator.LengthBetween(8, 36)` constraint are stricter than the behavior we now want to support. Because this is a narrow change isolated to one attribute, the main design concern is keeping schema validation, documentation, and acceptance coverage aligned.

## Goals / Non-Goals

**Goals:**
- Allow practitioner-supplied SLO IDs up to 48 characters long.
- Keep the existing minimum length and allowed character set unchanged.
- Add automated coverage that proves a 48-character SLO ID succeeds.
- Update requirements to reflect the new provider contract.

**Non-Goals:**
- Changing how server-generated SLO IDs work.
- Expanding the allowed character set for `slo_id`.
- Modifying composite `id` handling, import behavior, or replacement semantics.

## Decisions

- Update the Plugin Framework schema validator for `slo_id` from `LengthBetween(8, 36)` to `LengthBetween(8, 48)`. This is the authoritative plan-time validation point, so changing it is sufficient to unblock valid configurations before create.
- Update the `slo_id` attribute description and OpenSpec requirements in the same change so user-facing docs and functional requirements remain consistent with the schema.
- Add acceptance coverage using a user-supplied 48-character ID instead of relying only on unit tests. This verifies the full Terraform-to-Kibana path and guards against future regressions in schema validation or request mapping.
- Leave all other `slo_id` validation rules unchanged, including the 8-character minimum, allowed characters regex, and replacement-on-change semantics. Alternatives such as removing plan-time length validation entirely were rejected because the provider should still reject clearly invalid input before API calls.

## Risks / Trade-offs

- [Acceptance fixture accidentally exceeds 48 characters] → Build the test input deliberately and assert the exact accepted value in state.
- [Requirements or schema text drift from implementation again] → Update the schema description, delta spec, and tests together in this change.
- [Kibana enforces a different limit in some versions] → Keep the change limited to the provider contract requested here and validate through the existing acceptance environment.