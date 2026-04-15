## Why

`elasticsearch_connection` is protected by provider-level schema coverage tests. If `kibana_connection` is added without equivalent verification, future entities can silently miss the block or drift from the shared schema helper.

## What Changes

- Add provider-level automated coverage requirements for `kibana_connection` so in-scope Kibana and Fleet entities fail tests when they are missing the expected block definition.
- Define regression coverage expectations for the provider-level test layer so the verification runs in normal local and CI workflows.

## Capabilities

### New Capabilities
- `provider-kibana-connection-coverage`: provider test requirements for enforcing `kibana_connection` schema/block coverage across in-scope Kibana and Fleet entities

### Modified Capabilities
<!-- None. -->

## Impact

- Provider-level connection coverage tests under `provider/`
- OpenSpec specs describing coverage enforcement
