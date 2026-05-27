## Context

The Elasticsearch [Query Rules API](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-rules-apis.html) allows operators to define rulesets that control which documents are pinned or excluded from search results based on contextual match criteria. The API reached GA in 8.12 and is documented at `PUT/_GET/_DELETE /_query_rules/{ruleset_id}`.

The `go-elasticsearch` v8 client (v8.19.6) already contains a fully generated `typedapi/queryrules/` namespace with all required operations: `putruleset`, `getruleset`, `deleteruleset`, `listrulesets`, `putrule`, `getrule`, `deleterule`, and `test`. No client changes are needed.

Relevant client types:
- `typedapi/types.QueryRule` — `rule_id`, `type`, `criteria []QueryRuleCriteria`, `actions QueryRuleActions`, `priority *int`
- `typedapi/types.QueryRuleCriteria` — `type`, `metadata *string`, `values []json.RawMessage`
- `typedapi/types.QueryRuleActions` — `ids []string`, `docs []QueryRuleActionDoc` (mutually exclusive)
- `typedapi/types.QueryRuleActionDoc` — `_index string`, `_id string`
- `typedapi/queryrules/putruleset.Request` — `{rules []QueryRule}`
- `typedapi/queryrules/getruleset.Response` — `{rules []QueryRule, ruleset_id string}`

## Goals

- Expose `elasticstack_elasticsearch_query_ruleset` as a Terraform resource for full ruleset lifecycle management.
- Expose `data.elasticstack_elasticsearch_query_ruleset` as a read-only data source.
- Preserve rule ordering in state to avoid perpetual plan diffs.
- Support all criteria types and the `criteria.values` polymorphism (string and numeric values) through a JSON-encoded string attribute.

## Non-Goals

- Per-rule resource (`elasticstack_elasticsearch_query_rule`) — the ruleset-as-single-resource approach is sufficient and aligns with the atomic PUT semantics of the API.
- `POST /_query_rules/{ruleset_id}/_test` — evaluation endpoint; not a lifecycle operation.
- `GET /_query_rules` list endpoint as a data source.
- Modifications to `xpack.applications.rules.max_rules_per_ruleset` cluster setting.

## Decisions

| Topic | Decision |
|-------|-----------|
| Resource strategy | Single resource per ruleset (`elasticstack_elasticsearch_query_ruleset`); entire ruleset written atomically via `PUT /_query_rules/{ruleset_id}` |
| `criteria.values` encoding | `types.StringAttribute` holding a JSON-encoded array string (e.g. `jsonencode(["laptop", 42])`); rationale: the API accepts `[]json.RawMessage` which can contain strings or numbers — a typed list would lose numeric support |
| `rules` attribute type | `types.ListNestedAttribute` (ordered); ordering is semantically significant for pinned results |
| `actions` mutual exclusion | Exactly one of `ids` (list of string) or `docs` (list of `{_index, _id}`) must be set per rule; enforced via plan-time validator |
| `ruleset_id` mutability | Changing `ruleset_id` requires resource replacement; Elasticsearch ruleset IDs are immutable once created |
| `id` format | `<cluster_uuid>/<ruleset_id>` — consistent with the synonym-set and other Elasticsearch resource patterns |
| CRUD mapping | Create/update: `PUT /_query_rules/{ruleset_id}`; Read: `GET /_query_rules/{ruleset_id}`; Delete: `DELETE /_query_rules/{ruleset_id}` |
| Not-found on read | 404 removes resource from state (allows Terraform to plan a re-create) |
| Connection | Resource-level `elasticsearch_connection` block, falling back to provider-level client |
| Import | Import ID `<cluster_uuid>/<ruleset_id>`; `ImportState` reads API to populate all attributes |
| Data source | Companion data source reads via `GET`; same `ruleset_id` key; returns full rule list |
| `criteria.metadata` | Optional string; omitted in JSON body when null |
| `priority` | Optional int on `QueryRule`; omitted when null |

## Terraform schema sketch

### Resource

```
ruleset_id  (Required, String, PlanModifier: RequiresReplace)
id          (Computed, String)  — "<cluster_uuid>/<ruleset_id>"

rules (List, Required):
  rule_id   (Required, String)
  type      (Required, String)  — "pinned" | "exclude"
  priority  (Optional, Int64)
  criteria (List, Required, min 1):
    type      (Required, String)  — "always" | "exact" | "fuzzy" | "prefix" | "suffix" |
                                    "contains" | "lt" | "lte" | "gt" | "gte"
    metadata  (Optional, String)
    values    (Optional, String)  — JSON-encoded array; required unless criteria.type == "always"
  actions (Required, Object):
    ids  (Optional, List<String>)
    docs (Optional, List<Object>):
      _index (Required, String)
      _id    (Required, String)

elasticsearch_connection (Optional, SingleNestedBlock)
```

### Data source

```
ruleset_id  (Required, String)
id          (Computed, String)
rules       (Computed, List) — same nested structure as resource
elasticsearch_connection (Optional, SingleNestedBlock)
```

## Risks / Trade-offs

- **Single-rule change rewrites entire ruleset**: Any change to any rule triggers `PUT /_query_rules/{ruleset_id}` with the full list. This is acceptable because the PUT is atomic and the API is designed for this pattern; however, with rulesets near the 100-rule limit, Terraform configs may become verbose.
- **Rule ordering from API**: The research note flags uncertainty about whether `GET /_query_rules/{ruleset_id}` preserves declaration order. If Elasticsearch does not return rules in insertion order, the provider may need to sort by `rule_id` on read to produce a stable state. This is an implementation detail to verify during acceptance testing.
- **`criteria.values` as JSON string**: Practitioners must use `jsonencode([...])` rather than a native list. This is a trade-off for numeric criteria support. Validation that the string is valid JSON should be added.

## Open Questions

1. **Minimum ES version guard**: The Query Rules API was introduced as technical preview in 8.10 and reached GA in 8.12. Which version should the provider guard against? The issue owner indicated this is an implementation detail, but a concrete decision (8.10 vs 8.12) must be made during implementation. Recommendation: guard at 8.12 (GA) unless the product team explicitly wants to support the tech-preview path.

2. **Rule ordering guarantee**: Does `GET /_query_rules/{ruleset_id}` return rules in declaration/insertion order? If not, the read path must apply a stable sort (by `rule_id`) to avoid perpetual plan diffs. Verify during acceptance testing.

3. **`criteria.values` validation**: Should the provider validate that the `values` string is parseable JSON at plan time? Recommendation: yes, add a string validator that ensures valid JSON array.

## Migration / State

This is a new resource and data source; no migration is required. Existing manual API configurations can be imported using `terraform import elasticstack_elasticsearch_query_ruleset.<name> <cluster_uuid>/<ruleset_id>`.
