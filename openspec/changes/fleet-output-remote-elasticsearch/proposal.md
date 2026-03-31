## Why

The `elasticstack_fleet_output` resource currently supports only `elasticsearch` and `kafka` output types, while Fleet also supports `remote_elasticsearch`. This creates a provider capability gap that forces users to manage remote Elasticsearch outputs outside Terraform.

## What Changes

- Extend `elasticstack_fleet_output` resource requirements to support `type = "remote_elasticsearch"` alongside existing supported types.
- Define schema and validation behavior for remote Elasticsearch-specific authentication fields (service token) and optional TLS/mTLS settings.
- Define create/read/update/delete behavior and state mapping for remote Elasticsearch outputs, including sensitive value handling where Fleet stores secrets.
- Update acceptance and unit test requirements to cover remote Elasticsearch lifecycle and validation paths.
- Update user documentation requirements so `remote_elasticsearch` usage is discoverable and consistent with Fleet semantics.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `fleet-output`: Expand resource requirements to include `remote_elasticsearch` output type behavior, schema, validation, API mapping, and test coverage expectations.

## Impact

- Affected code: `internal/fleet/output` schema, model mapping, CRUD request/response handling, validators, and tests.
- Affected docs: `docs/resources/fleet_output.md` and related examples for output type coverage.
- Affected API behavior: provider requests to Fleet output APIs will include remote Elasticsearch payload fields.
- Dependencies: no new external dependencies expected; relies on existing Fleet API client capabilities for remote Elasticsearch outputs.
