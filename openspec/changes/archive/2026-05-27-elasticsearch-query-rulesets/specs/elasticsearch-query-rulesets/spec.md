# elasticsearch-query-rulesets Specification

Resource implementation: `internal/elasticsearch/queryrulesets`
Data source implementation: `internal/elasticsearch/queryrulesets`

## Purpose

Manage Elasticsearch [Query Rulesets](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-rules-apis.html) via Terraform. Query rulesets allow teams to declaratively pin or exclude search result documents based on contextual criteria, enabling consistent search-ranking behaviour as code.

## Schema

### Resource schema

```hcl
resource "elasticstack_elasticsearch_query_ruleset" "example" {
  ruleset_id = <required, string, PlanModifier: RequiresReplace>
  id         = <computed, string>  # "<cluster_uuid>/<ruleset_id>"

  rules = [
    {
      rule_id  = <required, string>
      type     = <required, string>  # "pinned" | "exclude"
      priority = <optional, int64>

      criteria = [
        {
          type     = <required, string>  # "always" | "exact" | "fuzzy" | "prefix" |
                                         # "suffix" | "contains" | "lt" | "lte" | "gt" | "gte"
          metadata = <optional, string>
          values   = <optional, string>  # JSON-encoded array; plan-time validation requires this when type != "always"
        }
      ]

      actions = {
        ids  = <optional, list(string)>
        docs = <optional, list(object({ _index = <required, string>, _id = <required, string> }))>  # mutually exclusive with ids
      }
    }
  ]

  elasticsearch_connection {  # optional
    endpoints    = <optional, list(string)>
    username     = <optional, string>
    password     = <optional, string>
    api_key      = <optional, string>
    bearer_token = <optional, string>
    es_client_authentication = <optional, string>
    insecure     = <optional, bool>
    ca_file      = <optional, string>
    ca_data      = <optional, string>
    cert_file    = <optional, string>
    cert_data    = <optional, string>
    key_file     = <optional, string>
    key_data     = <optional, string>
    headers      = <optional, map(string)>
  }
}
```

### Data source schema

```hcl
data "elasticstack_elasticsearch_query_ruleset" "example" {
  ruleset_id = <required, string>
  id         = <computed, string>  # "<cluster_uuid>/<ruleset_id>"
  rules      = <computed, list>    # same nested structure as resource

  elasticsearch_connection {  # optional
    endpoints    = <optional, list(string)>
    username     = <optional, string>
    password     = <optional, string>
    api_key      = <optional, string>
    bearer_token = <optional, string>
    es_client_authentication = <optional, string>
    insecure     = <optional, bool>
    ca_file      = <optional, string>
    ca_data      = <optional, string>
    cert_file    = <optional, string>
    cert_data    = <optional, string>
    key_file     = <optional, string>
    key_data     = <optional, string>
    headers      = <optional, map(string)>
  }
}
```

## ADDED Requirements

### Requirement: Query ruleset APIs (REQ-001)

The `elasticstack_elasticsearch_query_ruleset` resource SHALL use `PUT /_query_rules/{ruleset_id}` to create or atomically replace a ruleset (including all embedded rules), `GET /_query_rules/{ruleset_id}` to read it, and `DELETE /_query_rules/{ruleset_id}` to remove it.

The companion `data.elasticstack_elasticsearch_query_ruleset` data source SHALL use `GET /_query_rules/{ruleset_id}` to retrieve a ruleset read-only.

Non-success API responses (other than 404 on read) SHALL be surfaced as Terraform diagnostics.

Required cluster privilege: `manage_search_query_rules`.

#### Scenario: Resource create calls PUT

- GIVEN a valid `elasticstack_elasticsearch_query_ruleset` configuration
- WHEN create runs
- THEN the provider SHALL call `PUT /_query_rules/{ruleset_id}` with the full rules list

#### Scenario: API errors surfaced

- GIVEN a failing Elasticsearch API response (other than 404 on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error message

#### Scenario: Data source ruleset not found

- GIVEN a `ruleset_id` that does not exist in Elasticsearch
- WHEN the data source reads
- THEN diagnostics SHALL include an error indicating the ruleset was not found

### Requirement: Identity (REQ-002)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<ruleset_id>`. `id` SHALL be computed after a successful `PUT`. The data source SHALL expose the same computed `id` format.

#### Scenario: ID set after create

- GIVEN a successful `PUT /_query_rules/{ruleset_id}` call
- WHEN create completes
- THEN `id` in state SHALL be `<cluster_uuid>/<ruleset_id>`

#### Scenario: Data source ID set

- GIVEN a successful `GET /_query_rules/{ruleset_id}` call from the data source
- WHEN the data source reads
- THEN `id` in state SHALL be `<cluster_uuid>/<ruleset_id>`

### Requirement: `ruleset_id` requires replacement (REQ-003)

Changing `ruleset_id` SHALL require resource replacement, because Elasticsearch ruleset identifiers cannot be renamed.

#### Scenario: ruleset_id change triggers replace

- GIVEN `ruleset_id` changes in configuration
- WHEN Terraform plans
- THEN replacement SHALL be required

### Requirement: `rules` as an ordered list (REQ-004)

`rules` SHALL be a `ListNestedAttribute` (ordered, not a set). The resource SHALL preserve rule declaration order on create and update, and a subsequent plan against the same configuration SHALL show no changes.

Elasticsearch stores rules in declaration order and `GET /_query_rules/{ruleset_id}` returns them in that same order. On Read, the provider SHALL use the API response order directly when populating state after create, update, or refresh.

When prior Terraform state is available and its rule order differs from the API response order (for example after out-of-band API changes), the Read path SHALL reorder the API response to match the prior state order by `rule_id`. When no prior order is available (for example on import), the resource SHALL sort rules by `rule_id` to produce a stable state order that avoids perpetual plan diffs. The data source SHALL preserve the API response order because it has no prior state and users expect declaration order.

#### Scenario: Rule order preserved round-trip

- GIVEN a ruleset created with rules in a specific order
- WHEN the resource reads state after create
- THEN rules in state SHALL appear in the same order as configured
- AND a subsequent plan SHALL show no changes

#### Scenario: Data source preserves API rule order

- GIVEN a ruleset with rules in declaration order
- WHEN the data source reads
- THEN rules in state SHALL appear in the same order as the API response

### Requirement: Rule schema (REQ-005)

Each rule in the `rules` list SHALL have the following attributes:

- `rule_id` (Required, String): unique identifier within the ruleset.
- `type` (Required, String): must be `"pinned"` or `"exclude"`.
- `priority` (Optional, Int64): relative priority within the ruleset; omitted from the API request body when null.
- `criteria` (Required, List, min 1): list of match criteria; see REQ-006.
- `actions` (Required, Object): what to do when matched; see REQ-007.

#### Scenario: Rule with all fields

- GIVEN a rule with `rule_id`, `type = "pinned"`, `priority = 1`, one criterion, and `actions.ids`
- WHEN create runs
- THEN all fields SHALL be stored in state and round-trip without change on a subsequent plan

#### Scenario: Rule with optional priority omitted

- GIVEN a rule with no `priority` set
- WHEN create runs
- THEN the API request SHALL not include a `priority` key for that rule
- AND state SHALL reflect `priority` as null

### Requirement: Criteria schema (REQ-006)

Each entry in `criteria` SHALL have:

- `type` (Required, String): one of `"always"`, `"exact"`, `"fuzzy"`, `"prefix"`, `"suffix"`, `"contains"`, `"lt"`, `"lte"`, `"gt"`, `"gte"`.
- `metadata` (Optional, String): the query context field to match against (e.g. `"query"`); omitted from the API request body when null.
- `values` (Optional, String): a JSON-encoded array of string or numeric values (e.g. `jsonencode(["laptop", 42])`); omitted from the API request body when null. Required when `criteria.type != "always"`.

The provider SHALL validate at plan time that, when `values` is set, it is a syntactically valid JSON array string. The provider SHALL validate that `values` is omitted (or null) when `criteria.type == "always"`.

#### Scenario: Criteria with string values

- GIVEN `criteria.type = "exact"`, `metadata = "query"`, `values = jsonencode(["laptop", "notebook"])`
- WHEN create runs
- THEN the API request SHALL include `criteria.values = ["laptop", "notebook"]` (decoded from JSON)
- AND `values` in state SHALL be the JSON-encoded string

#### Scenario: Criteria with numeric values

- GIVEN `criteria.type = "gt"`, `metadata = "popularity"`, `values = jsonencode([100])`
- WHEN create runs
- THEN the API request SHALL include `criteria.values = [100]` (as a numeric, not string)

#### Scenario: Always criterion

- GIVEN `criteria.type = "always"` with no `metadata` or `values`
- WHEN Terraform validates the configuration
- THEN the configuration SHALL be accepted

#### Scenario: Invalid JSON in values

- GIVEN `values = "not-json"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic indicating that `values` must be a valid JSON array

### Requirement: Actions mutual exclusion (REQ-007)

Each rule's `actions` block SHALL contain exactly one of:

- `ids` (Optional, List of String): list of document IDs to pin or exclude (max 100 combined across all matching rules for pinned queries).
- `docs` (Optional, List of Object): list of `{_index (Required, String), _id (Required, String)}` pairs.

Setting both `ids` and `docs` simultaneously SHALL be invalid. Setting neither SHALL be invalid. The provider SHALL enforce this constraint at plan/validate time.

#### Scenario: Only ids set

- GIVEN `actions.ids = ["doc-1", "doc-2"]` and `actions.docs` absent
- WHEN Terraform validates
- THEN the configuration SHALL be accepted

#### Scenario: Only docs set

- GIVEN `actions.docs = [{_index = "my-index", _id = "42"}]` and `actions.ids` absent
- WHEN Terraform validates
- THEN the configuration SHALL be accepted

#### Scenario: Both ids and docs set

- GIVEN both `actions.ids` and `actions.docs` are set
- WHEN Terraform validates
- THEN the provider SHALL return a validation diagnostic describing the mutual exclusion constraint

#### Scenario: Neither ids nor docs set

- GIVEN `actions` block present but both `ids` and `docs` absent
- WHEN Terraform validates
- THEN the provider SHALL return a validation diagnostic

### Requirement: Read — not found removes resource from state (REQ-008)

When `GET /_query_rules/{ruleset_id}` returns a 404 response, the resource SHALL remove itself from state without returning an error diagnostic, allowing Terraform to plan a re-create on the next apply.

#### Scenario: Ruleset deleted outside Terraform

- GIVEN the ruleset was deleted in Elasticsearch outside of Terraform
- WHEN the resource read runs (e.g. during refresh)
- THEN the resource SHALL be removed from state

### Requirement: Update replaces entire ruleset (REQ-009)

Any change to `rules` (adding, removing, or modifying a rule) SHALL trigger an in-place update via `PUT /_query_rules/{ruleset_id}`, replacing all rules atomically. No replacement of the Terraform resource is required for rule-level changes.

#### Scenario: Rule modification triggers PUT without resource replace

- GIVEN a query ruleset resource exists in state
- WHEN a rule's `type` or `criteria` value changes in configuration
- THEN Terraform SHALL plan an in-place update (not a replacement)
- AND the PUT request SHALL contain the full updated rule list

#### Scenario: Rule addition triggers PUT

- GIVEN a ruleset resource exists with two rules in state
- WHEN a third rule is added to `rules` in configuration
- THEN Terraform SHALL plan an update and the PUT request SHALL contain all three rules

### Requirement: Connection (REQ-010)

By default the resource and data source SHALL use the provider-level Elasticsearch client. When `elasticsearch_connection` is configured, the resource or data source SHALL construct and use a resource-scoped Elasticsearch client for all API calls.

#### Scenario: Resource-level connection override

- GIVEN `elasticsearch_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource-scoped client SHALL be used instead of the provider client

#### Scenario: Data source connection override

- GIVEN `elasticsearch_connection` is configured on the data source
- WHEN the data source reads
- THEN the resource-scoped client SHALL be used instead of the provider client

### Requirement: Import state (REQ-011)

The resource SHALL implement `resource.ResourceWithImportState`. The import ID SHALL be the full resource `id` in the format `<cluster_uuid>/<ruleset_id>`. On import, the resource SHALL set `id` to the provided import ID and call the Read path to populate all other attributes.

#### Scenario: Import sets state from API

- GIVEN a valid import ID of the form `<cluster_uuid>/<ruleset_id>`
- WHEN `terraform import` runs
- THEN `id` in state SHALL equal the provided import ID
- AND all `rules` SHALL be populated from the API by the subsequent Read call

#### Scenario: Import followed by plan shows no diff

- GIVEN a resource was successfully imported
- WHEN `terraform plan` runs against a matching configuration
- THEN no attribute differences SHALL be shown and no replacement SHALL be planned

### Requirement: Minimum ES version (REQ-012)

The resource and data source SHALL enforce a minimum Elasticsearch version guard of **8.12.0** (Query Rules API GA). If the cluster reports a version below **8.12.0**, the provider SHALL return a clear diagnostic explaining the minimum version requirement rather than a raw API error.

#### Scenario: Unsupported cluster version

- GIVEN a cluster reporting a version below the minimum Query Rules API version
- WHEN the resource or data source runs any operation
- THEN the provider SHALL return a diagnostic citing the minimum required version

### Requirement: Acceptance test coverage (REQ-013)

The acceptance test suite SHALL cover:

1. Basic CRUD: create a ruleset with at least one `pinned` and one `exclude` rule; assert state matches; update rules (add a rule, modify criteria); assert state; destroy.
2. Rule ordering: verify round-trip preserves declaration order; confirm a subsequent plan shows no diff.
3. Numeric `criteria.values`: create a rule with `criteria.type = "gt"` and a numeric value; assert state round-trips the JSON string.
4. `actions.docs` variant: create a rule using `docs` instead of `ids`; assert state.
5. `criteria.type = "always"`: create a rule with an `always` criterion; assert accepted.
6. Import: create resource, import by composite ID, verify state, run plan, confirm no diff.
7. Data source: create resource, read via data source, verify all attributes match.
8. Not-found handling: delete ruleset outside Terraform; refresh; assert resource removed from state.

All acceptance tests SHALL be gated with a `SkipFunc` if the minimum Elasticsearch version for the Query Rules API cannot be met by the test environment.

#### Scenario: Acceptance test import step

- GIVEN an existing ruleset created in a prior acceptance test step
- WHEN `ImportState: true, ImportStateVerify: true` runs with the composite ID
- THEN all attributes in the imported state SHALL match the originally configured attributes

#### Scenario: Acceptance test data source

- GIVEN an existing ruleset managed by the resource
- WHEN the data source reads by `ruleset_id`
- THEN all `rules` returned SHALL match those in the resource state
