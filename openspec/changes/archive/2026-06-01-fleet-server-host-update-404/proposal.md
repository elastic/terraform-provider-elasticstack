## Why

When a practitioner omits `host_id` from an `elasticstack_fleet_server_host` configuration (the common case, since the attribute is `Optional+Computed` and Fleet auto-assigns the UUID), any subsequent `terraform apply` that changes `name`, `hosts`, or another attribute returns `{"statusCode":404,"error":"Not Found","message":"Not Found"}` from Kibana. Destroy runs in the same configuration exhibit the same failure.

The root cause is a missing `UseStateForUnknown()` plan modifier on `host_id`. Without it, the Terraform Plugin Framework leaves `host_id` as `null` in the update plan when the user has not configured the attribute. The resource's `Update` handler then calls `planModel.HostID.ValueString()`, gets an empty string, and constructs a request URL with an empty path segment (`PUT /api/fleet/fleet_server_hosts/`), which Kibana returns 404 for.

Every other equivalent fleet resource already carries the correct plan modifiers on its ID attribute (`fleet_output.output_id`, `fleet_agent_policy.policy_id`, `fleet_proxy.proxy_id`, `fleet_agent_download_source.source_id`). The `fleet_server_host.host_id` attribute is the sole outlier.

A secondary gap exists in the acceptance test suite: the test that covers API-assigned `host_id` (`TestAccResourceFleetServerHost_computedID`) only exercises the CREATE step. An UPDATE step is absent, meaning the 404 regression was not caught by CI.

## What Changes

- Add `UseStateForUnknown()` and `RequiresReplace()` plan modifiers to `host_id` in `internal/fleet/serverhost/schema.go`, matching the identical pattern on all other fleet ID attributes.
- Extend `TestAccResourceFleetServerHost_computedID` with an UPDATE step (changing `name` or `hosts` after creation with no explicit `host_id` in config) to guard the fixed path against future regressions.

## Capabilities

### New Capabilities

<!-- None — this change fixes a bug in an existing capability. -->

### Modified Capabilities

- `fleet-server-host`: Add plan modifier requirements for `host_id` (REQ-015) and acceptance test UPDATE coverage (REQ-016).

## Impact

- **`internal/fleet/serverhost/schema.go`**: Add `PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}` to the `host_id` attribute.
- **`internal/fleet/serverhost/*_test.go`** (or equivalent acceptance test file): Extend `TestAccResourceFleetServerHost_computedID` with an UPDATE step.
- **Backward compatibility**: `UseStateForUnknown()` is behavior-preserving for users who do not set `host_id`. `RequiresReplace()` changes behavior only when a user explicitly provides a different `host_id` — previously a broken update, now a safe destroy-and-recreate.
