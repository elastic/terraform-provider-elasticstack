## Why

Migrate the two most complex existing Plugin Framework index resources (`index_template` and `index_lifecycle`) to the entitycore `NewElasticsearchResource` envelope. Both currently embed `*entitycore.ResourceBase` directly and duplicate Read/Delete/Schema/client-resolution logic. Solving the "envelope + UpgradeState" coexistence problem for these two resources establishes the pattern for all remaining complex PF resources.

## What Changes

- `elasticstack_elasticsearch_index_template` (`internal/elasticsearch/index/template/`)
  - Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]`
  - Preserve `UpgradeState` (V0→V1), `ModifyPlan`, `ValidateConfig` on concrete type
  - Convert `Read`/`Delete` to envelope callbacks
  - `Create`/`Update` use config-derived alias reconciliation and version gating — these override envelope with placeholder callbacks
- `elasticstack_elasticsearch_index_lifecycle` (`internal/elasticsearch/index/ilm/`)
  - Same migration pattern, simpler: only `UpgradeState` to preserve
  - Read/Delete become callbacks; Create/Update fit envelope contract directly

## Capabilities

### New Capabilities
- `elasticsearch-index-template-via-envelope`
- `elasticsearch-index-lifecycle-via-envelope`

### Modified Capabilities
<!-- None. -->

## Impact

- Two packages in `internal/elasticsearch/index/`. No Terraform interface changes.
