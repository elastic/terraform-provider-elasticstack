## Context

This design follows the exploration and decisions recorded by the issue reporter in
[issue #94](https://github.com/elastic/terraform-provider-elasticstack/issues/94#issuecomment-6182116998)
(local experiments against Elasticsearch 9.4.0) and the detailed resource contract posted immediately
after in
[this follow-up comment](https://github.com/elastic/terraform-provider-elasticstack/issues/94#issuecomment-6182117013).
Both comments were authored by the issue's own triaging maintainer (`@tobio`, the actor who triggered
this proposal) and are the most specific, actionable statement of scope in the thread; this design
treats that contract as the spine of the proposal rather than re-deriving alternatives it already
evaluated (prefix-expression ownership vs. dedicated rollover-alias ownership; see "Decisions" below).

Two related problems motivate the new resource, scoped against the existing
`elasticstack_elasticsearch_index` resource and the existing `elasticstack_elasticsearch_alias`
resource:

1. An ordinary index resource correctly tries to restore its concrete index and configured alias
   properties. That conflicts with ILM moving the writer and eventually deleting the bootstrap index.
2. Managing only the bootstrap index provides no series-wide cleanup on destroy. Remaining generations
   can still reference the ILM policy and block its deletion.

```text
Terraform: bootstrap/adoption + series existence + final cleanup
ILM:       rollover + generation changes + retention

orders-000001 --> orders-000002 --> orders-000003
      |
      +--> eventually deleted; Terraform must not restore it
```

Existing building blocks this design relies on:

- `internal/clients/elasticsearch/alias.go`'s `GetAlias` already calls
  `Indices.GetAlias().Name(aliasName)` and returns `map[string]types.IndexAliases` keyed by target
  name. Because data-stream aliases can have the same response shape, shared discovery must also
  resolve target types and reject data streams before treating those keys as concrete indices.
  `DeleteIndexAlias` and `UpdateIndexAlias` exist but are not reused here (see Decisions).
- `internal/clients/elasticsearch/index.go`'s `DeleteIndex` and `PutIndex` cover concrete-index
  create/delete.
- `internal/entitycore` provides the Elasticsearch resource envelope
  (`entitycore.NewElasticsearchResource[Model]`), the `<cluster_uuid>/<resource_id>` composite-ID
  convention (`internal/entitycore/composite_id_importer.go`), and the shared operation-timeouts
  attribute — all used elsewhere (e.g. `internal/elasticsearch/ccr/followerindex`).
- `elasticstack_elasticsearch_index`'s `deletion_protection` (`internal/elasticsearch/index/index/schema.go`,
  `delete.go`) is the precedent for this resource's own `deletion_protection` attribute and error.

## Goals / Non-Goals

**Goals:**
- Let Terraform bootstrap (or adopt) a rollover series once, then leave ILM free to roll over, rename
  (shrink), and delete generations without the provider trying to undo those changes.
- Provide a single resource whose destroy cleans up every index currently attached to the series,
  unblocking ILM policy deletion.
- Make series existence and identity independent of the volatile writer/generation, so plans stay
  stable across rollovers.
- Default to safe deletion (`deletion_protection = true`), consistent with `elasticstack_elasticsearch_index`.

**Non-Goals:**
- No change to `elasticstack_elasticsearch_index` or `elasticstack_elasticsearch_alias` semantics.
- No settings/mappings management on the new resource — index/component templates remain the only
  source for those on every generation.
- No automatic writer repair, scheduled rollover, or ILM-policy ownership.
- No prefix-based orphan cleanup, no searchable-snapshot/repository cleanup.
- No guarantee of complete historical cleanup once the ownership alias has been removed or a member
  has been detached before destroy runs.

## Decisions

### Ownership model: dedicated alias, not prefix expression

Two options were evaluated (see the linked exploration comment for the full comparison table):

- **A: prefix-expression ownership** — membership resolved by `<prefix>-*`, including matching
  historical aliases.
- **B: dedicated rollover-alias ownership** — membership is every concrete index currently attached
  to `rollover_alias`.

**Selected: B.** Attaching an index to the dedicated alias puts it inside this resource's deletion
boundary; detaching it takes it out. This follows actual alias membership across ILM renaming
(shrink) and makes the alias — not the name prefix — the explicit series boundary. Consequences
documented in the capability spec: an index could join the namespace by name but still fall outside
ownership (not attached to the alias), and an unrelated index could be brought into ownership simply
by being attached to the alias. The whole alias must be reserved for this series and must not be
shared with unrelated data.

Two related ideas were considered and rejected:
- **Survivor heuristics** (prefix/last-known-member probing when the alias is missing) — these can
  find unrelated indices and miss detached renamed ones; they are not proof of alias-based ownership,
  so there is no survivor branch anywhere in this design.
- **Erroring on an unhealthy-writer Read** — this would block unrelated applies and normal destroy
  even though the owned alias and indices still exist. Existence and health are kept separate: Read
  warns, it does not error, for health problems (see "Read health vs. existence" below).

### Schema

```hcl
resource "elasticstack_elasticsearch_rollover_index_series" "orders" {
  prefix         = "orders"
  rollover_alias = "orders-write"

  deletion_protection = false # defaults to true; must be applied before destroy/replacement

  depends_on = [
    elasticstack_elasticsearch_index_template.orders,
  ]
}
```

| Attribute | Type / mode | Meaning |
|---|---|---|
| `prefix` | Required string; `RequiresReplace` | Literal prefix used to create `<prefix>-000001` only. Not a deletion selector. |
| `rollover_alias` | Required string; `RequiresReplace` | Dedicated ordinary-index alias; permanent logical identity and the sole membership boundary. |
| `deletion_protection` | Optional bool, default `true` | Blocks Delete while enabled; state-only attribute, mirrors `elasticstack_elasticsearch_index`. |
| `id` | Computed string | `<cluster_uuid>/<rollover_alias>`. Stable across writer changes, bootstrap deletion, shrink, and ordinary refresh. |
| `write_index` | Computed string, nullable | Concrete effective writer name, or null when no effective writer exists. If the writer exists but its name lacks the `-<digits>` suffix required for automatic rollover, its actual name is still exposed (with a Read warning), never hidden. |
| `indices` | Computed set of strings | Current concrete members of the dedicated alias (observed, not a persisted ownership ledger) — includes ILM-renamed and hidden/closed members. |
| `timeouts` | Standard entitycore operation-timeouts attribute | create/read/update/delete, following the shared envelope convention rather than bespoke parsing. |

Both `prefix` and `rollover_alias` are required explicitly; there is no default derived from the
other. `prefix` is bootstrap-naming-only — always `<prefix>-000001`, with no custom initial name,
generation override, or date-math support. Naming inputs are validated as literal index/alias names
(no wildcard, comma-separated, remote-cluster, or date-math expressions); the generated
`<prefix>-000001` name is validated in full (including total length) against Elasticsearch index-name
rules, and `rollover_alias` is validated with alias-appropriate rules rather than assuming identical
constraints to index names.

No `settings`, `settings_raw`, shard/replica attributes, mappings, extra aliases, alias
filter/routing configuration, lifecycle-policy selection, rollover conditions, configurable initial
index name, date math, or initial-generation override are exposed. Index/component templates remain
the only source of those for every generation; the index template should carry
`index.lifecycle.name` and `index.lifecycle.rollover_alias`, and must NOT also define the rollover
alias in its own `aliases` block (Elastic documents that combination as a cause of duplicate-alias
rollover errors) — the initial write alias is defined exactly once, by this resource's bootstrap PUT.

### Effective-writer resolution

Resolved using real Elasticsearch alias semantics, never by sorting names or picking a generation:

1. A member explicitly marked `is_write_index=true` is the writer.
2. A sole member with an **unset** write flag is an implicit writer.
3. A sole member with explicit `false` is not a writer.
4. Multiple members without a designated writer: no effective writer.

For Create/adoption, the effective writer's name must additionally end in `-<digits>` (the
auto-generated rollover-name requirement) — its prefix need not match `prefix`. A writer named e.g.
`orders-current` fails adoption validation because automatic rollover cannot derive a successor name
from it, even though the alias is otherwise writable.

"Healthy" means exactly "an ordinary index alias with one effective, numerically-suffixed writer" —
not a claim about shard health, ILM execution, or template/policy correctness.

### Create and adoption

1. Look up the exact alias (`GetAlias`) and resolve its target types, distinguishing ordinary index
   aliases from data-stream aliases/targets. Reject any data-stream target before adoption, and
   distinguish genuine absence from authorization, transport, decoding, or other API errors (errors
   must never be interpreted as absence).
2. **Alias missing:** `PUT /<prefix>-000001` with only
   `{"aliases": {"<rollover_alias>": {"is_write_index": true}}}` — no settings/mappings in the body.
   If `<prefix>-000001` already exists (as an index, or as an incompatible object such as an alias),
   fail with the collision error; do not silently adopt, attach the alias to it, delete it, or pick
   another generation. Do not scan `<prefix>-*` or old member names to decide whether bootstrapping
   is allowed — a detached `orders-000002` does not make a missing alias "exist", and a new bootstrap
   may legitimately coexist with such detached indices.
3. **Alias exists and is healthy:** adopt automatically (no opt-in adoption flag) — read back
   membership and the effective writer, but make no mutation (no index creation, no alias
   add/rewrite, no write-flag normalization, no settings mutation). Adopt all current members
   regardless of whether their names match `prefix`. Automatic adoption acquires the right to delete
   all current alias members once `deletion_protection` is disabled; this must be documented clearly
   (see capability spec and resource docs).
4. **Alias exists but is unhealthy** (no effective writer, an incompatible writer-name suffix, or an
   unsupported alias type such as a data-stream alias or data-stream target): fail Create with an
   actionable diagnostic and make no mutation. Do not fix the writer or attach unrelated indices to
   make adoption succeed.
5. After any mutation, re-read the live alias to derive final state (membership + writer), consistent
   with other entitycore resources deriving state from a post-mutation read. Treat an interrupted
   create as a signal to re-discover rather than to blindly retry — a subsequent Create (e.g. a
   retried apply) can adopt a healthy alias left behind by an earlier successful-but-interrupted
   bootstrap.

### Read: health vs. existence

| Observation | Terraform result |
|---|---|
| Alias exists with a valid writer | Stable identity retained; `indices` and `write_index` refreshed. |
| Bootstrap index gone, alias remains on newer members | Normal refresh; `000001` is never recreated or restored. |
| Alias exists, no effective writer | **Warning only**; resource retained, membership refreshed, `write_index` set to null. |
| Alias exists, writer name lacks `-<digits>` | **Warning only**; resource retained, membership refreshed, actual writer name exposed. |
| Exact alias absent | Resource removed from state; next apply attempts normal Create/bootstrap. There is no "possible survivors" branch — the bootstrap namespace and last-known members are not queried when the alias disappears. |
| Auth/connection/timeout/decoding/unexpected API failure | Error — never interpreted as absence, state is never discarded. |
| Unsupported data-stream alias/type | Error — never treated as an ordinary series and never used to justify deleting backing indices. |

Warnings explain why writes/automatic rollover may fail and that the resource will not repair the
configuration; they must not force replacement and must not block an otherwise valid
protection/timeout update or a series deletion. Read never restores `is_write_index=true` on the
bootstrap index, never moves the alias to a guessed latest generation, and never re-adds removed
members.

### Update and replacement

There is no in-place Elasticsearch mutation path for this resource's own attributes:

- `deletion_protection` and the operation-timeouts attribute update Terraform state only, followed by
  a readback of membership/writer. The required readback GET is allowed; no mutating API call is tied
  to these attribute changes.
- `prefix` and `rollover_alias` changes both require replacement (`RequiresReplace`), even on an
  adopted series where `prefix` was never used to create anything.
- Changes to the computed `indices`/`write_index` outputs reflect external lifecycle progress, not
  drift to repair.
- No placeholder API call (e.g. `_validate/query`) is sent to simulate an update when nothing in
  config changed.

For a `prefix`/`rollover_alias` replacement, the user must ensure destroy-before-create ordering.
Terraform Core controls replacement ordering; `RequiresReplace` cannot enforce this, and
`create_before_destroy` can be configured directly or propagated through dependencies. If a new
instance adopts the same alias before the old instance is destroyed, the old instance can delete the
new instance's members. Document this lifecycle precondition and explicitly warn against
`create_before_destroy` for this resource. The protection guard applies to replacement too: a user
must apply `deletion_protection = false` as a prior, separate apply — toggling it only in the same
operation that replaces the resource does not establish an already-applied unprotected prior state.

### Delete

1. Check `deletion_protection`; if enabled, return an actionable error and make no mutation.
2. Discover the exact alias's current membership (`GetAlias`) and resolve target types, explicitly
   including hidden and closed indices. If any target is a data stream, fail before deleting any
   target; only ordinary concrete indices are in scope.
3. If the alias is genuinely missing, succeed without deleting anything else — detached indices are
   explicitly out of scope for cleanup, and this must not be described as having "discovered
   survivors" (nothing was probed).
4. Delete the freshly discovered concrete index names (`DeleteIndex`), never the alias name and never
   a wildcard, respecting `action.destructive_requires_name=true`.
5. Tolerate members ILM deletes concurrently between discovery and deletion (404-as-success, same
   convention as `DeleteWithNotFoundAsSuccess`); do not suppress permission or other non-absence
   failures.
6. Re-discover membership once more; if a rollover introduced new members, repeat within the delete
   timeout rather than relying only on the state-stored `indices` set.
7. If cleanup fails or cannot converge before timeout, error and retain state for retry — partial
   deletion must be safely recoverable by re-discovery on the next Delete attempt, and success must
   never be reported while observed members remain.

An unhealthy writer never blocks Delete. Deletion protection guards provider-initiated deletion calls
only; it does not and cannot prevent an external actor from deleting the alias — if a subsequent Read
observes the alias absent, normal state removal still applies even though the prior resource was
protected.

Never: change `action.destructive_requires_name`, stop ILM globally, delete by shared-policy
membership, fall back to a prefix sweep, or remove the alias before capturing the membership needed
for deletion.

### Identity and import

Canonical state ID: `<cluster_uuid>/<rollover_alias>` — stable across writer changes, bootstrap
deletion, shrink, and refresh; never the writer, an index UUID, or a generation number.

Because `prefix` cannot always be reconstructed from an adopted series (an adopted writer's name need
not match any `prefix` the importer could guess), import uses a three-part identifier rather than the
generic two-part `entitycore.CompositeIDImporter`:

```text
<cluster_uuid>/<rollover_alias>/<prefix>
```

```shell
terraform import \
  elasticstack_elasticsearch_rollover_index_series.orders \
  '<cluster_uuid>/orders-write/orders'
```

Implementation: use `entitycore.ParseCompositeID` to split off `<cluster_uuid>`, then split the
remaining `<rollover_alias>/<prefix>` resource-ID portion on the first `/` (consistent with the doc
comment on `ParseCompositeID` recommending direct use when a resource needs to derive more than a
single passthrough field). The importer sets `rollover_alias`, `prefix`, and
`deletion_protection = true` before Read hydrates the remaining attributes; alias readback cannot
initialize this Terraform-only attribute. Read must preserve that imported value as well as any
subsequently applied value. If the alias resolves to a data-stream target, import hydration fails
rather than adopting it. Disabling protection still requires a subsequent apply.

### Why not extend an existing resource

An earlier idea from the issue thread (`create_initial_index` on the ILM policy resource, or a
template-driven bootstrap field) was considered and rejected: it would make a lifecycle-policy or
template resource responsible for one-time index creation and for a destroy-time cleanup concern that
has nothing to do with policy/template CRUD, and still would not solve series-wide destroy. A
dedicated resource keeps `elasticstack_elasticsearch_index_lifecycle` and the template resources
unchanged and gives the series its own identity, schema, and destroy path.

### Implementation direction

- Terraform Plugin Framework, following `internal/entitycore`'s Elasticsearch resource-envelope
  pattern (`entitycore.NewElasticsearchResource[Model]`) and existing Elasticsearch resource layout
  (cf. `internal/elasticsearch/ccr/followerindex`).
- Reuse the existing `go-elasticsearch/v8` typed client already used by
  `internal/clients/elasticsearch/alias.go` (`GetAlias`) and `index.go` (`PutIndex`, `DeleteIndex`).
  No new client library.
- Do not reuse `internal/elasticsearch/index/aliasutil`'s membership-reconciliation logic — that
  package exists to make `elasticstack_elasticsearch_alias` restore configured membership, which is
  the opposite of this resource's observe-don't-restore contract. A new, small alias-membership /
  effective-writer helper is shared internally between this resource's own Create/Read/Delete only.
- Derive final state from a post-mutation read on every operation, consistent with repository
  convention elsewhere in `entitycore`-based resources.

## Open questions

- **Stack version matrix**: live exploration covered Elasticsearch 9.4.0 only, including an actual
  ILM shrink transition (via a targeted `_ilm/move` step bypass) and alias-transfer behavior.
  Searchable-snapshot alias transfer is documentation/source-backed but not live-tested. This proposal
  does not establish a new minimum supported Stack version; acceptance testing on the provider's full
  supported version matrix is needed before claiming broader support, and should be tracked explicitly
  in `tasks.md` rather than assumed.
- **Delete convergence timeout**: the exploration establishes that delete must re-poll briefly for a
  rollover race between discovery and deletion, but does not fix a specific timeout/retry budget.
  This needs to be pinned down against the shared entitycore operation-timeouts convention during
  implementation (e.g. whether the existing `delete` timeout default is sufficient or needs a
  resource-specific minimum).
- **CI acceptance environment**: triggering an actual ILM shrink in CI required a targeted
  `_ilm/move` step bypass locally (see the linked exploration comment) rather than natural ILM timing.
  Implementation should confirm whether the existing acceptance-test Elasticsearch fixture can safely
  perform the same targeted step move, or whether shrink/searchable-snapshot membership-transfer
  coverage needs a different mechanism (e.g. a unit test against a hand-built `GetAlias` response).

## References

- [Elastic: bootstrap alias-based time series using policies and templates](https://www.elastic.co/docs/manage-data/lifecycle/index-lifecycle-management/tutorial-time-series-without-data-streams)
- [Elastic: aliases and implicit write-index semantics](https://www.elastic.co/docs/manage-data/data-store/aliases)
- [Elastic: ILM rollover requirements](https://www.elastic.co/docs/reference/elasticsearch/index-lifecycle-actions/ilm-rollover)
- [Elastic: delete-index API restrictions](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-indices-delete)
- [Elastic: shrink naming and alias transfer](https://www.elastic.co/docs/reference/elasticsearch/index-lifecycle-actions/ilm-shrink)
- [Elastic: searchable-snapshot naming](https://www.elastic.co/docs/reference/elasticsearch/index-lifecycle-actions/ilm-searchable-snapshot)
- [Elastic: manual index deletion does not delete the underlying searchable snapshot](https://www.elastic.co/docs/reference/elasticsearch/index-lifecycle-actions/ilm-delete)
- [Issue #94 exploration comment](https://github.com/elastic/terraform-provider-elasticstack/issues/94#issuecomment-6182116998)
- [Issue #94 proposed design comment](https://github.com/elastic/terraform-provider-elasticstack/issues/94#issuecomment-6182117013)
