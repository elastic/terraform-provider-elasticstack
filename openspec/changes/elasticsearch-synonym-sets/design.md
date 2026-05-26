## Context

Elasticsearch Synonym Sets are named, cluster-scoped collections of synonym rules. Each rule is a string in Solr format (e.g. `"car, auto, automobile"` for equivalent synonyms; `"bike => bicycle"` for one-directional). They are managed via three set-level REST APIs:

- `PUT /_synonyms/{id}` — create or replace an entire synonym set (idempotent)
- `GET /_synonyms/{id}` — retrieve the set, paginated via `from`/`size`
- `DELETE /_synonyms/{id}` — remove the set (blocked with HTTP 400 if any index analyzer still references it)

The `go-elasticsearch` v8 typed client at the pinned version (`v8.19.6`) exposes these under `typedapi/synonyms/`:

```go
// Create/Update
client.Synonyms.PutSynonym(id).Request(&putsynonym.Request{
    SynonymsSet: []types.SynonymRule{
        {Id: &ruleID, Synonyms: "car, auto, automobile"},
    },
}).Do(ctx)

// Read (response: {Count int64, SynonymsSet []types.SynonymRuleRead})
client.Synonyms.GetSynonym(id).From(0).Size(500).Do(ctx)

// Delete
client.Synonyms.DeleteSynonym(id).Do(ctx)
```

Key type details:
- `types.SynonymRule.Id` is `*string` (optional on write — Elasticsearch auto-generates an ID when omitted)
- `types.SynonymRuleRead.Id` is `string` (always present in the read response)
- `types.SynonymRuleRead.Synonyms` is `string`

The pattern follows the `internal/elasticsearch/enrich/` implementation: `entitycore.NewElasticsearchResource` + `entitycore.NewElasticsearchDataSource`, with CRUD functions in `internal/clients/elasticsearch/synonyms.go`.

## Goals

- Terraform-manage Elasticsearch synonym sets (create, read, update, delete).
- Expose a companion data source for read-only lookup.
- Handle optional rule IDs gracefully (Optional+Computed; provider generates a stable UUID when omitted).
- Retrieve all rules from the paginated GET API without a configurable cap.
- Return a descriptive error when DELETE fails because the set is in use by an index analyzer.

## Non-Goals

- Per-rule resource (`elasticstack_elasticsearch_synonym_rule`) using the rule-level APIs — follow-up issue.
- List data source (`GET /_synonyms`) listing all synonym sets — follow-up issue.
- Managing synonym token filter configuration within index settings — handled by `elasticstack_elasticsearch_index`.

## Decisions

| Topic | Decision |
|-------|----------|
| Resource granularity | Set-level only. Single `elasticstack_elasticsearch_synonym_set` resource. |
| Rule ID policy | `Optional+Computed`. When the user omits `id`, the provider generates a stable UUID (e.g. `uuid.New().String()`) on first create and stores it in state. On subsequent applies the stored ID is used in the PUT request, keeping the set deterministic. |
| Pagination | Read loop: `from=0, size=500` per page, loop until `from+len(page) >= total_count`. No configurable cap. |
| Delete failure | Detect HTTP 400 response. Surface a diagnostic: `"Synonym set '<id>' cannot be deleted because it is referenced by one or more index analyzers. Remove the synonym set from all analyzer configurations first."` |
| Minimum ES version | Not enforced via `EnforceVersion` guard for now. Defer to implementation; add a guard if the version floor is confirmed. |
| `synonyms_set` ordering | The API returns rules in the order they were stored. Use a `list` (not `set`) in the schema so ordering is deterministic and round-trips without spurious diffs. On update, the full set is replaced atomically. |
| `synonym_set_id` mutability | `RequiresReplace` — the set identifier is its primary key in Elasticsearch; there is no rename operation. |
| Import | Resource implements `ResourceWithImportState`. Import ID is the full resource `id` in the format `<cluster_uuid>/<synonym_set_id>`. On import, parse the ID and call the Read path to restore state. |
| Data source | Uses `entitycore.NewElasticsearchDataSource`. Takes `synonym_set_id` as the lookup key; returns all attributes computed. |
| entitycore pattern | Follow `internal/elasticsearch/enrich/` exactly: `entitycore.ElasticsearchResource[SynonymSetData]`, schema factory, typed CRUD functions in `internal/clients/elasticsearch/synonyms.go`. |

## Risks / Trade-offs

- **Large synonym sets**: Sets may contain up to 10,000 rules. Each rule is stored verbatim in Terraform state. Accepted — the user is responsible for the size of their synonym set.
- **Rule ordering drift**: If the API reorders rules on each read, Terraform would show a perpetual diff. The GET API returns rules in insertion order (stable), so this should not occur. Acceptance test should verify round-trip ordering.
- **Auto-generated IDs**: When the user omits `id` from a rule, the provider generates a UUID. If the user later adds an explicit `id` to that rule in config, Terraform will treat it as a different rule (update). Users are encouraged to supply stable `id` values for long-lived rules. Document in resource docs.
- **Analyzer reference on delete**: Elasticsearch blocks deletion when the set is still referenced. The provider returns a clear diagnostic; the user must update their analyzer configuration outside Terraform or via the `elasticstack_elasticsearch_index` resource before retrying destroy.

## Open Questions

1. **Minimum ES version**: What is the minimum Elasticsearch version that supports the Synonyms API? (Likely 8.10 or 8.11 — confirm during implementation and add an `EnforceVersion` guard if needed.)
2. **Blob normalisation on read**: Does the GET API normalise synonym strings (e.g. strip trailing whitespace)? If so, confirm that round-trip produces identical strings to avoid spurious plan diffs. Verify during acceptance testing.
