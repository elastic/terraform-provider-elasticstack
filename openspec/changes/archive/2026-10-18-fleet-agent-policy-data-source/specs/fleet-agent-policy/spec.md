## ADDED Requirements

### Requirement: Data source schema and identity

The provider SHALL expose a read-only `elasticstack_fleet_agent_policy` data source with the
following schema:

```hcl
data "elasticstack_fleet_agent_policy" "example" {
  id        = <computed, string> # set explicitly by the read function from the API policy id
  policy_id = <optional+computed, string> # selector or output; exactly one of policy_id/name
  name      = <optional+computed, string> # selector or output; exactly one of policy_id/name
  space_id  = <optional, string>

  namespace             = <computed, string>
  description           = <computed, string>
  data_output_id        = <computed, string>
  monitoring_output_id  = <computed, string>
  fleet_server_host_id  = <computed, string>
  download_source_id    = <computed, string>
  monitor_logs          = <computed, bool>
  monitor_metrics       = <computed, bool>
  is_protected          = <computed, bool>
  host_name_format      = <computed, string>
  supports_agentless    = <computed, bool>
  inactivity_timeout    = <computed, duration string>
  unenrollment_timeout  = <computed, duration string>
  global_data_tags      = <computed, map(object)>  # same nested shape as the resource
  space_ids             = <computed, set(string)>
  required_versions     = <computed, map(int32)>
  advanced_settings            = <computed, object> # same nested shape as the resource
  advanced_monitoring_options  = <computed, object> # same nested shape as the resource
}
```

`skip_destroy` and `sys_monitoring` SHALL NOT be exposed; both are write-time-only resource
behaviors with no meaningful read-only value.

#### Scenario: id populated on by-ID lookup

- GIVEN `policy_id = "abc123"` is configured
- WHEN the data source is read
- THEN `id` SHALL equal `"abc123"`

#### Scenario: id populated on by-name lookup

- GIVEN `name = "k8s-metrics-pernode"` is configured and exactly one policy has that name
- WHEN the data source is read
- THEN `id` SHALL equal that policy's API-assigned `policy_id`

### Requirement: Exactly one of policy_id or name

The data source SHALL require exactly one of `policy_id` or `name` to be set in configuration.
Both attributes SHALL be `Optional` and `Computed`: whichever one was not used as the selector
SHALL be populated from the API response as an output, and a value populated by a previous read
SHALL NOT cause the selector validation to fail on a subsequent plan where only one of the two is
configured.

#### Scenario: Neither selector configured

- GIVEN neither `policy_id` nor `name` is set in configuration
- WHEN a plan is generated
- THEN Terraform SHALL report a plan-time validation error

#### Scenario: Both selectors configured

- GIVEN both `policy_id` and `name` are set in configuration
- WHEN a plan is generated
- THEN Terraform SHALL report a plan-time validation error

#### Scenario: Selector switch across applies does not trip validation

- GIVEN a prior apply looked up a policy by `policy_id`, populating computed `name` in state
- WHEN a subsequent configuration sets `name` instead of `policy_id` for the same data source
  instance
- THEN the plan SHALL NOT fail validation, because only the attribute set in configuration (`name`)
  is evaluated by `ExactlyOneOf`

### Requirement: Lookup by policy_id

When `policy_id` is configured, the data source SHALL call the Fleet Get Agent Policy API
(`GET /api/fleet/agent_policies/{agentPolicyId}`) with that ID and the configured `space_id`. When
the API returns 404, the data source SHALL fail with a "not found" error identifying the configured
`policy_id` and `space_id`.

#### Scenario: Successful by-ID lookup

- GIVEN `policy_id = "abc123"` identifies an existing agent policy
- WHEN the data source is read
- THEN the Get Agent Policy API SHALL be called with `"abc123"`
- AND `name` and every other computed attribute SHALL be populated from the response

#### Scenario: by-ID lookup not found

- GIVEN `policy_id = "does-not-exist"` does not identify an existing agent policy
- WHEN the data source is read
- THEN the Get Agent Policy API SHALL return 404
- AND Terraform SHALL fail with a "not found" error naming `"does-not-exist"`

### Requirement: Lookup by name

When `name` is configured, the data source SHALL call the Fleet List Agent Policies API
(`GET /api/fleet/agent_policies`) for the configured `space_id` with a server-side `Kuery` filter of
the form `ingest-agent-policies.name:"<escaped name>"`, where backslashes and double quotes in the
configured `name` are escaped (`\` → `\\`, `"` → `\"`) before being interpolated into the quoted
Kuery literal. The data source SHALL then apply an exact-match compare (`item.Name == name`) against
every item the filtered query returns. The data source SHALL NOT page the full, unfiltered
agent-policy list as its lookup strategy.

#### Scenario: Kuery filter sent on the wire

- GIVEN `name = "k8s-metrics-pernode"` is configured
- WHEN the data source is read
- THEN the List Agent Policies API SHALL be called with `Kuery =
  ingest-agent-policies.name:"k8s-metrics-pernode"`

#### Scenario: Name requiring escaping

- GIVEN `name = "has \"quotes\" and \\backslash"` is configured
- WHEN the data source is read
- THEN the `Kuery` value sent on the wire SHALL contain the backslash- and quote-escaped form of
  that name inside the quoted literal

#### Scenario: Exact-match compare applied to filtered results

- GIVEN the filtered query returns items whose names are similar to, but not identical to, the
  configured `name`
- WHEN the data source evaluates the results
- THEN only items with `Name` exactly equal to the configured `name` SHALL be considered matches

### Requirement: Exhaustive, anomaly-safe pagination of filtered results

The data source SHALL retrieve every page of the `Kuery`-filtered result set before evaluating the
exact-match count, and SHALL NOT report a unique match without having reached the server-reported
`Total` for the filtered query. When the filtered, paginated result set is not fully and
consistently retrieved — a page is shorter than its expected size before `Total` is reached, `Total`
changes between page requests, or the same `Id` is returned more than once — the data source SHALL
fail with an error rather than report a possibly-incomplete match.

#### Scenario: Pagination of filtered results is exhaustive before matching

- GIVEN the filtered query's result set spans more than one page
- WHEN the data source is read
- THEN the data source SHALL retrieve every page of the filtered result set before evaluating the
  exact-match count, and SHALL NOT report a unique match without having reached the server-reported
  `Total` for the filtered query

#### Scenario: Inconsistent pagination fails closed

- GIVEN a page of the filtered result set is returned shorter than expected before `Total` is
  reached, or `Total` changes between page requests, or the same `Id` appears on two different pages
- WHEN the data source is read
- THEN the data source SHALL fail with an error rather than report a match

### Requirement: Ambiguous and zero-match name lookup errors

When a `name` lookup matches zero policies after a complete, anomaly-free pagination of the filtered
result set, the data source SHALL fail with an error identifying the configured `name` and
`space_id`. When a `name` lookup matches more than one policy, the data source SHALL fail with an
error identifying the configured `name` and `space_id` and enumerating the conflicting `policy_id`
of every match.

#### Scenario: No match

- GIVEN `name = "does-not-exist"` matches zero agent policies in the configured space
- WHEN the data source is read
- THEN Terraform SHALL fail with an error naming `"does-not-exist"` and the configured `space_id`

#### Scenario: Ambiguous match

- GIVEN two agent policies in the same space both have the configured `name` (e.g. created in
  different spaces and then viewed from a space-unaware `name` lookup, or any other path that
  produces more than one exact match)
- WHEN the data source is read
- THEN Terraform SHALL fail with an error naming the configured `name` and listing the conflicting
  `policy_id` of each match

### Requirement: Attribute mapping reflects the raw API response

Every computed attribute other than `id` SHALL be mapped directly from the Fleet API response, with
no null-preservation against a prior configured value (the data source's configuration has no prior
value to preserve nulls against). `description` SHALL be reported exactly as returned by the API: a
`null` API value SHALL be stored as null, and an empty-string API value (Fleet's representation of
an omitted description) SHALL be stored as an empty string, not null. `monitor_logs` and
`monitor_metrics` SHALL be computed unconditionally from whether the API response's
`monitoring_enabled` array contains `"logs"` or `"metrics"` respectively, regardless of any prior
value.

#### Scenario: Omitted description reported as empty string

- GIVEN the underlying agent policy was created without a `description`
- WHEN the data source is read
- THEN `description` SHALL be the empty string, not null, matching the raw API response

#### Scenario: Monitoring flags reflect current API state

- GIVEN an agent policy whose monitoring was previously enabled for `logs` and is now disabled
- WHEN the data source is read
- THEN `monitor_logs` SHALL be `false`, reflecting the current API response, not any previously
  cached value

### Requirement: No version gate on the core lookup

The data source SHALL NOT impose a minimum stack version requirement for the by-ID or by-name
lookup itself, since both the Get Agent Policy and List Agent Policies APIs are unconditionally
available. The data source SHALL NOT impose per-attribute version gates on any of its `Computed`
attributes, since reading an older or absent API field never corresponds to a user-configured write
that the resource's per-attribute version gates are designed to protect.

#### Scenario: Lookup succeeds on the provider's minimum supported stack version

- GIVEN a stack version at or above the provider's minimum supported Elastic Stack version
- WHEN the data source is read with either `policy_id` or `name` configured
- THEN the lookup SHALL proceed without a version-gate error
