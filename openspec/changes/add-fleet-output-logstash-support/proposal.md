## Why

The Fleet output resource implementation currently supports `elasticsearch` and `kafka` output types, but does not fully support `logstash`. Fleet supports Logstash outputs as a first-class output type, so the Terraform provider should expose this behavior to keep provider functionality aligned with Fleet capabilities.

## What Changes

- Add full `logstash` output type support to `elasticstack_fleet_output` CRUD behavior.
- Ensure create/update request mapping and read/state mapping support `logstash` outputs consistently.
- Add or update acceptance and unit tests to cover `logstash` output creation, read/refresh, update, and import compatibility paths.
- Update provider documentation and examples for `elasticstack_fleet_output` to include `logstash` output usage and constraints.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `fleet-output`: Extend/clarify requirements so provider behavior for `logstash` output type is explicitly covered and verified by tests, matching Fleet API support.

## Impact

- Affected code: `internal/fleet/output` resource implementation, request/state mapping helpers, and tests under Fleet output resource coverage.
- Affected docs: Fleet output resource documentation and example configurations.
- External dependency alignment: Fleet Logstash output semantics and supported settings remain the source of truth.
