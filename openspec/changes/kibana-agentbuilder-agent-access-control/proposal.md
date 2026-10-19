## Why

The Kibana Agent Builder agents API (`POST /api/agent_builder/agents` and `PUT /api/agent_builder/agents/{id}`) accepts an optional `access_control` object controlling who can read and write an agent (`access_mode`: `public` | `shared` | `private`). `elasticstack_kibana_agentbuilder_agent` never sends or reads this field — `internal/kibana/agentbuilderagent` (`schema.go`, `models.go`, `create.go`, `update.go`) has no `access_control`/`AccessControl` references today, even though the generated client already carries the field on both write bodies (`PostAgentBuilderAgentsJSONBody`, `PutAgentBuilderAgentsIdJSONBody`). This is a user-controllable permissions setting, not server-managed state, so it belongs in Terraform ([issue #5377](https://github.com/elastic/terraform-provider-elasticstack/issues/5377)).

This closes the `kibana-api-gap` finding for `access_control` on both the create and update request bodies.

## What Changes

- Add an optional, computed `access_control` single-nested attribute to the `elasticstack_kibana_agentbuilder_agent` resource schema, containing one required string attribute `access_mode` (enum `public` | `shared` | `private`).
- Gate the entire `access_control` attribute behind Elastic Stack 9.5.0+ via a field-conditional version requirement (modeled on `internal/kibana/alertingrule/models.go`'s per-field `GetVersionRequirements` pattern): configuring `access_control` on a stack below 9.5.0 SHALL produce a version-requirement error; omitting it SHALL be unaffected on any supported stack.
- On create and update, send `access_control.access_mode` to the API when configured.
- On update, omit `access_control` from the request body when it is absent from **configuration** (not just absent from a computed-carrying plan), so that removing the block from HCL relinquishes management of the mode rather than resetting it to a specific value.
- On read, populate `access_control.access_mode` by calling the dedicated `GET /api/agent_builder/agents/{id}/access_control` sub-resource endpoint (`x-state: Generally available; added in 9.5.0`) — not the main agent `GET`, whose behavior for this field is unconfirmed on 9.4.x — gated behind the same 9.5.0 capability check already used for `MinExtendedAPIVersion` in `create.go`/`update.go`. Below 9.5.0, `access_control` is left null in state and this additional call is never attempted.
- Add a new version constant (alongside `MinCoreAPIVersion` and `MinExtendedAPIVersion` in `internal/kibana/agentbuilder/versions.go`) for the Agent Builder access-control sub-resource API.
- Add a thin `internal/clients/kibanaoapi` wrapper for the dedicated access-control `GET` endpoint, next to the existing `GetAgent`/`CreateAgent`/`UpdateAgent` wrappers.

## Capabilities

### Modified Capabilities

- `kibana-agentbuilder-agent`: adds the `access_control` attribute (schema, version gate, write mapping, and dedicated-endpoint read-back) to the existing resource requirements.

### New Capabilities

<!-- None — this change extends the existing kibana-agentbuilder-agent capability only. -->

## Impact

- **Modified code**: `internal/kibana/agentbuilderagent/{schema.go,models.go,create.go,update.go,read.go}` (read-after-write + refresh paths that currently call `kibanaoapi.GetAgent` need a conditional second call), `internal/clients/kibanaoapi/agentbuilder_agent.go` (new wrapper), `internal/kibana/agentbuilder/versions.go` (new version constant).
- **Acceptance tests**: `internal/kibana/agentbuilderagent/acc_test.go` plus new `testdata`/config fixtures for `access_control`, gated at the new minimum version via `versionutils.SkipIfUnsupported`.
- **Docs/examples**: `docs/resources/kibana_agentbuilder_agent.md` and `examples/resources/elasticstack_kibana_agentbuilder_agent/` need regeneration/updates once implemented (not part of this proposal-only change).
- **Backward compatibility**: additive only. Existing configurations without `access_control` are unaffected on every supported stack version; the new attribute is Optional so no existing `.tf` file needs to change.
- **Data source parity**: explicitly out of scope for this change (see design.md Non-Goals); the `elasticstack_kibana_agentbuilder_agent` data source is not touched.
