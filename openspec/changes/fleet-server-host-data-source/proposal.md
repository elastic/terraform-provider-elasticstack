## Why

There is no way to look up an existing Fleet server host from Terraform. When one stack owns the Fleet server host (managed with `elasticstack_fleet_server_host`) and another stack needs its id — for `elasticstack_fleet_agent_policy.fleet_server_host_id`, or for a module that deploys Elastic Agent — the consuming stack has to hardcode the UUID or read the owning stack's state with `terraform_remote_state`.

The Fleet data sources available today don't cover this. `elasticstack_fleet_agent_policy` exposes a policy's `fleet_server_host_id`, but only for a policy that sets the host explicitly, and it couples the consumer to an unrelated policy. `elasticstack_fleet_output` lists outputs; there is no equivalent for Fleet server hosts.

This change adds a read-only `elasticstack_fleet_server_host` data source that resolves one Fleet server host by `host_id`, by `name`, or as the default host, closing issue #5423. It follows the `elasticstack_fleet_agent_policy` data source (#5369).

## What Changes

- Add an `elasticstack_fleet_server_host` data source backed by the Fleet Get Fleet Server host API (`GET /api/fleet/fleet_server_hosts/{itemId}`) and the Fleet Get Fleet Server hosts API (`GET /api/fleet/fleet_server_hosts`).
- The data source accepts exactly one of `host_id`, `name` or `default = true` as the lookup selector. All three are `Optional+Computed`, so the ones not used as the selector are populated from the API response. `default = false` is rejected at plan time: it has no meaning as a selector.
- The list API takes no filter or paging parameters, so `name` and `default` lookups read the whole list and filter it in the provider. Zero matches is an error; more than one match is an error that lists the matching `host_id`s. A response that reports more hosts than it returns is an error, so a truncated list is never treated as complete.
- Exported attributes mirror the `elasticstack_fleet_server_host` resource: `host_id`, `name`, `hosts`, `default`, `proxy_id` and `ssl` (`certificate`, `certificate_authorities`, `client_auth`). The private key `ssl.key` is not part of the data source schema.
- An optional `space_id` selects the Kibana space the request is sent through. Fleet server hosts are shared by all spaces, so it does not filter the lookup.

Not breaking: additive only.

## Capabilities

### New Capabilities

_None — this extends the existing `fleet-server-host` capability._

### Modified Capabilities

- `fleet-server-host`: adds the read-only `elasticstack_fleet_server_host` data source — schema, selector validation, lookup by `host_id`, by `name` and by `default`, not-found and ambiguous-match errors, and attribute mapping — to the capability that already covers the `elasticstack_fleet_server_host` resource.

## Impact

- New package `internal/fleet/serverhostds` (`data_source.go`, `schema.go`, `models.go`, `read.go`), built on `entitycore.NewKibanaDataSource`.
- New client helper `ListFleetServerHosts` in `internal/clients/fleet/fleet_server_host.go`.
- Provider registration in `provider/plugin_framework.go`.
- Docs and example: `docs/data-sources/fleet_server_host.md`, `examples/data-sources/elasticstack_fleet_server_host/data-source.tf`. The example is on the `TestAccExamples_planOnly` skip list because its lookups need hosts that already exist.
- Tests: unit tests for the client helper and the read function; acceptance tests in `internal/fleet/serverhostds/acc_test.go`, and the lookup by `default` in `internal/fleet/serverhost/acc_test.go`, next to the resource test that also changes the default host.
- The `elasticstack_fleet_server_host` resource is not changed.

## Out of Scope

- The `agent_*` and `es_*` SSL fields and the SSL key references under the API's `secrets` object, which the resource does not manage either.
- `space_ids`: the Fleet API does not return it for server hosts.
- A list-typed data source returning several Fleet server hosts.
