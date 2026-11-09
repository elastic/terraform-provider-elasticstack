## Purpose

Defines the `elasticstack_kibana_action_policy` resource, which manages a single Kibana Alerting v2 action policy (Kibana 9.6.0+, technical preview in the provider). Practitioners use it to declare which alerts are notified, how they are grouped and throttled, and which workflows receive them. The policy can then be reviewed and drift-detected as code instead of being configured by hand in the Kibana UI.

## ADDED Requirements

### Requirement: Action policy API usage

The resource SHALL manage an action policy through the Kibana Alerting v2 action policy API (`/api/alerting/v2/action_policies`): create, get by id, full replace by id, and delete by id. Every call SHALL target the Kibana space given by `space_id`. After a successful create or update, the resource SHALL read the policy back and derive final state from that read.

#### Scenario: Create writes the configured policy

- GIVEN a configuration with `name = "ops-pager"` and one workflow destination
- WHEN Terraform applies the configuration
- THEN an action policy named `ops-pager` with that destination exists in the `default` space
- AND Terraform state reflects the policy as read back from Kibana

#### Scenario: Policy in a named space

- GIVEN a configuration with `space_id = "observability"`
- WHEN Terraform applies the configuration
- THEN the action policy is created in the `observability` space
- BUT no policy is created in the `default` space

### Requirement: Composite identity and import

The resource SHALL expose a computed `id` of the form `<space_id>/<policy_id>`. Import SHALL accept an identifier in that same form, and the following read SHALL populate every configurable attribute from the policy stored in Kibana.

#### Scenario: Import an existing policy

- GIVEN an action policy with id `ops-pager` in the `observability` space
- WHEN the practitioner imports it with id `observability/ops-pager`
- THEN Terraform state has `id = "observability/ops-pager"`, `space_id = "observability"` and `policy_id = "ops-pager"`
- AND the name, description, enabled state, destinations, matcher, grouping and throttle in state match the policy in Kibana

#### Scenario: Import id without a space segment

- GIVEN an import id `ops-pager` that has no `<space_id>/` prefix
- WHEN the practitioner imports it
- THEN the import fails with an error describing the expected `<space_id>/<policy_id>` format

### Requirement: Optional client-chosen policy identifier

When `policy_id` is configured, the resource SHALL create the policy with exactly that identifier. When `policy_id` is omitted, Kibana SHALL generate the identifier and the resource SHALL record it in `policy_id`.

#### Scenario: Configured identifier is used

- GIVEN a configuration with `policy_id = "ops-pager"` and no policy with that id in the space
- WHEN Terraform applies the configuration
- THEN the created policy has id `ops-pager`
- AND state has `id = "default/ops-pager"`

#### Scenario: Server-generated identifier

- GIVEN a configuration without `policy_id`
- WHEN Terraform applies the configuration
- THEN state records the identifier Kibana assigned in `policy_id`
- AND `id` is `default/<assigned identifier>`

### Requirement: Existing policy is never taken over on create

When `policy_id` is configured and a policy with that identifier already exists in the target space when create runs, create SHALL fail with an error that tells the practitioner to import the policy. The existing policy SHALL NOT be modified.

#### Scenario: Identifier already in use

- GIVEN an action policy with id `ops-pager` already exists in the `default` space and is not in Terraform state
- WHEN Terraform applies a configuration with `policy_id = "ops-pager"`
- THEN the apply fails with an error that mentions importing `default/ops-pager`
- BUT the existing policy's name, destinations and other settings are unchanged

### Requirement: Replacement on identity changes

Changing `space_id` or `policy_id` SHALL force replacement of the resource. Changing any other configurable attribute SHALL update the policy in place.

#### Scenario: Moving a policy to another space

- GIVEN a managed policy in the `default` space
- WHEN the configuration changes `space_id` to `observability`
- THEN the plan shows the resource being replaced

#### Scenario: Renaming a policy

- GIVEN a managed policy named `ops-pager`
- WHEN the configuration changes `name` to `ops-oncall`
- THEN the plan shows an in-place update
- AND after apply the policy keeps its identifier and is named `ops-oncall`

### Requirement: Updates replace the full policy definition

An update SHALL write the complete configured definition. An optional setting removed from configuration (`description`, `matcher`, `grouping`, `throttle`) SHALL be cleared on the policy rather than retained from its previous value.

#### Scenario: Removing the matcher makes the policy catch-all

- GIVEN a managed policy with `matcher.tags = ["payments"]`
- WHEN the `matcher` block is removed from configuration and Terraform applies
- THEN the policy in Kibana has no matcher and applies to all alerts in its space
- AND state has `matcher` null

#### Scenario: Removing the throttle

- GIVEN a managed policy with `throttle.strategy = "time_interval"` and `throttle.interval = "1h"`
- WHEN the `throttle` block is removed from configuration and Terraform applies
- THEN the policy in Kibana has no throttle configured
- AND the next plan is empty

### Requirement: Enabled state management

When `enabled` is configured, the resource SHALL make the policy's enabled state match it after create and after update. When `enabled` is not configured, the resource SHALL record the server-reported enabled state and SHALL NOT change it. New policies are enabled by Kibana.

#### Scenario: Create a disabled policy

- GIVEN a configuration with `enabled = false`
- WHEN Terraform applies the configuration
- THEN the policy exists in Kibana and is disabled
- AND state has `enabled = false`

#### Scenario: Re-enable a policy

- GIVEN a managed policy with `enabled = false`
- WHEN the configuration changes `enabled` to `true` and Terraform applies
- THEN the policy is enabled in Kibana

#### Scenario: Unset enabled follows the server

- GIVEN a managed policy whose configuration does not set `enabled`
- AND someone disables the policy in the Kibana UI
- WHEN Terraform refreshes and plans
- THEN state has `enabled = false`
- BUT the plan does not re-enable the policy

### Requirement: Partial failure while applying enabled state

If the policy is written successfully but changing its enabled state fails, the resource SHALL keep the policy in Terraform state and SHALL report the failure as an error. After a failed create, Terraform's standard handling applies: the policy is marked tainted and replaced on the next apply. After a failed update, the next plan SHALL show the enabled change still to apply.

#### Scenario: Disable call fails after create

- GIVEN a configuration with `enabled = false`
- AND Kibana rejects the request to disable the policy after creating it
- WHEN Terraform applies the configuration
- THEN the apply fails with an error describing the disable failure
- AND the created policy is recorded in state as tainted so it is not orphaned
- AND the next plan replaces the policy

#### Scenario: Disable call fails after update

- GIVEN a managed, enabled policy
- AND the configuration changes `name` and sets `enabled = false`
- AND Kibana accepts the policy update but rejects the request to disable it
- WHEN Terraform applies the configuration
- THEN the apply fails with an error describing the disable failure
- AND state records the updated name with `enabled = true`
- AND the next plan shows `enabled` changing to `false`

### Requirement: Snoozing is not managed

The resource SHALL NOT expose or change a policy's snooze state. Snoozing or unsnoozing a policy outside Terraform SHALL NOT produce a difference in the plan.

#### Scenario: Policy snoozed in the UI

- GIVEN a managed policy with no configuration changes
- AND someone snoozes the policy in the Kibana UI
- WHEN Terraform refreshes and plans
- THEN the plan is empty
- AND the policy stays snoozed

### Requirement: Workflow destinations

`destinations` SHALL be an unordered set with at least one entry. Each entry SHALL have a `type` and an `id`. `workflow` SHALL be the only accepted `type`, and `id` SHALL identify the workflow that receives notifications. Reordering destinations in configuration SHALL NOT produce a plan difference.

#### Scenario: Reordered destinations

- GIVEN a managed policy with workflow destinations `wf-a` and `wf-b`
- WHEN the configuration lists the same destinations in the order `wf-b`, `wf-a`
- THEN the plan is empty

#### Scenario: Unsupported destination type

- GIVEN a destination with `type = "connector"`
- WHEN Terraform validates the configuration
- THEN validation fails with an error listing `workflow` as the accepted type
- AND no Kibana API call is made

#### Scenario: No destinations

- GIVEN a configuration with an empty `destinations` set
- WHEN Terraform validates the configuration
- THEN validation fails with an error requiring at least one destination

### Requirement: Alert matching and rule routing

`matcher` SHALL select the alerts the policy applies to. `matcher.tags` SHALL match alerts from any rule whose routing tags contain at least one of the tags. `matcher.expression` SHALL be a KQL condition evaluated against each alert. When both are set, an alert SHALL match both. When `matcher` is omitted, the policy SHALL apply to every alert in its space.

#### Scenario: Tag-based routing

- GIVEN a policy with `matcher.tags = ["payments"]`
- WHEN Terraform applies the configuration
- THEN the stored policy matches alerts from rules whose routing tags include `payments`
- AND state has `matcher.tags = ["payments"]` and `matcher.expression` null

#### Scenario: Tags and expression combined

- GIVEN a policy with `matcher.tags = ["payments"]` and `matcher.expression = "severity : \"critical\""`
- WHEN Terraform applies and refreshes
- THEN both values are stored on the policy and recorded in state unchanged

#### Scenario: Empty matcher rejected

- GIVEN a configuration with a `matcher` block that sets neither `tags` nor `expression`
- WHEN Terraform validates the configuration
- THEN validation fails with an error requiring `tags` or `expression`, or removing `matcher` to match all alerts

#### Scenario: Empty tag set rejected

- GIVEN a configuration with `matcher.tags = []`
- WHEN Terraform validates the configuration
- THEN validation fails with an error requiring at least one tag

### Requirement: Grouping and throttle settings

`grouping` and `throttle` SHALL be stored as configured. When either is omitted from configuration, the policy SHALL be stored without it and state SHALL record it as null. Kibana's dispatch-time defaults SHALL NOT be written into state.

#### Scenario: Per-field grouping with time interval throttle

- GIVEN `grouping = { mode = "per_field", fields = ["host.name"] }` and `throttle = { strategy = "time_interval", interval = "15m" }`
- WHEN Terraform applies and refreshes
- THEN the policy in Kibana groups by `host.name` and throttles every `15m`
- AND the next plan is empty

#### Scenario: Grouping and throttle omitted

- GIVEN a configuration without `grouping` or `throttle`
- WHEN Terraform applies and refreshes
- THEN state has `grouping` and `throttle` null
- AND the next plan is empty

### Requirement: Plan-time shape validation

The resource SHALL reject the following at validation time, before any Kibana API call:

- unknown `grouping.mode` or `throttle.strategy` values
- `grouping.fields` that is missing for `per_field` or present for any other mode
- `throttle.interval` that is missing for `per_status_interval`/`time_interval`, present for other strategies, or not a duration such as `5m`
- a `policy_id` outside 1–150 characters of `[a-zA-Z0-9_-]`

#### Scenario: Fields without per_field mode

- GIVEN `grouping = { mode = "all", fields = ["host.name"] }`
- WHEN Terraform validates the configuration
- THEN validation fails with an error stating that `fields` is only allowed when `mode` is `per_field`

#### Scenario: Missing interval

- GIVEN `throttle = { strategy = "per_status_interval" }`
- WHEN Terraform validates the configuration
- THEN validation fails with an error requiring `interval`

#### Scenario: Malformed interval

- GIVEN `throttle = { strategy = "time_interval", interval = "5 minutes" }`
- WHEN Terraform validates the configuration
- THEN validation fails with an error describing the expected duration format

#### Scenario: Invalid policy identifier

- GIVEN `policy_id = "ops pager!"`
- WHEN Terraform validates the configuration
- THEN validation fails with an error describing the allowed characters

### Requirement: Server-side constraints are left to Kibana

Kibana SHALL be the authority for the rules that plan-time validation does not check:

- compatibility between grouping mode and throttle strategy
- size limits
- KQL syntax

A Kibana rejection of these rules SHALL fail the apply with an error that contains Kibana's message.

#### Scenario: Incompatible grouping and throttle

- GIVEN `grouping = { mode = "all" }` and `throttle = { strategy = "on_status_change" }`
- WHEN Terraform plans and applies the configuration
- THEN the plan succeeds without a validation error
- BUT the apply fails with an error that includes Kibana's message about the invalid strategy for the grouping mode

### Requirement: Surrounding whitespace is rejected

`name`, `description`, `matcher.expression` and each `grouping.fields` entry SHALL be rejected at validation time if they have leading or trailing whitespace. These values, when set, SHALL be non-empty.

#### Scenario: Name with trailing space

- GIVEN `name = "ops-pager "`
- WHEN Terraform validates the configuration
- THEN validation fails with an error stating that `name` must not have leading or trailing whitespace
- AND no Kibana API call is made

#### Scenario: Empty description

- GIVEN `description = ""`
- WHEN Terraform validates the configuration
- THEN validation fails with an error requiring a non-empty description or omitting it

### Requirement: Missing policy on read

When Kibana reports that the policy does not exist during refresh, the resource SHALL be removed from Terraform state without an error.

#### Scenario: Policy deleted outside Terraform

- GIVEN a managed policy that someone deleted in the Kibana UI
- WHEN Terraform refreshes
- THEN the resource is removed from state
- AND the next plan proposes creating it again

### Requirement: Delete

Destroy SHALL delete the policy. If Kibana reports the policy as already missing, destroy SHALL succeed.

#### Scenario: Destroy a managed policy

- GIVEN a managed policy
- WHEN Terraform destroys the resource
- THEN the policy no longer exists in Kibana

#### Scenario: Destroy an already-deleted policy

- GIVEN a managed policy that someone already deleted in Kibana
- WHEN Terraform destroys the resource
- THEN destroy succeeds without an error

### Requirement: API error surfacing

The resource SHALL report each of the following as an error diagnostic that includes the HTTP status and Kibana's error message:

- request rejections (400)
- missing privileges or an insufficient license (403)
- concurrent modification conflicts (409)
- Alerting v2 being disabled (503)
- unexpected statuses or transport failures

#### Scenario: License does not allow action policies

- GIVEN a Kibana 9.6.0+ deployment without an Enterprise or trial license
- WHEN Terraform applies a new policy
- THEN the apply fails with an error that includes the 403 status and Kibana's license message

#### Scenario: Alerting v2 disabled

- GIVEN a Kibana deployment where Alerting v2 is turned off
- WHEN Terraform applies a new policy
- THEN the apply fails with an error that includes the 503 status and Kibana's message

### Requirement: Minimum Kibana version

Before create, read and update, the resource SHALL check that the target Kibana is version 9.6.0 or newer. On older versions, the operation SHALL fail with an `Unsupported server version` error, and the resource SHALL make no action policy API call.

#### Scenario: Kibana older than 9.6.0

- GIVEN the provider targets Kibana 9.5.3
- WHEN Terraform applies a new action policy
- THEN the apply fails with an `Unsupported server version` error naming 9.6.0 as the minimum
- BUT no action policy is created

### Requirement: Kibana connection

The resource SHALL use the provider's Kibana connection by default. It SHALL also honor a resource-level `kibana_connection` block that overrides the provider connection for all of this resource's API calls.

#### Scenario: Resource-level connection override

- GIVEN a `kibana_connection` block pointing at a second Kibana instance
- WHEN Terraform applies the configuration
- THEN the policy is created on the Kibana instance from the `kibana_connection` block
- BUT not on the provider's default Kibana instance

### Requirement: Experimental registration

The resource SHALL be available only when the provider's experimental resources are enabled through `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`. Without opting in, the resource type SHALL be unknown to the provider.

#### Scenario: Experimental resources not enabled

- GIVEN `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` is not set
- WHEN a configuration declares an `elasticstack_kibana_action_policy` resource
- THEN Terraform reports that the resource type is not supported by the provider

#### Scenario: Experimental resources enabled

- GIVEN `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`
- WHEN a configuration declares an `elasticstack_kibana_action_policy` resource
- THEN Terraform plans the resource normally

### Requirement: Documentation of availability

The resource documentation SHALL state:

- the resource is in technical preview and may change or be removed
- it requires Kibana 9.6.0 or newer, an Enterprise license and Alerting v2 enabled
- it is unavailable in serverless Elasticsearch and Security projects
- rules reach a policy through their routing tags

#### Scenario: Availability notice

- GIVEN the generated documentation page for `elasticstack_kibana_action_policy`
- WHEN a practitioner reads the resource description
- THEN it contains the technical preview notice, the version, license and serverless availability prerequisites, and the routing-tags explanation

### Requirement: Documentation of notification credentials

The resource documentation SHALL state:

- notifications run with the credentials of whoever last wrote the policy, and every Terraform update re-binds them to the provider's Kibana credentials
- when the provider authenticates with an Elasticsearch API key, that key itself is stored on the policy, so revoking it breaks dispatch
- Kibana rejects Cloud API key authentication for these policies

#### Scenario: Credentials notice

- GIVEN the generated documentation page for `elasticstack_kibana_action_policy`
- WHEN a practitioner reads the resource description
- THEN it explains the last-writer credentials binding, the Elasticsearch API key caveat, and the Cloud API key rejection
