## Why

`elasticsearch_connection` is protected by provider-level schema coverage tests and a custom lint rule that catches helper-bypass client resolution. If `kibana_connection` is added without equivalent verification, future entities can silently miss the block or expose it without actually using the scoped client.

## What Changes

- Add provider-level automated coverage requirements for `kibana_connection` so in-scope Kibana and Fleet entities fail tests when they are missing the expected block definition.
- Add lint requirements that enforce helper-derived client resolution for Kibana and Fleet code paths that are supposed to honor `kibana_connection`.
- Define regression coverage expectations for both the unit-test and lint layers so the verification runs in normal local and CI workflows.

## Capabilities

### New Capabilities
- `provider-kibana-connection-coverage`: provider test requirements for enforcing `kibana_connection` schema/block coverage across in-scope Kibana and Fleet entities
- `kibana-fleet-client-resolution-lint`: lint requirements that prove Kibana and Fleet entity code uses approved helper-derived client resolution when `kibana_connection` is in scope

### Modified Capabilities
<!-- None. -->

## Impact

- Provider-level connection coverage tests under `provider/`
- Custom analyzers under `analysis/`
- `.golangci.yaml` and repository lint execution
- Kibana and Fleet entity code under `internal/kibana/` and `internal/fleet/`
- OpenSpec specs describing coverage and lint enforcement
