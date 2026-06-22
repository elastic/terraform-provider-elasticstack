## Why

The `elasticstack_elasticsearch_ml_datafeed` and `elasticstack_elasticsearch_ml_filter` resources
validate their primary ID attributes (`datafeed_id` and `filter_id`) through the shared
`ml.IDValidator()` helper, which enforces a hard maximum of **64 characters**. Elasticsearch
imposes no such length restriction for datafeed IDs or filter IDs — the API only requires that
identifiers consist of lowercase alphanumeric characters, hyphens, underscores, and dots, starting
and ending with an alphanumeric character.

As a result, any datafeed or filter whose ID exceeds 64 characters cannot be declared in HCL at
all: `terraform validate` fails before any API call, making it impossible to import or manage
existing long-named resources. This is a provider-only restriction with no API-side basis.

Related issue: #3762

## What Changes

- Add `IDValidatorWithoutLength()` to `internal/elasticsearch/ml/idvalidator.go` — a validator
  that enforces only the character-class regex (no upper-bound length), while still requiring at
  least one character (`LengthAtLeast(1)`).
- Update `internal/elasticsearch/ml/datafeed/schema.go` to use `ml.IDValidatorWithoutLength()` for
  the `datafeed_id` attribute.
- Update `internal/elasticsearch/ml/datafeed_state/schema.go` to use
  `ml.IDValidatorWithoutLength()` for its `datafeed_id` attribute.
- Update `internal/elasticsearch/ml/filter/schema.go` to use `ml.IDValidatorWithoutLength()` for
  the `filter_id` attribute.
- Add unit tests for `IDValidatorWithoutLength()` to `internal/elasticsearch/ml/idvalidator_test.go`.
- Update the OpenSpec delta specs to remove the 64-character upper-bound constraint from
  `datafeed_id` (in both `elasticsearch-ml-datafeed` and `elasticsearch-ml-datafeed-state`) and
  from `filter_id` (in `elasticsearch-ml-filter`).

The shared `IDValidator()` (used by `anomaly_detection_job`, `calendar`, `calendar_event`,
`calendar_job`, `job_state`) is **not** changed by this proposal.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `elasticsearch-ml-datafeed`: remove the 64-character upper-bound from `datafeed_id` validation.
- `elasticsearch-ml-datafeed-state`: remove the 64-character upper-bound from `datafeed_id`
  validation.
- `elasticsearch-ml-filter`: remove the 64-character upper-bound from `filter_id` validation.

## Impact

- `internal/elasticsearch/ml/idvalidator.go` — add `IDValidatorWithoutLength()`.
- `internal/elasticsearch/ml/idvalidator_test.go` — add tests for `IDValidatorWithoutLength()`.
- `internal/elasticsearch/ml/datafeed/schema.go:65` — swap `ml.IDValidator()` → `ml.IDValidatorWithoutLength()`.
- `internal/elasticsearch/ml/datafeed_state/schema.go:57` — swap `ml.IDValidator()` → `ml.IDValidatorWithoutLength()`.
- `internal/elasticsearch/ml/filter/schema.go:52` — swap `ml.IDValidator()` → `ml.IDValidatorWithoutLength()`.
- Delta specs under `openspec/changes/ml-datafeed-filter-id-validator-without-length/specs/`.
