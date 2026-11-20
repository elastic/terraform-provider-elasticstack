## ADDED Requirements

### Requirement: Data source schema and identity

The provider SHALL expose a read-only `elasticstack_fleet_server_host` data source, implemented in `internal/fleet/serverhostds`, with the following schema:

```hcl
data "elasticstack_fleet_server_host" "example" {
  id       = <computed, string>          # mirrors host_id
  host_id  = <optional+computed, string> # selector or output
  name     = <optional+computed, string> # selector or output
  default  = <optional+computed, bool>   # selector (true only) or output
  space_id = <optional+computed, string> # space the request is sent through

  hosts    = <computed, list(string)>
  proxy_id = <computed, string>
  ssl = <computed, object({
    certificate             = string
    certificate_authorities = list(string)
    client_auth             = string
  })>
}
```

The data source SHALL NOT expose `ssl.key`, the `agent_*` and `es_*` SSL fields, or `space_ids`. After a successful read, `id` and `host_id` SHALL both equal the identifier of the matched Fleet server host.

#### Scenario: id populated on lookup by host_id

- GIVEN `host_id = "abc123"` is configured and identifies an existing Fleet server host
- WHEN the data source is read
- THEN `id` and `host_id` SHALL equal `"abc123"`

#### Scenario: id populated on lookup by name

- GIVEN `name = "Fleet"` is configured and exactly one Fleet server host has that name
- WHEN the data source is read
- THEN `id` and `host_id` SHALL equal that host's identifier

#### Scenario: Private key is not part of the schema

- GIVEN a Fleet server host whose `ssl` has a `certificate` and a `key`
- WHEN the data source is read
- THEN `ssl.certificate` SHALL be populated
- AND the data source state SHALL NOT contain an `ssl.key` attribute

### Requirement: Exactly one selector

The data source SHALL require exactly one of `host_id`, `name` or `default` to be set in configuration and SHALL report a violation as one plan-time diagnostic that names all three attributes. All three selectors SHALL be `Optional` and `Computed`: the ones not used as the selector SHALL be populated from the API response, and a value populated by a previous read SHALL NOT cause the selector validation to fail on a later plan.

#### Scenario: No selector configured

- GIVEN none of `host_id`, `name` or `default` is set in configuration
- WHEN a plan is generated
- THEN Terraform SHALL report a plan-time validation error naming `host_id`, `name` and `default`

#### Scenario: Two selectors configured

- GIVEN two of `host_id`, `name` and `default` are set in configuration
- WHEN a plan is generated
- THEN Terraform SHALL report a plan-time validation error

#### Scenario: Selector switch across applies

- GIVEN a prior apply looked up a Fleet server host by `host_id`, populating computed `name` and `default` in state
- WHEN a later configuration sets `name` instead of `host_id` for the same data source
- THEN the plan SHALL NOT fail validation

### Requirement: Selector value validation

`host_id` and `name` SHALL reject an empty string at plan time. `default` SHALL only accept `true` as a configured value: `default = false` SHALL be rejected at plan time.

#### Scenario: default set to false

- GIVEN `default = false` is the only selector set in configuration
- WHEN a plan is generated
- THEN Terraform SHALL report a plan-time validation error

#### Scenario: Empty name

- GIVEN `name = ""` is the only selector set in configuration
- WHEN a plan is generated
- THEN Terraform SHALL report a plan-time validation error

### Requirement: Lookup by host_id

When `host_id` is configured, the data source SHALL call the Fleet Get Fleet Server host API (`GET /api/fleet/fleet_server_hosts/{itemId}`) with that identifier. When the API returns 404, the data source SHALL fail with a "not found" error that names the configured `host_id`.

#### Scenario: Successful lookup by host_id

- GIVEN `host_id = "abc123"` identifies an existing Fleet server host
- WHEN the data source is read
- THEN the Get Fleet Server host API SHALL be called with `"abc123"`
- AND `name`, `default` and every other computed attribute SHALL be populated from the response

#### Scenario: Lookup by host_id not found

- GIVEN `host_id = "does-not-exist"` does not identify a Fleet server host
- WHEN the data source is read
- THEN Terraform SHALL fail with a "not found" error naming `"does-not-exist"`

### Requirement: Lookup by name and by default

When `name` or `default = true` is configured, the data source SHALL call the Fleet Get Fleet Server hosts API (`GET /api/fleet/fleet_server_hosts`) once and select the match from the returned list: for `name`, the hosts whose name is exactly equal to the configured value; for `default = true`, the hosts whose `is_default` is `true`. The data source SHALL populate its attributes from the matching list item.

#### Scenario: Lookup by name

- GIVEN `name = "Secondary"` is configured and the list contains hosts named `"Primary"`, `"Secondary"` and `"Secondary EU"`
- WHEN the data source is read
- THEN the host named `"Secondary"` SHALL be the match
- AND `host_id` SHALL equal that host's identifier

#### Scenario: Lookup of the default host

- GIVEN `default = true` is configured and exactly one host has `is_default = true`
- WHEN the data source is read
- THEN `host_id` and `name` SHALL be those of the default host
- AND `default` SHALL be `true`

### Requirement: List lookups need the complete list

When the list response reports a `total` greater than the number of returned hosts, the data source SHALL fail with an error instead of matching against the incomplete list. When the Fleet API returns a non-success status, the data source SHALL surface the error to Terraform diagnostics.

#### Scenario: Incomplete list

- GIVEN the list response reports `total = 3` and returns two hosts
- WHEN the data source is read by `name` or by `default`
- THEN Terraform SHALL fail with an error stating that the list is incomplete

#### Scenario: API error on list

- GIVEN the Fleet API returns an error for the list request
- WHEN the data source is read by `name` or by `default`
- THEN diagnostics SHALL contain the API error

### Requirement: Zero and ambiguous matches

When a lookup by `name` or by `default = true` matches no Fleet server host, the data source SHALL fail with a "not found" error that names the configured `name`, or states that there is no default Fleet server host. When it matches more than one Fleet server host, the data source SHALL fail with an error that lists the `host_id` of every matching host. The data source SHALL NOT pick one of several matches.

#### Scenario: Name matches no host

- GIVEN `name = "does-not-exist"` is configured and no host has that name
- WHEN the data source is read
- THEN Terraform SHALL fail with an error containing `no Fleet server host found matching name "does-not-exist"`

#### Scenario: Name matches two hosts

- GIVEN `name = "Shared"` is configured and hosts `"a"` and `"c"` both have that name
- WHEN the data source is read
- THEN Terraform SHALL fail with an error containing `host_id "a"` and `host_id "c"`

#### Scenario: No default host

- GIVEN `default = true` is configured and no host has `is_default = true`
- WHEN the data source is read
- THEN Terraform SHALL fail with an error stating that no default Fleet server host was found

### Requirement: Attribute mapping

After a successful lookup by any selector, the data source SHALL map the API response as follows:

- `name` from `name`, and `hosts` from `host_urls` in API order.
- `default` from `is_default`; an absent `is_default` SHALL map to `false`.
- `proxy_id` from `proxy_id`; an absent or empty value SHALL map to null.
- `ssl.certificate`, `ssl.certificate_authorities` and `ssl.client_auth` from the matching `ssl` fields. `ssl.certificate_authorities` SHALL be null, not an empty list, when the API returns none. `ssl` SHALL be null when the API returns no `ssl` object or none of these three fields.

#### Scenario: default reported on lookup by host_id

- GIVEN `host_id` identifies the default Fleet server host
- WHEN the data source is read
- THEN `default` SHALL be `true`

#### Scenario: Host without proxy or ssl

- GIVEN a Fleet server host with no proxy and no SSL settings
- WHEN the data source is read
- THEN `proxy_id` SHALL be null
- AND `ssl` SHALL be null

#### Scenario: ssl with certificate only

- GIVEN a Fleet server host whose `ssl` has only `certificate` and `key`
- WHEN the data source is read
- THEN `ssl.certificate` SHALL be populated
- AND `ssl.certificate_authorities` and `ssl.client_auth` SHALL be null

#### Scenario: Data source follows an update of the host

- GIVEN a prior read of a Fleet server host
- WHEN the host's `host_urls` are changed and the data source is read again
- THEN `hosts` SHALL report the new URLs

### Requirement: space_id routes the request and does not filter

When `space_id` is configured, the data source SHALL send its Fleet API requests through that Kibana space (`/s/{space_id}/api/fleet/...`); otherwise it SHALL use the default space. Fleet server hosts are shared by all spaces, so the data source SHALL NOT treat `space_id` as a filter, and SHALL NOT expose `space_ids`.

#### Scenario: Host created in another space

- GIVEN a Fleet server host created by the resource with `space_ids = ["team-a"]`
- WHEN the data source looks it up without `space_id`, with `space_id = "team-a"` and with `space_id = "default"`
- THEN every lookup SHALL return that host

### Requirement: No version gate on the data source

The data source SHALL NOT enforce a minimum Elastic Stack version. On stack versions whose Fleet API does not return `proxy_id` or `ssl`, those attributes SHALL be null.

#### Scenario: Stack without ssl support

- GIVEN a stack whose Fleet API does not return `ssl` for Fleet server hosts
- WHEN the data source is read
- THEN the read SHALL succeed
- AND `ssl` SHALL be null
