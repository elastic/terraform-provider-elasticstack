## Why

There is no way to look up an existing Fleet agent policy from Terraform. When one stack owns an
agent policy (managed with `elasticstack_fleet_agent_policy`) and another stack needs its
`policy_id` — for example to read enrollment tokens with `elasticstack_fleet_enrollment_tokens` and
enroll Elastic Agents through a Helm chart or a Kubernetes manifest — the consuming stack must either
hardcode the policy UUID or read the owning stack's state with `terraform_remote_state`, coupling the
two stacks' state and requiring the owning stack to expose and apply an output first.

The Fleet data sources available today don't cover this: `elasticstack_fleet_enrollment_tokens` can
filter by `policy_id` (the value we are trying to find), and `elasticstack_fleet_integration` /
`elasticstack_fleet_output` return packages and outputs, not agent policies. A resource,
`elasticstack_fleet_agent_policy`, exists; no matching data source does.

This change adds a read-only, singular `elasticstack_fleet_agent_policy` data source that resolves one
agent policy by `policy_id` or by `name`, closing issue #5324.

## What Changes

- Add a new `elasticstack_fleet_agent_policy` data source backed by the Fleet Get Agent Policy API
  (`GET /api/fleet/agent_policies/{agentPolicyId}`) and the Fleet List Agent Policies API
  (`GET /api/fleet/agent_policies`).
- The data source accepts exactly one of `policy_id` or `name` as the lookup selector (both
  `Optional+Computed`, so the unused one is populated from the API response as an output), and an
  optional `space_id`.
- `name` lookup uses a server-side `Kuery` pre-filter against the saved-object field
  (`ingest-agent-policies.name:"<escaped name>"`), followed by a client-side exact compare against
  each returned item's `Name` so an escaping mistake fails closed rather than silently widening the
  match. Zero matches after the filtered query is an error; more than one match is an error that
  names the conflicting `policy_id`s.
- Exported attributes mirror the read-only surface of the `elasticstack_fleet_agent_policy` resource
  (`namespace`, `description`, `data_output_id`, `monitoring_output_id`, `fleet_server_host_id`,
  `download_source_id`, `monitor_logs`, `monitor_metrics`, `is_protected`, `host_name_format`,
  `supports_agentless`, `inactivity_timeout`, `unenrollment_timeout`, `global_data_tags`,
  `space_ids`, `required_versions`, `advanced_settings`, `advanced_monitoring_options`), read
  directly from the API response rather than through the resource's drift-avoidance mapping, since a
  data source has no configured value to preserve nulls against.
- This is a singular, single-match data source only; a filterable list of agent policies is out of
  scope (see Out of Scope below).

## Capabilities

### New Capabilities

<!-- None — this extends the existing `fleet-agent-policy` capability. -->

### Modified Capabilities

- `fleet-agent-policy`: Adds the read-only `elasticstack_fleet_agent_policy` data source — schema,
  selector validation, by-ID and by-name lookup behavior, ambiguous/not-found error handling, and
  attribute mapping — to the capability that already covers the
  `elasticstack_fleet_agent_policy` resource.

## Impact

- **New code**: `internal/fleet/agentpolicyds/` (data source package: `schema.go`, `models.go`,
  `read.go`, `data_source.go`), plus a new client helper `GetAgentPolicies` (name/Kuery list lookup)
  alongside the existing `GetAgentPolicy` in `internal/clients/fleet/agent_policy.go`.
- **New docs/examples**: `docs/data-sources/fleet_agent_policy.md`,
  `examples/data-sources/elasticstack_fleet_agent_policy/data-source.tf`.
- **Provider registration**: register the new data source in `provider/plugin_framework.go` next to
  the other Fleet data sources (`enrollmenttokens.NewDataSource`, `outputds.NewDataSource`,
  `integrationds.NewDataSource`).
- **OpenSpec**: delta requirements added to `openspec/specs/fleet-agent-policy/spec.md` (the
  canonical spec for this entity); `openspec/specs/elasticstack-fleet-agent-policy/spec.md` is a
  leftover bugfix spec and is left untouched.
- **Backward compatibility**: additive only — no changes to the existing resource's schema, CRUD
  behavior, or its `populateFromAPI` mapping function.

## Out of Scope

- Any change to the existing `elasticstack_fleet_agent_policy` resource's schema, behavior, or its
  `populateFromAPI` mapping function.
- Supporting lookup by any key other than `policy_id` or `name` (e.g. by package/integration
  content).
- Returning multiple matching policies as a list. The data source errors on an ambiguous `name`
  match rather than returning a list; a separate, list-typed data source is a distinct future type
  if ever needed.
- Resolving `terraform_remote_state` cross-stack coupling in general; this only removes the need for
  it in the single-policy-lookup case this issue describes.
