# `elasticstack_kibana_install_prebuilt_rules` — Schema and Functional Requirements

Resource implementation: `internal/kibana/prebuilt_rules`

## Purpose

Install and keep up to date Elastic prebuilt security detection rules and timelines in a Kibana space. On create and update the resource calls Kibana's install/update prebuilt rules API, then reads the resulting status counters (rules installed, rules not installed, rules not updated, timelines installed, timelines not installed, timelines not updated) and stores them in state. Delete is a no-op: removing the resource from Terraform does not uninstall the prebuilt rules from Kibana. Requires Elastic Stack 8.0.0 or higher.

## Schema

```hcl
resource "elasticstack_kibana_install_prebuilt_rules" "example" {
  # Identity
  id       = <computed, string>           # equal to space_id; UseStateForUnknown
  space_id = <optional, computed, string> # default "default"; RequiresReplace

  # Status counters (all computed, populated from API after install)
  rules_installed        = <computed, int64>
  rules_not_installed    = <computed, int64>
  rules_not_updated      = <computed, int64>
  timelines_installed    = <computed, int64>
  timelines_not_installed = <computed, int64>
  timelines_not_updated  = <computed, int64>
}
```

## Requirements

### Requirement: Kibana prebuilt rules install API

On create and update, the resource SHALL call Kibana's install/update prebuilt rules and timelines API (`InstallPrebuiltRulesAndTimelines`) for the configured `space_id`. A non-200 response from that API SHALL produce an error diagnostic and SHALL prevent the status read from being performed.

#### Scenario: Successful install

- GIVEN `space_id = "default"`
- WHEN create or update runs
- THEN the resource SHALL call `InstallPrebuiltRulesAndTimelines` scoped to `"default"`

#### Scenario: Install API error

- GIVEN the install API returns a non-200 response
- WHEN create or update runs
- THEN the resource SHALL surface an error diagnostic and SHALL NOT read the status counters

### Requirement: Post-install status read

After a successful install call, the resource SHALL call the prebuilt rules status API (`ReadPrebuiltRulesAndTimelinesStatus`) for the same `space_id` and SHALL populate all six counter attributes from the response. A non-200 response from the status API SHALL produce an error diagnostic.

#### Scenario: Status counters populated after install

- GIVEN the install API and the status API both succeed
- WHEN state is written after create or update
- THEN all six counter attributes SHALL reflect the values returned by the status API

### Requirement: Read — status refresh

On read (refresh), the resource SHALL call the prebuilt rules status API and update all six counter attributes in state from the response. The `id` used to scope the status request SHALL be the value stored in state for `id` (which equals `space_id`).

#### Scenario: Read updates counters

- GIVEN an existing resource in state with `id = "default"`
- WHEN read runs
- THEN the resource SHALL call the status API scoped to `"default"` and SHALL update all six counter attributes

#### Scenario: Status API error on read

- GIVEN the status API returns a non-200 response during read
- WHEN read runs
- THEN the resource SHALL surface an error diagnostic

### Requirement: Delete is a no-op

Removing the resource from Terraform state SHALL NOT call any Kibana API. Prebuilt rules already installed in Kibana SHALL remain in place after destroy.

#### Scenario: Destroy does not call API

- GIVEN an existing resource in state
- WHEN destroy runs
- THEN no Kibana API call SHALL be made

### Requirement: Identity

After a successful create or update, the resource SHALL set `id` to the value of `space_id`. The `id` attribute SHALL be preserved across reads using `UseStateForUnknown`.

#### Scenario: id equals space_id

- GIVEN `space_id = "my-space"`
- WHEN create or update succeeds
- THEN `id` SHALL equal `"my-space"`

### Requirement: Lifecycle — force replacement on space_id

Changing `space_id` SHALL require destroying and recreating the resource.

#### Scenario: Replace on space_id change

- GIVEN an existing resource and a plan that changes `space_id`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create)

### Requirement: Compatibility — minimum server version

On create and update, the resource SHALL obtain the Elastic Stack version from the provider and SHALL fail with an `Unsupported server version` error diagnostic if the server version is strictly below **8.0.0**.

#### Scenario: Server below 8.0.0

- GIVEN the target Elastic Stack is below version 8.0.0
- WHEN create or update runs
- THEN the resource SHALL fail with `Unsupported server version` and SHALL NOT call the install API

### Requirement: Provider configuration and Kibana client

On create, read, and update, the resource SHALL use the provider's configured Kibana OAPI client. If the provider data cannot be converted to a valid API client, the resource SHALL return a configuration error diagnostic.

#### Scenario: Unconfigured provider

- GIVEN the provider has not supplied a usable API client
- WHEN create, read, or update runs
- THEN the operation SHALL fail with a provider configuration error

### Requirement: Plan modification — normalize counters when resource already exists

When the resource already exists (the plan `id` is a known value) and the planned values for `rules_not_installed`, `rules_not_updated`, `timelines_not_installed`, or `timelines_not_updated` are greater than zero, the resource SHALL normalize the plan at `ModifyPlan` time by setting those counters to zero and adding the uninstalled counts to the corresponding installed counters, so that Terraform does not show a spurious diff.

#### Scenario: Not-installed rules normalized in plan

- GIVEN an existing resource with `rules_not_installed = 5` and `rules_installed = 10` in the plan
- WHEN `ModifyPlan` runs
- THEN the plan SHALL set `rules_not_installed = 0` and `rules_installed = 15`

#### Scenario: Not-updated timelines normalized in plan

- GIVEN an existing resource with `timelines_not_updated = 3` in the plan
- WHEN `ModifyPlan` runs
- THEN the plan SHALL set `timelines_not_updated = 0`

#### Scenario: No normalization during create

- GIVEN the plan `id` is unknown (resource is being created)
- WHEN `ModifyPlan` runs
- THEN the counters SHALL NOT be modified by the plan modifier
