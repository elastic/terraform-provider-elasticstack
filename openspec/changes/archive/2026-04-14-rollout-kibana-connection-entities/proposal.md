## Why

Even with provider-level `kibana_connection` support, the feature is not useful until Kibana and Fleet Terraform entities actually expose and honor the block. Many existing specs currently state that these entities only use provider-level clients, so the rollout needs an explicit spec update across the affected resources and data sources.

## What Changes

- Add a provider-level rollout contract that identifies which Kibana and Fleet Terraform entities expose `kibana_connection` and how they resolve an effective scoped client.
- Define rollout scope from the Kibana and Fleet entity registrations in `provider/provider.go` and `provider/plugin_framework.go`, including the Plugin Framework's conditionally registered experimental Kibana resources.
- Add `kibana_connection` to the in-scope Kibana and Fleet Terraform entities, keeping the block consistent with the shared provider-level `kibana_connection` schema.
- Update entity requirements so CRUD and read operations use the resource- or data-source-scoped connection when the block is configured, and fall back to the provider client otherwise.
- Keep existing entity-specific behavior such as space handling, import, and version gates, but require those operations to execute against the effective scoped client.
- Remove the rollout gap for registered Kibana and Fleet entities so the plan matches the provider surface exposed from both provider implementations.

## Capabilities

### New Capabilities
- `provider-kibana-connection-entities`: provider-level requirements for which Kibana and Fleet Terraform entities expose `kibana_connection` and use the effective scoped client

### Modified Capabilities
- `fleet-agent-policy`: replace the provider-only Fleet client requirement with optional `kibana_connection` override behavior
- `fleet-elastic-defend-integration-policy`: replace the provider-only Fleet client requirement with optional `kibana_connection` override behavior
- `fleet-enrollment-tokens`: add optional `kibana_connection` override behavior for the Fleet enrollment tokens data source
- `fleet-integration`: replace the provider-only Fleet client requirement with optional `kibana_connection` override behavior
- `fleet-integration-policy`: replace the provider-only Fleet client requirement with optional `kibana_connection` override behavior
- `fleet-output`: add optional `kibana_connection` override behavior for the Fleet output resource and data source
- `fleet-server-host`: add optional `kibana_connection` override behavior for the Fleet server host resource
- `kibana-action-connector`: add rollout coverage for the action connector resource and data source connection behavior
- `kibana-agentbuilder-export-workflow`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-agentbuilder-workflow`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-alerting-rule`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-dashboard`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-data-view`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-default-data-view`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-export-saved-objects`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-import-saved-objects`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-install-prebuilt-rules`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-maintenance-window`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-detection-rule`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-enable-rule`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-exception-item`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-exception-list`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-list`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-list-data-streams`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-list-item`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-security-role`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-slo`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-space`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-spaces`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-synthetics-monitor`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-synthetics-parameter`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior
- `kibana-synthetics-private-location`: replace the provider-only Kibana client requirement with optional `kibana_connection` override behavior

The experimental `elasticstack_kibana_stream` registration is also in rollout scope because it is registered in `provider/plugin_framework.go`. This change tracks that entity through the provider-level rollout capability and the implementation task plan; there is not yet a standalone `openspec/specs/kibana-stream/` capability to modify alongside the other registered entities.

## Impact

- Kibana and Fleet resource/data-source schemas under `internal/kibana/` and `internal/fleet/`
- Resource and data source create/read/update/delete paths that currently use provider-level clients only
- OpenSpec capability specs for the adopted Kibana and Fleet entities, plus the provider-level rollout scope that covers the registered experimental `elasticstack_kibana_stream` resource
- Generated documentation for the affected Terraform entities
