## Why

`elasticstack_fleet_integration_policy` fails every apply with `Provider produced inconsistent result after apply` when `output_id` is not configured but the package policy has an output assigned outside Terraform, for example in the Kibana UI:

```
Error: Provider produced inconsistent result after apply
...produced an unexpected new value: .output_id: was null, but now cty.StringVal("...")
```

From Kibana 9.0.0 the Fleet update API treats an omitted `output_id` as "keep the existing value", so the response keeps returning the out-of-band value (earlier versions clear it on update). The resource copied that value into state, which contradicts the planned null. The same happens when the server value is an empty string (`cty.StringVal("")`). Before this change the same error also blocked removing a configured `output_id` from the configuration.

`elasticstack_fleet_agent_policy` already keeps unconfigured `data_output_id`, `monitoring_output_id`, `fleet_server_host_id` and `download_source_id` null for the same reason, and documents that clearing the server-side value is out of scope.

## What Changes

- When `output_id` is not configured (null), keep it null in state after create, update and read, whatever the API returns (an output ID or `""`).
- On import, populate `output_id` from the API response, since there is no configuration yet.
- Document on the `output_id` attribute that state remains null when it is not configured, and that removing it from configuration retains the server-side output on Kibana 9.0 and later, while earlier versions clear it on the next update.

Not breaking: no schema change. Configurations that previously failed now apply. A change to `output_id` made outside Terraform is no longer shown as drift while `output_id` is unset in configuration.

## Capabilities

### New Capabilities

_None — this fixes the state handling of an existing attribute._

### Modified Capabilities

- `fleet-integration-policy`: REQ-022 keeps an unconfigured `output_id` null in state, populates it on import, and states that removing `output_id` from configuration retains the server-side value on Kibana 9.0 and later (earlier versions clear it).

## Impact

- `internal/fleet/integration_policy/models.go`: `populateFromAPI` takes an `isImport` flag and only maps `output_id` from the API when it is configured or on import.
- `internal/fleet/integration_policy/create.go`, `update.go`, `read.go`: pass the `isImport` flag.
- `internal/fleet/integration_policy/schema.go`: `output_id` description.
- `internal/acctest`: shared helper for changing Fleet objects outside Terraform in acceptance tests.
- Tests: unit tests for `populateFromAPI` (configured, unconfigured, empty string, import); acceptance tests for an output set outside Terraform (including `""`), removal of a configured `output_id`, and import of a policy with an output; regression test for `elasticstack_fleet_agent_policy`.
