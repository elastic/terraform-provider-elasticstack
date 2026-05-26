## Why

The Elasticsearch provider has no support for Synonym Sets today. Synonym sets are named, cluster-scoped collections of synonym rules (expressed in Solr format) that power search-time synonym expansion in analyzers. Teams using Terraform to manage their Elasticsearch index configuration cannot currently manage synonyms as code, forcing manual API calls or out-of-band scripts.

The Elasticsearch Synonyms API (`PUT`/`GET`/`DELETE /_synonyms/{id}`) has been stable since Elasticsearch 8.10 and is fully covered by the `go-elasticsearch` v8 typed client (`typedapi/synonyms/`).

## What Changes

- Add a new `elasticstack_elasticsearch_synonym_set` **resource** that manages a named synonym set in Elasticsearch.
- Add a companion `elasticstack_elasticsearch_synonym_set` **data source** for read-only lookup.

### Resource schema sketch

```hcl
resource "elasticstack_elasticsearch_synonym_set" "example" {
  synonym_set_id = "my-synonyms"           # required string; forces new

  synonyms_set = [                          # required list of synonym rule blocks
    {
      id       = "rule-car"                 # optional, computed string; provider generates UUID when omitted
      synonyms = "car, auto, automobile"    # required string; Solr synonym format
    },
    {
      id       = "rule-bike"
      synonyms = "bike => bicycle"
    },
  ]

  elasticsearch_connection { ... }         # optional; standard provider-level connection override
}
```

### Data source schema sketch

```hcl
data "elasticstack_elasticsearch_synonym_set" "example" {
  synonym_set_id = "my-synonyms"           # required string

  # All other fields are computed
  synonyms_set = [
    {
      id       = "rule-car"
      synonyms = "car, auto, automobile"
    },
  ]

  elasticsearch_connection { ... }
}
```

### Key design decisions

- **Single resource per set** (not per rule): The set-level `PUT`/`GET`/`DELETE` APIs map cleanly onto the Terraform CRUD lifecycle and any single-rule change triggers a full analyzer reload regardless, so per-rule granularity provides no operational benefit.
- **`id` within synonym rules is Optional+Computed**: Users may supply an explicit `id` (stable across plans) or omit it (provider generates a stable UUID, stored in state, reused on subsequent applies). This avoids plan-diff loops while not requiring users to supply IDs.
- **Read fetches all rules via pagination loop**: `GET /_synonyms/{id}` supports `from`/`size`; the resource loops until all rules are retrieved, so large sets are fully represented in state.
- **Delete failure surfaces a clear error**: When `terraform destroy` is called against a set that is still referenced by an index analyzer, the API returns HTTP 400. The provider returns a descriptive error diagnostic explaining the constraint.

## Capabilities

### New Capabilities

- `elasticsearch-synonym-sets`: New resource and data source for Elasticsearch synonym sets.

### Modified Capabilities

None.

## Impact

- **New package**: `internal/elasticsearch/synonyms/` — resource, data source, models, CRUD handlers.
- **New client functions**: `internal/clients/elasticsearch/synonyms.go` — `GetSynonymSet`, `PutSynonymSet`, `DeleteSynonymSet`.
- **Provider registration**: `provider/plugin_framework.go` — register resource and data source.
- **Descriptions**: `internal/elasticsearch/synonyms/descriptions/` — markdown description files.
- **Acceptance tests**: `internal/elasticsearch/synonyms/acc_test.go`.
- **Docs**: Generated via `make generate-docs`.
