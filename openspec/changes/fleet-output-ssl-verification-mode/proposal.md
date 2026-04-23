## Why

The `ssl` block on `elasticstack_fleet_output` is missing the `verification_mode` attribute, even though the Fleet API (`KibanaHTTPAPIsOutputSsl`) already supports it. Users who need to set `ssl.verification_mode` (e.g. to `"none"` for self-signed cert environments) are forced to use the opaque `config_yaml` escape hatch instead of a proper first-class attribute (see [#1375](https://github.com/elastic/terraform-provider-elasticstack/issues/1375)).

## What Changes

- Add `verification_mode` to the `ssl` block schema for `elasticstack_fleet_output` — optional string, valid values: `"certificate"`, `"full"`, `"none"`, `"strict"`
- Populate `verification_mode` in the intermediate `outputSslModel` and `outputSSLAPIModel` structs
- Map `verification_mode` when converting from Terraform state → API request (`toAPI()`)
- Map `verification_mode` when converting from API response → Terraform state (`sslToObjectValue()`)
- Update all callers of `sslToObjectValue()` across all output type models (Elasticsearch, Logstash, Kafka, Remote Elasticsearch)

## Capabilities

### New Capabilities

_(none — this is a gap-fill on an existing capability)_

### Modified Capabilities

- `fleet-output`: Adding `ssl.verification_mode` attribute — a new optional field on the existing `ssl` block. The requirement for what the `ssl` block supports is changing.

## Impact

- `internal/fleet/output/schema.go` — add attribute to ssl block
- `internal/fleet/output/models_ssl.go` — update model structs, `toAPI()`, `objectValueToSSL()`, `sslToObjectValue()`
- `internal/fleet/output/models_kafka.go` — update `fromAPIKafkaModel()` caller of `sslToObjectValue()`
- `internal/fleet/output/models_elasticsearch.go` — update `fromAPIElasticsearchModel()` caller
- `internal/fleet/output/models_logstash.go` — update `fromAPILogstashModel()` caller
- `internal/fleet/output/models_remote_elasticsearch.go` — update `fromAPIRemoteElasticsearchModel()` caller
- `internal/fleet/output/models_ssl_test.go` — update unit tests for SSL model functions
- No API client changes needed; `KibanaHTTPAPIsOutputSsl.VerificationMode` is already present in the generated client
