## Why

Migrate `elasticstack_elasticsearch_logstash_pipeline` from Plugin SDK to Plugin Framework and the entitycore envelope. This is a medium-complexity SDK resource with a flat settings model (key/value pairs with types) that maps well to PF once the schema is restructured.

## What Changes

- Replace `ResourceLogstashPipeline()` SDK resource in `internal/elasticsearch/logstash/pipeline.go` with PF resource.
- Define PF schema for pipeline metadata, settings, and pipeline configuration.
- Convert model to PF types with envelope-required getters.
- Wire envelope callbacks.
- Maintain the existing `allSettingsKeys` type mapping for validation/flattening logic.

## Capabilities

### New Capabilities
- `logstash-pipeline-resource-via-envelope`

### Modified Capabilities
<!-- None. -->

## Impact

- `internal/elasticsearch/logstash/pipeline.go` and `pipeline_test.go`.
