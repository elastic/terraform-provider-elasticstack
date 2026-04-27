## Why

`elasticstack_fleet_output` and `elasticstack_fleet_server_host` use `ImportStatePassthroughID`, which passes the raw import string through as the resource ID. When a user imports from a non-default Kibana space using `<space_id>/<resource_id>`, the composite string lands in `output_id`/`host_id` and `space_ids` is never set — causing the subsequent Read to query the wrong ID against the wrong space (404). Additionally, the four Fleet resources that already handle space-aware import each do so with bespoke logic, creating fragmentation and subtle behavioral inconsistencies.

## What Changes

- Introduce `fleet.SpaceImporter`, an embeddable struct in `internal/fleet/` that provides a canonical `ImportState` implementation for resources with `space_ids`
- `SpaceImporter` accepts one or more `path.Path` values (the resource-specific ID fields to seed), parses composite `<space_id>/<resource_id>` import IDs via `clients.CompositeIDFromStrFw`, and falls back to a plain ID with no `space_ids` set
- Fix `fleet_output` by embedding `*SpaceImporter` wired to `path.Root("output_id")`
- Fix `fleet_server_host` by embedding `*SpaceImporter` wired to `path.Root("host_id")`
- Migrate `fleet_agent_policy`, `fleet_integration_policy`, `fleet_elastic_defend_integration_policy`, and `fleet_agent_binary_download_source` to use `*SpaceImporter`, removing their bespoke `ImportState` methods
- Normalize `fleet_agent_binary_download_source` behavior: plain import ID leaves `space_ids` unset (was: hardcoded to `["default"]`); plain ID no longer also sets `id` (Terraform resource ID is repopulated by Read)
- Add acceptance tests for `fleet_output` and `fleet_server_host` space-aware import
- Update OpenSpec requirement docs for all four specs whose import REQs change

## Capabilities

### New Capabilities

_(none — `SpaceImporter` is internal implementation scaffolding, not a user-visible capability)_

### Modified Capabilities

- `fleet-output`: REQ-008 currently documents `ImportStatePassthroughID` semantics; replace with composite ID import behavior matching `fleet-integration-policy` REQ-006
- `fleet-server-host`: REQ-007 currently documents `ImportStatePassthroughID` semantics; replace with composite ID import behavior
- `fleet-agent-download-source`: "Terraform import" requirement currently documents hardcoded `["default"]` fallback and dual `id`+`source_id` assignment; align with standard behavior (no `space_ids` set on plain ID, `source_id` only)
- `fleet-elastic-defend-integration-policy`: REQ-004 says "import passthrough semantics" but the implementation already supports composite IDs; update spec to document actual behavior

## Impact

- **Fixed resources**: `internal/fleet/output/resource.go`, `internal/fleet/serverhost/resource.go`
- **Migrated resources**: `internal/fleet/agentpolicy/resource.go`, `internal/fleet/integration_policy/resource.go`, `internal/fleet/elastic_defend_integration_policy/resource.go`, `internal/fleet/agentdownloadsource/resource.go`
- **New file**: `internal/fleet/space_importer.go`
- **Specs updated**: `fleet-output`, `fleet-server-host`, `fleet-agent-download-source`, `fleet-elastic-defend-integration-policy`
- **No breaking changes**: composite import ID was not previously supported for `output` and `server_host`, so adding it is purely additive; `agentdownloadsource` plain-ID behavior change is observable only in intermediate import state (Read converges correctly either way)
