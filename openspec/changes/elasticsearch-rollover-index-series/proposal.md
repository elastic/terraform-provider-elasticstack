## Why

Practitioners who bootstrap an ILM-managed rollover series (`<prefix>-000001` plus a dedicated write
alias) with `elasticstack_elasticsearch_index` hit a structural conflict: ILM moves the write alias to
each new generation and eventually deletes retired generations, but the ordinary index resource
correctly tries to restore its one concrete index and its configured alias properties on every plan.
That makes "no changes" plans impossible once ILM has rolled over or deleted the bootstrap index
(tracked in this issue since 2022, with multiple workaround threads: `lifecycle.ignore_changes`,
`local-exec` provisioners, and a third-party REST-provider object). A second, related gap is that
nothing in the provider can clean up an entire generation series on `terraform destroy` — ILM policies
cannot be deleted while any generation still references them, and the bootstrap index alone does not
represent the series.

This change adds a dedicated resource that owns the **logical series** (bootstrap-or-adopt plus
series-wide destroy) while leaving ILM to own **generation changes** (rollover, renames, retention).
It does not change the existing index resource's one-concrete-index semantics.

## What Changes

### New Resource: `elasticstack_elasticsearch_rollover_index_series`

- Bootstraps `<prefix>-000001` with a dedicated `rollover_alias` definition only (no settings/mappings
  — those come from matching index/component templates), or adopts an already-healthy existing alias
  automatically (no opt-in flag required).
- Treats the dedicated alias as both the ownership boundary and the existence check: membership is
  "whatever concrete index is currently attached to `rollover_alias`", not a prefix scan and not a
  persisted membership ledger. This tracks ILM-renamed (shrink) and hidden/closed members and
  deliberately does not try to rediscover indices that have been detached from the alias or deleted
  outside Terraform.
- Resolves an "effective writer" using Elasticsearch's real alias write-index semantics (explicit
  `is_write_index=true`; sole member with unset flag; otherwise no writer) and never guesses a writer
  by sorting names or picking the newest generation.
- Exposes `deletion_protection` (default `true`, matching `elasticstack_elasticsearch_index`) that
  must be explicitly disabled via a separate apply before Delete/replacement can proceed.
- On Delete, discovers current alias membership (including hidden/closed/renamed members) and deletes
  those concrete index names — never the alias name, never a wildcard — tolerating members ILM deletes
  concurrently and re-polling briefly for rollover races within the delete timeout.
- Surfaces unhealthy state (no effective writer, or a writer whose name does not end in `-<digits>`)
  as a Read-time warning (state is retained, not removed) but as a hard Create/adoption error.

### New Internal Packages

- `internal/elasticsearch/index/rolloverindexseries/` — Plugin Framework resource (schema, models,
  create/read/update/delete, entitycore wiring, import), built on `entitycore.NewElasticsearchResource`.
- Reuses the existing `internal/clients/elasticsearch` alias/index helpers (`GetAlias`, `DeleteIndex`,
  `UpdateIndexAlias`) rather than the alias resource's membership-reconciliation logic in
  `internal/elasticsearch/index/aliasutil`, since this resource observes membership instead of
  restoring it.

## Capabilities

### New Capabilities

- **`elasticsearch-rollover-index-series`** — schema, bootstrap/adoption, read-health-warning,
  update/replacement, and series-wide delete behavior of the new
  `elasticstack_elasticsearch_rollover_index_series` resource.

## Impact

- `internal/elasticsearch/index/rolloverindexseries/` (new resource package)
- Provider registration: one new resource added to the Plugin Framework provider entrypoint
- Documentation: new `templates/resources/elasticsearch_rollover_index_series.md.tmpl` and generated
  `docs/resources/elasticsearch_rollover_index_series.md`
- No changes to `elasticstack_elasticsearch_index`, `elasticstack_elasticsearch_alias`,
  `elasticstack_elasticsearch_index_lifecycle`, or any index/component template resource
- Acceptance tests require a live Elasticsearch cluster capable of ILM rollover and (for one scenario)
  a real shrink step transition; no new minimum Stack version is established by this proposal alone

## Out of Scope

- Any change to `elasticstack_elasticsearch_index`'s drift/restore behavior
- Index `settings`/`mappings` management on the new resource (left to templates)
- Automatic writer repair, scheduled rollover, or ILM-policy ownership/management
- Prefix-based orphan cleanup of indices detached from the alias
- Searchable-snapshot/repository cleanup on delete
- A `create_initial_index`-style field on the ILM policy or index template resources (considered and
  rejected in favor of a dedicated resource; see design.md)
