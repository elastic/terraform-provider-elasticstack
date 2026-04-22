## Why

The current `fleet-integration-policy` capability is intentionally built around Fleet's simplified mapped package policy shape: package-level vars, map-keyed `inputs`, and map-keyed `streams`. That design matches most integrations and keeps the generic resource understandable, but Elastic Defend uses a different request/response format centered on a typed `inputs` list and `config` payloads such as `integration_config`, `artifact_manifest`, and `policy`.

Adding Elastic Defend support to `elasticstack_fleet_integration_policy` would force the generic resource to carry two incompatible internal models, reopen the generated-client normalization that currently prefers the simplified format, and complicate defaults, secret handling, state mapping, and acceptance coverage for one special-case integration.

## What Changes

- Add a dedicated `fleet-elastic-defend-integration-policy` capability for a new `elasticstack_fleet_elastic_defend_integration_policy` resource.
- Update the shared `generated/kbapi` Fleet package policy client to support both mapped and typed input encodings, including typed-input `type` and `config` fields plus the top-level package policy `version` used for Defend updates, following the same direction explored in [PR #1500](https://github.com/elastic/terraform-provider-elasticstack/pull/1500).
- Keep the resource schema aligned with the existing integration policy resource for common package-policy envelope fields such as identity, naming, namespace, enablement, agent policy attachment, and package versioning.
- Model only Defend-owned configuration in Terraform using typed attributes and nested attributes, instead of exposing generic `vars_json`, generic `inputs`, or arbitrary raw package policy payloads.
- Treat server-managed Defend payloads such as `artifact_manifest` and update concurrency tokens as opaque implementation details handled internally by the provider rather than user-configurable schema.
- Preserve `fleet-integration-policy` as the generic mapped-input resource even after the shared client supports both encodings.
- Restrict `fleet_elastic_defend_integration_policy` to the typed-input encoding used by Elastic Defend.

## Capabilities

### New Capabilities

- `fleet-elastic-defend-integration-policy`: define the schema and runtime behavior for the dedicated Elastic Defend integration policy resource

### Modified Capabilities

<!-- None. -->

## Impact

- `openspec/changes/add-fleet-elastic-defend-integration-policy/specs/fleet-elastic-defend-integration-policy/spec.md`
- `generated/kbapi`
- `generated/kbapi/transform_schema.go`
- `internal/clients/fleet/fleet.go`
- `internal/fleet/elastic_defend_integration_policy`
- Fleet provider registration and generated documentation for the new resource
- Acceptance and unit coverage for mapped-vs-typed Fleet package policy handling, plus Defend-specific request mapping, readback, import, and update behavior
