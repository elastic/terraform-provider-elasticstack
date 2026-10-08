## Why

`elasticstack_fleet_server_host` has no way to link a Fleet server host to a `elasticstack_fleet_proxy`. Operators who route each Fleet Server host through a per-client proxy must currently manage that link outside Terraform, even though the Kibana Fleet API already accepts `proxy_id` on both create and update of fleet server hosts, and the sibling `elasticstack_fleet_proxy` resource already fully supports `proxy_id`. This closes [#5126](https://github.com/elastic/terraform-provider-elasticstack/issues/5126).

## What Changes

- Add an optional `proxy_id` (string) attribute to `elasticstack_fleet_server_host`, sent on create and update, and read back into state.
- On update, mirror the `proxy_id` unset-handling fix already shipped in `elasticstack_fleet_agent_download_source`: the generated `PutFleetFleetServerHostsItemidJSONBody.ProxyId` field carries `json:"proxy_id,omitempty"`, so sending `nil` when a previously-set `proxy_id` is cleared in config would be dropped from the request body entirely and (per the precedent documented in `agentdownloadsource/models.go`) risks being treated by Fleet as "leave unchanged" rather than "clear it" — producing a Terraform "inconsistent result after apply" error. The resource will instead send an explicit empty string (`""`) in that case, and `nil` only when `proxy_id` was never set, using the same `prior`-aware helper pattern as `agentdownloadsource.proxyIDForUpdate`.
- Update `examples/resources/elasticstack_fleet_server_host/resource.tf` to show `proxy_id` wired to an `elasticstack_fleet_proxy` resource, matching the issue's requested HCL shape.
- Add an acceptance test mirroring `TestAccResourceFleetAgentDownloadSource_ProxyID` that sets, updates, and unsets `proxy_id` on a `fleet_server_host`, wiring in a real `elasticstack_fleet_proxy`.

## Capabilities

### Modified Capabilities

- `fleet-server-host`: Adds the `proxy_id` attribute to the resource schema, create/update API body mapping (including the unset-to-empty-string update semantics), and state mapping.

## Impact

- **Code**: `internal/fleet/serverhost/schema.go` (new attribute), `internal/fleet/serverhost/models.go` (model field, create/update body mapping, `populateFromAPI`), `internal/fleet/serverhost/update.go` (`toAPIUpdateModel` signature grows a `prior serverHostModel` parameter so the unset-to-empty-string rule has the prior value to compare against, matching `agentdownloadsource`'s `req.Prior`-aware call site).
- **Generated clients**: none — `generated/kbapi` already carries `ProxyId *string` on `ServerHost`, `PostFleetFleetServerHostsJSONBody`, and `PutFleetFleetServerHostsItemidJSONBody`. No regeneration needed.
- **Docs/examples**: `examples/resources/elasticstack_fleet_server_host/resource.tf` updated; provider docs regenerated from the schema description in a later implementation pass.
- **Tests**: new acceptance test in `internal/fleet/serverhost/acc_test.go` plus associated `testdata/` fixtures (set / update / unset of `proxy_id` against a real `elasticstack_fleet_proxy`).
- **Backward compatibility**: additive only. `proxy_id` is optional with no default; existing configurations without it are unaffected.
- **Out of scope**: changes to `elasticstack_fleet_proxy` itself, `generated/kbapi` regeneration, and `proxy_id` support on other Fleet sub-resources not named in the issue.
