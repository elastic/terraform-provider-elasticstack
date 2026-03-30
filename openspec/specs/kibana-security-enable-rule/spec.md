# `elasticstack_kibana_security_enable_rule` — Schema and Functional Requirements

Resource implementation: `internal/kibana/security_enable_rule`

## Purpose

Manage the enabled state of Elastic Security detection rules by tag in a Kibana space. The resource uses Kibana's bulk action API to enable all rules that match a specified tag key-value pair (e.g. `OS: Windows`). On destroy, the resource optionally disables those same rules. Drift is detected by querying whether all matching rules are currently enabled. Requires Elastic Stack 8.11.0 or higher.

## Schema

```hcl
resource "elasticstack_kibana_security_enable_rule" "example" {
  # Identity
  id    = <computed, string>   # composite: "<space_id>/<key>:<value>"; UseStateForUnknown

  # Targeting
  space_id = <optional, computed, string> # default "default"; RequiresReplace
  key      = <required, string>           # tag key to filter rules (e.g. "OS"); RequiresReplace
  value    = <required, string>           # tag value to filter rules (e.g. "Windows"); RequiresReplace

  # Behavior
  disable_on_destroy = <optional, computed, bool> # default true; whether to disable rules on destroy
  all_rules_enabled  = <computed, bool>           # true when all matching rules are currently enabled; default true
}
```

## Requirements

### Requirement: Kibana bulk action API — enable and disable

The resource SHALL enable all security detection rules matching the configured tag key-value pair by calling Kibana's bulk rules action API (`PerformRulesBulkAction`) with the `enable` action. On destroy, when `disable_on_destroy` is `true`, the resource SHALL call the same API with the `disable` action to disable those rules. The tag filter query sent to the API SHALL be in the form `alert.attributes.tags:("<key>: <value>")`.

#### Scenario: Enable on create and update

- GIVEN a plan with `key = "OS"` and `value = "Windows"` in `space_id = "my-space"`
- WHEN create or update runs
- THEN the resource SHALL call the bulk action API with action `enable` and query `alert.attributes.tags:("OS: Windows")` scoped to `my-space`

#### Scenario: Disable on destroy when disable_on_destroy is true

- GIVEN `disable_on_destroy = true` and an existing resource in state
- WHEN destroy runs
- THEN the resource SHALL call the bulk action API with action `disable` for the same tag filter

#### Scenario: Skip disable on destroy when disable_on_destroy is false

- GIVEN `disable_on_destroy = false`
- WHEN destroy runs
- THEN the resource SHALL NOT call the disable API and SHALL complete successfully

### Requirement: API error surfacing

When the bulk action API returns a non-200 HTTP status on enable or disable, the resource SHALL surface an error diagnostic. Transport-layer failures SHALL also produce error diagnostics.

#### Scenario: Non-200 on enable

- GIVEN the bulk action API returns a non-200 response during create or update
- WHEN the operation completes
- THEN Terraform SHALL receive an error diagnostic

### Requirement: Read — drift detection via FindRules

On read (refresh), the resource SHALL call Kibana's `FindRules` API filtered to rules that match the tag key-value pair AND are currently disabled. If that query returns a total of zero matching disabled rules, `all_rules_enabled` SHALL be set to `true`; otherwise it SHALL be set to `false`.

#### Scenario: All rules enabled

- GIVEN `FindRules` returns total=0 for disabled rules matching the tag
- WHEN read runs
- THEN state SHALL have `all_rules_enabled = true`

#### Scenario: Some rules disabled

- GIVEN `FindRules` returns total > 0 for disabled rules matching the tag
- WHEN read runs
- THEN state SHALL have `all_rules_enabled = false`

#### Scenario: API error on read

- GIVEN the `FindRules` API returns a non-200 response
- WHEN read runs
- THEN the resource SHALL surface an error diagnostic

### Requirement: Identity

After a successful create or update, the resource SHALL set `id` to the composite string `<space_id>/<key>:<value>`. The `id` attribute SHALL be preserved across reads using `UseStateForUnknown`.

#### Scenario: Composite id on create

- GIVEN `space_id = "default"`, `key = "OS"`, `value = "Windows"`
- WHEN create succeeds
- THEN `id` SHALL equal `"default/OS:Windows"`

### Requirement: Lifecycle — force replacement

Changing any of `space_id`, `key`, or `value` SHALL require destroying and recreating the resource rather than an in-place update.

#### Scenario: Replace on key change

- GIVEN an existing resource and a plan change to `key`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

### Requirement: Compatibility — minimum server version

On every CRUD operation, the resource SHALL obtain the Elastic Stack version from the provider and SHALL fail with an `Unsupported server version` error diagnostic if the server version is strictly below **8.11.0**, without proceeding to any Kibana API call.

#### Scenario: Server below 8.11.0

- GIVEN the target Elastic Stack is version 8.10.x or lower
- WHEN any CRUD operation runs
- THEN the resource SHALL fail with `Unsupported server version` and SHALL NOT call the bulk action or find-rules API

### Requirement: Provider configuration and Kibana client

On every CRUD operation, the resource SHALL use the provider's configured Kibana OAPI client. If the provider data cannot be converted to a valid API client, the resource SHALL return a configuration error diagnostic.

#### Scenario: Unconfigured provider

- GIVEN the provider has not supplied a usable API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

### Requirement: State mapping — all_rules_enabled on create and update

After a successful bulk enable API call, the resource SHALL set `all_rules_enabled` to `true` in state without an additional read call. The value SHALL reflect the intended post-enable state rather than requiring a round-trip query.

#### Scenario: State after successful enable

- GIVEN the bulk enable API call succeeds
- WHEN state is written after create or update
- THEN `all_rules_enabled` SHALL be `true`

### Requirement: State mapping — disable_on_destroy default

When `disable_on_destroy` is null in the plan at create/update time, the resource SHALL treat it as `true` and store `true` in state.

#### Scenario: Null disable_on_destroy

- GIVEN `disable_on_destroy` is not set in configuration
- WHEN create or update runs
- THEN state SHALL store `disable_on_destroy = true`
