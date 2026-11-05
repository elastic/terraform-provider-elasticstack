# `elasticstack_kibana_alerting_v2_action_policy` — Schema and Functional Requirements

Resource implementation: `internal/kibana/alertingv2actionpolicy`

## Purpose

Define the Terraform schema and runtime behaviour of the `elasticstack_kibana_alerting_v2_action_policy` resource, which manages a single Kibana Alerting v2 action policy through `/api/alerting/v2/action_policies`: typed destinations, matcher, grouping, and throttle configuration, create-or-replace lifecycle, space-aware identity and import, enabled-state handling, plan-time validation mirroring the API, and the Elastic Stack 9.6.0 version gate.

## Schema

```hcl
resource "elasticstack_kibana_alerting_v2_action_policy" "example" {
  # Identity
  id               = <computed, string>            # composite "<space_id>/<action_policy_id>"; UseStateForUnknown
  action_policy_id = <optional, computed, string>  # 1-150 chars, letters/digits/underscore/hyphen; RequiresReplace; UseStateForUnknown; generated when omitted
  space_id         = <optional, computed, string>  # default "default"; RequiresReplace

  # Policy definition
  name        = <required, string>          # 1-256 chars
  description = <optional, string>          # 1-1024 chars
  enabled     = <optional, computed, bool>  # default true

  destinations = [                          # required; 1-10 items; each (type, id) pair unique
    {
      type = <required, string>             # "workflow"
      id   = <required, string>             # 1-150 chars; the workflow identifier
    },
  ]

  matcher = {                               # optional; omit for a catch-all policy; at least one of tags / expression
    tags       = <optional, list(string)>   # 1-50 items, each 1-128 chars; matched against rule metadata.routing_tags
    expression = <optional, string>         # KQL, 1-4096 chars
  }

  grouping = {                              # optional; absent means per_alert
    mode   = <required, string>             # "per_alert" | "all" | "per_field"
    fields = <optional, list(string)>       # 1-16 items, each 1-256 chars; required when mode = "per_field", not allowed otherwise
  }

  throttle = {                              # optional
    strategy = <required, string>           # "on_status_change" | "per_status_interval" | "time_interval" | "every_time"
    interval = <optional, duration string>  # required when strategy is per_status_interval or time_interval, not allowed otherwise
  }

  # Server-owned
  created_at = <computed, string>  # ISO 8601; UseStateForUnknown
  updated_at = <computed, string>  # ISO 8601
}
```

Notes:

- `kibana_connection` and `timeouts` are injected by the Kibana resource envelope and are not part of the resource's own schema.
- A duration string is an unsigned integer followed by one of `ms`, `s`, `m`, `h`, `d`, or `w`, at most 365 days in total.
- Snooze state (`snoozed_until`) is runtime operational state and is not exposed or managed.
- `type` is an enum with a single value today. It is kept as an explicit attribute so new destination types can be added without a schema change.

## ADDED Requirements

### Requirement: Resource type and registration

The provider SHALL expose a managed resource with type name `elasticstack_kibana_alerting_v2_action_policy`, registered in the provider's standard resource set, with generated registry documentation under `docs/resources/` and a usage example under `examples/resources/`.

#### Scenario: Resource is available without opt-in

- **GIVEN** a provider configured with Kibana credentials and no experimental environment variable set
- **WHEN** a configuration declares `elasticstack_kibana_alerting_v2_action_policy`
- **THEN** Terraform SHALL recognise the resource type
- **AND** `make docs-generate` SHALL produce `docs/resources/kibana_alerting_v2_action_policy.md`

### Requirement: Minimum Elastic Stack version

The resource SHALL require Elastic Stack 9.6.0 or later and SHALL fail every lifecycle operation with a diagnostic naming the required version when the target stack is older. Serverless targets SHALL bypass the version check.

#### Scenario: Older stack is rejected before any API call

- **GIVEN** a target Elastic Stack at version 9.5.x
- **WHEN** create, read, update, or delete runs
- **THEN** the provider SHALL return an error diagnostic stating that Alerting v2 action policies require Elastic Stack 9.6.0 or later
- **AND** SHALL NOT call any `/api/alerting/v2/` endpoint

#### Scenario: Supported stack proceeds

- **GIVEN** a target Elastic Stack at version 9.6.0 or later, or a serverless project
- **WHEN** any lifecycle operation runs
- **THEN** the provider SHALL proceed with the operation

### Requirement: CRUD endpoint mapping

The resource SHALL manage an action policy with `PUT /api/alerting/v2/action_policies/{id}` for both create and update, `GET /api/alerting/v2/action_policies/{id}` for read, and `DELETE /api/alerting/v2/action_policies/{id}` for delete, each addressed to the space in `space_id`. It SHALL NOT call `POST /api/alerting/v2/action_policies`, `PATCH`, bulk, snooze, API-key, or history endpoints. Final state after create and update SHALL come from a subsequent read.

#### Scenario: Create issues a PUT with the full body and reads back

- **GIVEN** a new resource configuration
- **WHEN** create runs
- **THEN** the provider SHALL issue one `PUT /api/alerting/v2/action_policies/{id}` whose body is derived entirely from the plan
- **AND** SHALL then issue `GET /api/alerting/v2/action_policies/{id}` and derive state from that response

#### Scenario: Update replaces rather than patches

- **GIVEN** an existing managed action policy with a changed attribute
- **WHEN** update runs
- **THEN** the provider SHALL issue one `PUT` with the full replacement body
- **AND** SHALL NOT issue a `PATCH` request

#### Scenario: Requests target the configured space

- **GIVEN** `space_id = "team-a"`
- **WHEN** any lifecycle operation runs
- **THEN** the request path SHALL be prefixed with `/s/team-a`

#### Scenario: Default space uses the unprefixed path

- **GIVEN** `space_id` omitted or `"default"`
- **WHEN** any lifecycle operation runs
- **THEN** the request path SHALL NOT carry an `/s/` prefix

### Requirement: Create refuses to adopt an existing action policy

Before the first write, create SHALL check whether an action policy with the target id already exists in the space (`GET` answering 200) and, if so, SHALL return an error diagnostic naming the id and pointing at `terraform import` without writing. If the `PUT` nevertheless answers HTTP 200 (replaced) instead of 201 (created), the provider SHALL return the same diagnostic and state that the replacement has already happened server-side.

#### Scenario: Create against a pre-existing id

- **GIVEN** an action policy `oncall` already exists in the target space outside Terraform
- **AND** a new resource with `action_policy_id = "oncall"`
- **WHEN** create runs
- **THEN** the provider SHALL return an error diagnostic naming the id and recommending import
- **AND** SHALL NOT issue a `PUT`
- **AND** the resource SHALL NOT be recorded in state

#### Scenario: Race between existence check and write

- **GIVEN** a new resource whose id was free at the existence check
- **WHEN** the `PUT` answers HTTP 200 because the policy appeared in between
- **THEN** the provider SHALL return an error diagnostic naming the id, recommending import, and stating the policy was replaced

#### Scenario: Normal create

- **GIVEN** a new resource whose id is not in use
- **WHEN** create runs and the API answers HTTP 201
- **THEN** the provider SHALL continue to the read-after-write and record the resource in state

### Requirement: Identity and `action_policy_id` generation

`id` SHALL be `<space_id>/<action_policy_id>`. `action_policy_id` SHALL be optional and computed: when supplied it SHALL be used verbatim. When omitted the provider SHALL generate a UUID before the first write and persist it. A configured value SHALL be 1-150 characters of letters, digits, underscores, or hyphens. Changing `action_policy_id` or `space_id` SHALL require replacement.

#### Scenario: Generated id is stable

- **GIVEN** a configuration omitting `action_policy_id`
- **WHEN** the resource is applied and then re-planned without configuration changes
- **THEN** `action_policy_id` in state SHALL be unchanged
- **AND** the plan SHALL be empty

#### Scenario: Invalid id rejected at plan time

- **GIVEN** `action_policy_id = "on call"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic describing the permitted characters and length

### Requirement: Import

The resource SHALL support `terraform import` with an id of the form `<space_id>/<action_policy_id>`. On success it SHALL populate `space_id`, `action_policy_id`, and `id`, and the following read SHALL populate every other attribute. A malformed import id SHALL produce an error diagnostic naming the required format.

#### Scenario: Import round-trip

- **GIVEN** an action policy `oncall` in space `team-a`
- **WHEN** `terraform import` runs with id `team-a/oncall`
- **THEN** state SHALL hold `space_id = "team-a"`, `action_policy_id = "oncall"`, `id = "team-a/oncall"`
- **AND** a subsequent plan against a matching configuration SHALL be empty

#### Scenario: Malformed import id

- **GIVEN** an import id without a `/` separator or with an empty segment
- **WHEN** import runs
- **THEN** the provider SHALL return an error diagnostic describing the `<space_id>/<action_policy_id>` format

### Requirement: `enabled` is written in the create-or-replace body

`enabled` SHALL be optional, computed, and default to `true`. The provider SHALL always send `enabled` in the `PUT` body so the policy is created in, or replaced into, the configured state in one request, with no follow-up enable or disable call. This depends on Kibana accepting `enabled` in `PUT /api/alerting/v2/action_policies/{id}`, which is an upstream prerequisite for implementing this resource.

#### Scenario: Create a disabled policy

- **GIVEN** a new resource with `enabled = false`
- **WHEN** create runs
- **THEN** the `PUT` body SHALL contain `"enabled": false`
- **AND** the read-after-write SHALL show `enabled = false`
- **AND** no `_disable` request SHALL be sent

#### Scenario: Toggle enabled on update

- **GIVEN** a managed policy with `enabled = false`
- **WHEN** configuration sets `enabled = true` and update runs
- **THEN** the `PUT` body SHALL contain `"enabled": true`
- **AND** state SHALL hold `enabled = true`

#### Scenario: Omitted enabled defaults to true

- **GIVEN** a configuration omitting `enabled`
- **WHEN** create runs
- **THEN** the plan SHALL show `enabled = true` and the `PUT` body SHALL contain `"enabled": true`

### Requirement: Destinations mapping

`destinations` SHALL be a required ordered list of 1-10 objects, each with required `type` (currently only `workflow`) and `id` (1-150 chars). The same `(type, id)` pair SHALL NOT appear twice. Each entry SHALL map to and from the API's type-discriminated destination object.

#### Scenario: Workflow destination round-trip

- **GIVEN** `destinations = [{ type = "workflow", id = "wf-notify-oncall" }]`
- **WHEN** the policy is applied and read back
- **THEN** state SHALL hold one destination with the same `type` and `id`

#### Scenario: Empty destinations rejected at plan time

- **GIVEN** `destinations = []`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating at least one destination is required

#### Scenario: Duplicate destination rejected at plan time

- **GIVEN** two destinations with `type = "workflow"` and `id = "wf-1"`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic

#### Scenario: Unknown destination type rejected at plan time

- **GIVEN** `destinations = [{ type = "email", id = "x" }]`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic listing the supported types

### Requirement: Matcher mapping

`matcher` SHALL be an optional object with optional `tags` (1-50 items, each 1-128 chars) and optional `expression` (KQL, 1-4096 chars). When `matcher` is set, at least one of `tags` and `expression` SHALL be set. Omitting `matcher` SHALL produce a catch-all policy and SHALL omit the key from the request. The provider SHALL NOT parse or validate KQL.

#### Scenario: Catch-all policy

- **GIVEN** a configuration omitting `matcher`
- **WHEN** create runs and the policy is read back
- **THEN** the `PUT` body SHALL NOT contain `matcher`
- **AND** state SHALL hold `matcher` null

#### Scenario: Tags and expression round-trip

- **GIVEN** `matcher = { tags = ["prod", "db"], expression = "data.host.name : web-*" }`
- **WHEN** the policy is applied and read back
- **THEN** state SHALL hold the same `tags` in order and the same `expression`

#### Scenario: Empty matcher rejected at plan time

- **GIVEN** `matcher = {}`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating at least one of `tags` and `expression` is required, or `matcher` must be omitted

### Requirement: Grouping mapping

`grouping` SHALL be an optional object with a required `mode` of `per_alert`, `all`, or `per_field`, and optional `fields` (1-16 items, each 1-256 chars) that SHALL be required when `mode = "per_field"` and not allowed otherwise. The object SHALL map to and from the API's mode-discriminated grouping body. Omitting `grouping` SHALL omit the key from the request.

#### Scenario: Per-field grouping round-trip

- **GIVEN** `grouping = { mode = "per_field", fields = ["host.name"] }`
- **WHEN** the policy is applied and read back
- **THEN** state SHALL hold the same `mode` and `fields`

#### Scenario: Fields with a fieldless mode rejected at plan time

- **GIVEN** `grouping = { mode = "all", fields = ["host.name"] }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `fields` is only allowed with `mode = "per_field"`

#### Scenario: Per-field without fields rejected at plan time

- **GIVEN** `grouping = { mode = "per_field" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `fields` is required for `mode = "per_field"`

### Requirement: Throttle mapping and mode compatibility

`throttle` SHALL be an optional object with a required `strategy` of `on_status_change`, `per_status_interval`, `time_interval`, or `every_time`, and an optional duration `interval` that SHALL be required for `per_status_interval` and `time_interval` and not allowed otherwise. When `grouping.mode` is `per_alert` or `grouping` is omitted, `strategy` SHALL be one of `on_status_change`, `per_status_interval`, `every_time`. When it is `all` or `per_field`, `strategy` SHALL be `time_interval` or `every_time`.

#### Scenario: Interval throttle round-trip

- **GIVEN** `grouping = { mode = "all" }` and `throttle = { strategy = "time_interval", interval = "15m" }`
- **WHEN** the policy is applied and read back
- **THEN** state SHALL hold the same `strategy` and `interval` exactly as written

#### Scenario: Interval with a non-interval strategy rejected at plan time

- **GIVEN** `throttle = { strategy = "on_status_change", interval = "5m" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `interval` is only allowed with interval strategies

#### Scenario: Missing interval rejected at plan time

- **GIVEN** `throttle = { strategy = "time_interval" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic stating `interval` is required

#### Scenario: Strategy incompatible with grouping mode rejected at plan time

- **GIVEN** `grouping = { mode = "all" }` and `throttle = { strategy = "on_status_change" }`
- **WHEN** Terraform validates the configuration
- **THEN** the provider SHALL return a validation diagnostic naming the allowed strategies for `mode = "all"`

#### Scenario: Unknown grouping mode suppresses the compatibility check

- **GIVEN** `grouping.mode` is unknown at plan time
- **WHEN** the throttle compatibility check runs
- **THEN** the check SHALL be skipped

### Requirement: Collections and optional objects round-trip

Optional objects (`matcher`, `grouping`, `throttle`) and `description` SHALL map null in configuration to absent in the request, and absent in the response to null in state. Removing one from configuration SHALL remove it from the policy on the next apply because the write is a full replacement. Lists SHALL be stored in the order returned by the API.

#### Scenario: Removing throttle clears it server-side

- **GIVEN** a managed policy applied with `throttle`, and a new configuration omitting `throttle`
- **WHEN** update runs
- **THEN** the `PUT` body SHALL NOT contain `throttle`
- **AND** the read-after-write SHALL show `throttle` null
- **AND** the following plan SHALL be empty

#### Scenario: Absent description reads back as null

- **GIVEN** a policy applied without `description`
- **WHEN** state is refreshed
- **THEN** `description` SHALL be null

### Requirement: Server-owned fields

`created_at` and `updated_at` SHALL be computed and never sent. `created_at` SHALL keep its state value when the plan is otherwise unchanged. `updated_at` SHALL be planned as unknown whenever another attribute changes and SHALL keep its prior value when the plan is otherwise empty. `snoozed_until` and actor fields SHALL NOT be exposed.

#### Scenario: No perpetual diff

- **GIVEN** an applied policy with unchanged configuration
- **WHEN** Terraform re-plans after refresh
- **THEN** the plan SHALL be empty

#### Scenario: Snoozing outside Terraform does not create drift

- **GIVEN** a managed policy snoozed in the Kibana UI
- **WHEN** Terraform re-plans after refresh
- **THEN** the plan SHALL be empty

### Requirement: Not-found and engine-disabled handling

On read, HTTP 404 SHALL remove the resource from state without error. On delete, HTTP 404 SHALL be treated as success. On any operation, HTTP 503 whose body carries error code `ALERTING_DISABLED` SHALL produce an error diagnostic naming the `alerting:v2:enabled` Kibana advanced setting.

#### Scenario: Externally deleted policy is dropped from state

- **GIVEN** a managed policy deleted in Kibana outside Terraform
- **WHEN** refresh runs and `GET` returns 404
- **THEN** the provider SHALL remove the resource from state without an error diagnostic
- **AND** the next plan SHALL propose creating it

#### Scenario: Delete tolerates an already-absent policy

- **GIVEN** a managed policy deleted outside Terraform
- **WHEN** destroy runs and `DELETE` returns 404
- **THEN** the provider SHALL treat the operation as successful

#### Scenario: Engine disabled

- **GIVEN** a target Kibana space with `alerting:v2:enabled` set to false
- **WHEN** any lifecycle operation runs and the API answers 503 with code `ALERTING_DISABLED`
- **THEN** the provider SHALL return an error diagnostic naming `alerting:v2:enabled`

### Requirement: Error surfacing and connection handling

Transport failures, unexpected HTTP statuses, and successful responses with an empty body SHALL produce error diagnostics that include the server's message when one is present. The resource SHALL use the provider-level Kibana client by default and the scoped client from a resource-level `kibana_connection` block when present.

#### Scenario: Validation error from the server

- **GIVEN** a request the server rejects with HTTP 400 and a JSON error body
- **WHEN** the operation completes
- **THEN** the diagnostic SHALL include the server's `message`

#### Scenario: Scoped connection is honoured

- **GIVEN** a resource with a `kibana_connection` block
- **WHEN** any lifecycle operation runs
- **THEN** the request SHALL be issued through the client derived from that block

### Requirement: Acceptance tests

The acceptance test suite SHALL be skipped below Elastic Stack 9.6.0, SHALL follow the repository's acceptance-test isolation conventions, and SHALL cover:

1. A policy with `description`, a workflow destination, `matcher` (tags and expression), `grouping = per_field`, and an interval `throttle`: create, update, import, and an empty re-plan.
2. A minimal catch-all policy with only `name` and `destinations`.
3. `enabled = false` on create and toggling it on update.
4. Removing `matcher`, `grouping`, and `throttle` and asserting they are cleared.
5. A policy in a non-default space, with composite-id import.
6. `action_policy_id` omitted (generated) and supplied.
7. Plan-only `ExpectError` steps for the cross-field rules and schema bounds, runnable without a live stack.

Tests that need a workflow destination SHALL create the workflow in the same configuration through `elasticstack_kibana_agentbuilder_workflow` so the suite is self-contained.

#### Scenario: Empty plan after every apply step

- **GIVEN** any acceptance configuration applied successfully
- **WHEN** a plan runs with no configuration change
- **THEN** the plan SHALL be empty

#### Scenario: Validation cases run without a stack

- **GIVEN** the plan-only validation test cases
- **WHEN** they run without `TF_ACC` or a live stack
- **THEN** they SHALL still execute and assert the expected diagnostics
