# Fleet output data source design

## Status
This data source is GA and should be exposed by default.

## Overview
The `elasticstack_fleet_output` data source retrieves a single Fleet output from Kibana Fleet by ID.
It is read-only and mirrors the output configuration returned by the Fleet Output API.

## API interactions
- Uses the generated Kibana client `kbapi` via `internal/clients/fleet.GetOutput`.
- Calls `GET /api/fleet/outputs/{outputId}` with a space-aware request editor when `space_id` is provided.
- The response is deserialized into `kbapi.OutputUnion` and mapped into Terraform state.

## Schema
### Inputs
- `output_id` (string, required): Fleet output identifier.
- `space_id` (string, optional): Kibana space ID used for the request path. Omitted means the default space. This remains null if unknown during planning.

### Computed attributes
- `id` (string): Data source identifier in the format `<space_id>/<output_id>`.
- `name` (string): Output name from the API response.
- `type` (string): Output type discriminator from the API response.
- `hosts` (list of string): Output hosts from the API response.
- `ca_sha256` (string): CA SHA256 fingerprint from the API response.
- `ca_trusted_fingerprint` (string): Trusted CA fingerprint from the API response. Empty strings are normalized to null.
- `default_integrations` (bool): Maps from `is_default` in the API response.
- `default_monitoring` (bool): Maps from `is_default_monitoring` in the API response.
- `config_yaml` (string, sensitive): Advanced YAML configuration returned by the API.
- `ssl` (object):
  - `certificate_authorities` (list of string): CA certificates; null when the API returns none.
  - `certificate` (string): Client certificate.
  - `key` (string, sensitive): Client certificate key.
  - The `ssl` object is null when all SSL fields are empty.
- `kafka` (object): Kafka-specific settings when the output type is `kafka`.
  - `auth_type`, `broker_timeout`, `client_id`, `compression`, `compression_level`, `connection_type`,
    `topic`, `partition`, `required_acks`, `timeout`, `version`, `username`, `password`, `key`.
  - `headers` (list of objects): Each header has `key` and `value`.
  - `hash` (object): `hash`, `random`.
  - `random` (object): `group_events`.
  - `round_robin` (object): `group_events`.
  - `sasl` (object): `mechanism`.
  - Each nested object is set to null when the API omits that sub-structure.

## Translation details
- The data source reads a `kbapi.OutputUnion`, inspects its discriminator, and maps it into a shared output model.
- Kafka numeric fields map from optional API pointers to Terraform numbers; absent values become null.

## Limitations
- Only output types `elasticsearch`, `logstash`, and `kafka` are supported. Any other discriminator is rejected.

## Required acceptance test cases
- Read an Elasticsearch output via `output_id` and verify every computed attribute.
- Read a Logstash output via `output_id` and verify every computed attribute.
- Read a Kafka output via `output_id` and verify every computed attribute.
- Read an output via `output_id` in a non-default space when the resource supports `space_id`.
