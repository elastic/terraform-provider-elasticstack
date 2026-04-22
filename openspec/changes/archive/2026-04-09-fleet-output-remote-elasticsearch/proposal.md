## Why

The `elasticstack_fleet_output` resource currently supports only `elasticsearch` and `kafka` output types, while Fleet also supports `remote_elasticsearch`. This creates a provider capability gap that forces users to manage remote Elasticsearch outputs outside Terraform.

## What Changes

- Extend `elasticstack_fleet_output` resource requirements to support `type = "remote_elasticsearch"` alongside existing supported types.
- Define schema and validation behavior for remote Elasticsearch-specific authentication fields (service token) and optional TLS/mTLS settings.
- Expose **automatic integrations synchronization** for remote outputs: a Terraform attribute (aligned with Fleet’s `sync_integrations`) so users can turn integration asset sync to the remote cluster on or off without relying solely on `config_yaml` or out-of-band UI changes.
- Where the Fleet API exposes them for remote Elasticsearch outputs, cover companion toggles needed for a complete remote-output story (for example `sync_uninstalled_integrations` and `write_to_logs_streams` / wired streams), subject to server version and subscription constraints documented by Elastic.
- Define create/read/update/delete behavior and state mapping for remote Elasticsearch outputs, including sensitive value handling where Fleet stores secrets.
- Update acceptance and unit test requirements to cover remote Elasticsearch lifecycle, validation paths, and toggles for sync-related fields when the target stack supports them.
- Update user documentation requirements so `remote_elasticsearch` usage is discoverable and consistent with Fleet semantics, including limitations (e.g. subscription tiers for multi-cluster sync, Elastic Defend limitations on remote outputs).

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `fleet-output`: Expand resource requirements to include `remote_elasticsearch` output type behavior, schema, validation, API mapping, and test coverage expectations—including **sync integrations** and other remote-output-specific Fleet fields needed for full configuration parity with the API.

## Impact

- Affected code: `internal/fleet/output` schema, model mapping, CRUD request/response handling, validators, and tests.
- Affected docs: `docs/resources/fleet_output.md` and related examples for output type coverage.
- Affected API behavior: provider requests to Fleet output APIs will include remote Elasticsearch payload fields (auth, TLS, shipper tuning as exposed, **sync_integrations** and related flags where applicable).
- Dependencies: no new external dependencies expected; relies on existing Fleet API client capabilities for remote Elasticsearch outputs.
