# Fleet outputs data source design

## Status
This data source is GA and should be exposed by default.

## Overview
The `elasticstack_fleet_outputs` data source retrieves a list of Fleet outputs from Kibana Fleet.
It is read-only and mirrors the output configuration returned by the Fleet Outputs API.
The outputs can optionally be filtered by attributes supplied to the data source. 

## API interactions
- Uses the generated Kibana client `kbapi` via `internal/clients/fleet.GetOutputs`.
- Calls `GET /api/fleet/outputs` with a space-aware request editor when `space_id` is provided.
- The response is deserialized into `kbapi.OutputUnion` and mapped into Terraform state.

## Schema
### Inputs
- `output_id` (string): Filters the returned outputs by output ID.
- `type` (string): Filters the returned outputs by type. Only `elasticsearch`, `logstash`, and `kafka` are currently supported. 
- `default_integrations` (bool): Filters the returned outputs to the default integration output
- `default_monitoring` (bool): Filters the returned outputs to the default monitoring output
- `space_id` (string, optional): Kibana space ID used for the request path. Omitted means the default space. This remains null if unknown during planning.

### Computed attributes
- `items` (list): A list of matching outputs

#### Output item attributes
- `id` (string): The ouptut ID
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

## Required acceptance test cases:
- A Kibana space with no  that sub-structure.

## Translation details
- The data source reads a `kbapi.OutputUnion`, inspects its discriminator, and maps it into a shared output model.
- Kafka numeric fields map from optional API pointers to Terraform numbers; absent values become null.

## Limitations
- Only output types `elasticsearch`, `logstash`, and `kafka` are supported. Any other discriminator is rejected.


## Required acceptance test cases

- A Kibana space with no Fleet outputs. 
    - The data source shall:
        - Return successfully. No panics or errors
        - Return an empty items list
- A Kibana space with a single output, with no filters applied. 
    - The data source shall:
        - Return a list containing the single matching output
- Multiple outputs of a single type within a single space and no filters applied. 
    - The data source shall:
        - Return a list of all matching outputs.
- Multiple outputs of different supported types within a single space and no filters applied.
    - The data source shall:
        - Return a list of all supported outputs
- Multiple outputs of different types within a single space. 
    - Each filter and combination of filters shall be tested
    - The data source shall:
        - Return only those outputs that match a specific filter combination. 