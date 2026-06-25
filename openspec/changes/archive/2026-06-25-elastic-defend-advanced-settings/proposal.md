## Why

`elasticstack_fleet_elastic_defend_integration_policy` exposes typed protection, event collection, and popup settings but has no way to configure Elastic Defend **advanced settings**. Many production scenarios—especially [air-gapped environments](https://www.elastic.co/docs/solutions/security/configure-elastic-defend/configure-offline-endpoints-air-gapped-environments)—require advanced settings such as `linux.advanced.artifacts.global.base_url` and related artifact/proxy/TLS options documented in [Elastic Defend advanced settings](https://www.elastic.co/docs/reference/security/defend-advanced-settings). Without provider support, users cannot fully manage Defend integration policies in Terraform and must fall back to manual Kibana changes or workarounds outside IaC.

## What Changes

- Add an optional `advanced_settings` attribute (`map(string)`) to `elasticstack_fleet_elastic_defend_integration_policy`.
- Map Terraform keys (Elastic's documented dot-notation names, e.g. `windows.advanced.artifacts.global.base_url`) to the nested `policy.{os}.advanced` structure in the Fleet typed Defend package policy payload.
- Round-trip `advanced_settings` on read and import when the attribute is managed in configuration; preserve existing lifecycle behavior for finalize/update (version token, artifact manifest handling).
- Add unit tests for advanced-settings flatten/unflatten and request construction; add acceptance coverage for a representative air-gapped setting (artifact base URL).
- Update resource documentation and the `fleet-elastic-defend-integration-policy` capability spec.

Non-breaking: existing configurations without `advanced_settings` continue to work unchanged.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `fleet-elastic-defend-integration-policy`: add `advanced_settings` schema, API mapping requirements, and lifecycle behavior for create/read/update/import.

## Impact

- **Code**: `internal/fleet/elastic_defend_integration_policy/` (`schema.go`, `models.go`, `mapping.go`, `request.go`, tests), generated docs.
- **Specs**: delta under this change; canonical `openspec/specs/fleet-elastic-defend-integration-policy/spec.md` updated when archived.
- **API**: no `kbapi` changes; advanced settings are part of the existing typed `policy` config envelope.
- **Users**: can configure any documented advanced setting via Terraform; values are opaque strings validated by Elastic Endpoint at runtime.
