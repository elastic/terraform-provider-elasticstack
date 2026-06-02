## ADDED Requirements

### Requirement: `host_id` plan modifiers (REQ-015)

The `host_id` attribute SHALL carry `UseStateForUnknown()` and `RequiresReplace()` plan modifiers.

`UseStateForUnknown()` ensures that when `host_id` is absent from the practitioner's configuration, the Plugin Framework carries the prior-state value (the Fleet-assigned UUID) into the update plan so that the `Update` handler receives the correct resource identifier. Without this modifier, `host_id` is `null` in the plan, causing the `Update` handler to construct a request URL with an empty path segment and Kibana to return 404.

`RequiresReplace()` ensures that if a practitioner explicitly provides a new `host_id` value that differs from the one in state, the resource is destroyed and recreated rather than attempting an in-place update with a mismatched ID.

Both modifiers align `fleet_server_host.host_id` with the identical pattern already applied to every other fleet resource ID attribute (`fleet_output.output_id`, `fleet_agent_policy.policy_id`, `fleet_proxy.proxy_id`, `fleet_agent_download_source.source_id`).

#### Scenario: Update succeeds when host_id is not configured

- GIVEN a `fleet_server_host` resource created with `host_id` omitted from config (computed by Fleet)
- WHEN the practitioner changes `name` or `hosts` and runs `terraform apply`
- THEN the `Update` handler SHALL receive the prior-state `host_id` value in the plan (not null)
- AND the Fleet update API SHALL be called with the correct resource identifier
- AND the apply SHALL succeed without a 404 error

#### Scenario: Explicit host_id change triggers replacement

- GIVEN a `fleet_server_host` resource with a known `host_id` in state
- WHEN the practitioner sets a different explicit `host_id` value in config
- THEN Terraform SHALL plan a destroy-and-recreate rather than an in-place update

### Requirement: Acceptance test UPDATE coverage for computed host_id (REQ-016)

The acceptance test `TestAccResourceFleetServerHost_computedID` SHALL include at least one UPDATE `resource.TestStep` that:

1. Changes a mutable attribute (`name` or `hosts`) while omitting `host_id` from the test config.
2. Asserts that the apply succeeds.
3. Asserts that `host_id` remains set to a non-empty value in state after the update.

This guards the `UseStateForUnknown()` fix against future regressions.

#### Scenario: Acceptance test UPDATE step passes with computed host_id

- GIVEN a `fleet_server_host` acceptance test that created a host without an explicit `host_id`
- WHEN an UPDATE step is applied that changes `name` or `hosts` and still omits `host_id`
- THEN the apply SHALL succeed and state SHALL contain a non-empty `host_id`
