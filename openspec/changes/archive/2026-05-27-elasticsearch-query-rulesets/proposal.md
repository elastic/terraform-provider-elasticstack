## Why

Teams using Elasticsearch's [Query Rules API](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-rules-apis.html) cannot manage query rulesets via Terraform today ([#1064](https://github.com/elastic/terraform-provider-elasticstack/issues/1064)). Query rules allow teams to declaratively pin or exclude documents from search results based on match criteria (query text, metadata, user properties, etc.). Without provider support, teams must use manual API calls or the generic HTTP provider, which makes IaC workflows painful, hard to diff, and difficult to keep in sync.

The Elasticsearch Query Rules API reached GA in 8.12. The `go-elasticsearch` v8 client (v8.19.6, already in use) ships a fully generated `typedapi/queryrules/` namespace covering all required operations. No client changes are needed.

## What Changes

- Add a new **resource** `elasticstack_elasticsearch_query_ruleset` that manages a full query ruleset and all its embedded rules via `PUT /_query_rules/{ruleset_id}`.
- Add a companion **data source** `data.elasticstack_elasticsearch_query_ruleset` that retrieves a ruleset read-only via `GET /_query_rules/{ruleset_id}`.
- **Out of scope for this proposal artifact**: editing `openspec/specs/` directly; that happens when the change is synced or archived.

### Resource shape

```hcl
resource "elasticstack_elasticsearch_query_ruleset" "example" {
  ruleset_id = "my-search-rules"   # required, forces new

  rules = [
    {
      rule_id  = "pin-featured"
      type     = "pinned"          # "pinned" | "exclude"
      priority = 1                 # optional int

      criteria = [
        {
          type     = "exact"       # see criteria types below
          metadata = "query"       # optional string
          values   = jsonencode(["laptop", "notebook"])  # JSON-encoded array
        }
      ]

      actions = {
        ids  = ["doc-1", "doc-2"]  # mutually exclusive with docs
        # docs = [{_index = "idx", _id = "x"}]
      }
    }
  ]

  elasticsearch_connection { ... }
}

data "elasticstack_elasticsearch_query_ruleset" "example" {
  ruleset_id = "my-search-rules"

  elasticsearch_connection { ... }
}
```

### Key design decisions

- **Ruleset-as-single-resource**: `PUT /_query_rules/{ruleset_id}` is the atomic management surface; the entire ruleset (all rules) is written in one API call per apply. This matches how the synonym_set resource works and avoids cross-resource dependency management.
- **`criteria.values` encoding**: stored as a JSON-encoded string attribute (e.g. `jsonencode(["laptop"])`) to support both string and numeric criteria values, which is required because `criteria.values` is `[]json.RawMessage` in the Go client.
- **`rules` as a list (ordered)**: rule ordering within a ruleset is semantically significant for pinned results; a `ListNestedAttribute` preserves declaration order and avoids perpetual plan diffs.
- **`actions` mutual exclusion**: exactly one of `ids` or `docs` may be set per rule; enforced at plan/validate time.
- **Data source included** per issue owner direction.

### Version

The Query Rules API reached GA in Elasticsearch 8.12. The provider SHALL enforce a minimum Elasticsearch version guard of **8.12.0** (see REQ-012).

### Acceptance tests

- Basic CRUD: create a ruleset with `pinned` and `exclude` rules, assert state, update rules, assert state, destroy.
- Rule ordering: verify round-trip preserves declaration order; confirm subsequent plan shows no diff.
- `criteria.values` with numeric values: create rule with numeric criteria values; assert state round-trips correctly.
- `actions.docs` variant: create rule using `docs` instead of `ids`; assert state.
- Import: create resource, import by composite ID, verify state, run plan, confirm no diff.
- Data source: create resource, read via data source, verify all attributes match.

## Capabilities

### New Capabilities

- `elasticsearch-query-rulesets`: resource `elasticstack_elasticsearch_query_ruleset` and data source `data.elasticstack_elasticsearch_query_ruleset`.

### Modified Capabilities

- _(none)_

## Impact

- **Specs**: Delta under `openspec/changes/elasticsearch-query-rulesets/specs/elasticsearch-query-rulesets/spec.md` until merged into canonical spec.
- **Implementation** (future): new files under `internal/elasticsearch/queryrulesets/` (schema, model, CRUD); new client wrapper in `internal/clients/elasticsearch/`; registration in `provider/plugin_framework.go`; docs/descriptions; acceptance tests.
