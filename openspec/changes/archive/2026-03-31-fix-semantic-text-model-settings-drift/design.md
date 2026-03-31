## Context

The `mappings` attribute on `elasticstack_elasticsearch_index` uses a custom plan modifier (`mappingsPlanModifier`) to handle the fact that Elasticsearch silently ignores field removals. The modifier compares state and config field-by-field, carrying fields forward or signaling replacement as appropriate.

When a `semantic_text` field is created with an `inference_id`, Elasticsearch automatically enriches the stored mapping with a `model_settings` object (containing `dimensions`, `element_type`, `service`, `similarity`, `task_type`). Prior to this fix the plan modifier did not account for this enrichment, so the plan value lacked `model_settings` while the post-apply read returned it — causing "Provider produced inconsistent result after apply."

## Goals / Non-Goals

**Goals:**
- Prevent "inconsistent result after apply" when Elasticsearch enriches `semantic_text` fields with `model_settings`
- Respect user-specified `model_settings` (not overwritten if present in config)

**Non-Goals:**
- General carry-forward of arbitrary server-enriched keys for other field types
- Acceptance tests requiring a live inference endpoint (deferred — environment dependency)
- Docs site update (no user-facing schema change)

## Decisions

### Decision: semantic_text-specific model_settings handling, not a general key carry-forward

**Chosen**: In the type-match branch of `modifyMappings`, when the field type is `semantic_text`, copy `model_settings` from state into the plan if and only if `model_settings` is absent from the config.

**Alternative**: Copy any key present in state but absent from config for any matching-type field.

**Rationale**: The general approach silently masks drift for all field types. Targeting `semantic_text` + `model_settings` makes the intent explicit and avoids unintended side-effects on other field types where missing keys should still be visible as drift.

## Risks / Trade-offs

- **Risk**: A future field type has similar server-enriched sub-objects → Mitigation: add a targeted case then, same pattern.
- **Trade-off**: If the user explicitly specifies `model_settings` in config, those values are kept as-is (not overwritten by state). This is the correct behaviour but means a user who changes `model_settings` manually may see replacement prompts — which is expected.
