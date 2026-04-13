## Why

When a `semantic_text` field is defined in an index mapping with an `inference_id`, Elasticsearch automatically enriches the stored mapping with a `model_settings` object (containing `dimensions`, `element_type`, `service`, `similarity`, `task_type`). The mappings plan modifier does not account for this server-side enrichment, causing Terraform to report "Provider produced inconsistent result after apply" on every subsequent apply — making the resource unusable with `semantic_text` fields.

## What Changes

- The mappings plan modifier (`mapping_modifier.go`) is extended to detect `semantic_text` fields and copy server-enriched `model_settings` from state into the plan when the user has not explicitly specified `model_settings` in their config.
- If the user explicitly specifies `model_settings`, those values are respected and not overwritten.
- Unit tests are added to `mapping_modifier_test.go` covering: semantic_text without explicit model_settings, with explicit model_settings, and nested semantic_text fields.
- The `elasticsearch-index` spec is updated to document the new modifier behavior for `semantic_text` fields (REQ-022–REQ-024 extension).

## Capabilities

### New Capabilities

_(none — this is a bug fix to existing behavior)_

### Modified Capabilities

- `elasticsearch-index`: The mappings plan modifier requirement (REQ-022–REQ-024) is extended to describe how server-enriched fields on `semantic_text` mappings are handled to prevent spurious drift detection.

## Impact

- `internal/elasticsearch/index/index/mapping_modifier.go` — core logic change
- `internal/elasticsearch/index/index/mapping_modifier_test.go` — new unit test cases
- `openspec/specs/elasticsearch-index/spec.md` — updated requirements for the plan modifier
