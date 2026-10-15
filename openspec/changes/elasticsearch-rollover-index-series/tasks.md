## 1. Prep and discovery

- [ ] 1.1 Confirm the delete re-discovery/convergence timeout budget against the shared entitycore
  `timeouts` convention (see design.md "Open questions") and record the decision here before
  implementing `delete.go`.
- [ ] 1.2 Confirm whether the acceptance-test Elasticsearch fixture can safely perform the targeted
  `_ilm/move` step bypass used during exploration to trigger a real shrink, or whether
  shrink/searchable-snapshot membership-transfer coverage needs a unit test against a hand-built
  `GetAlias` response instead.
- [ ] 1.3 Confirm the minimum Elasticsearch version to document for this resource; live exploration
  covered 9.4.0 only, so do not claim support below the provider's existing documented minimum
  without separate verification.

## 2. Alias-membership and effective-writer helper

- [ ] 2.1 Create a small package-local helper (not `internal/elasticsearch/index/aliasutil`, which
  restores configured membership rather than observing it) that: given a `map[string]types.IndexAliases`
  from `clients/elasticsearch.GetAlias`, computes (a) the effective writer per the precedence rules in
  `specs/elasticsearch-rollover-index-series/spec.md` "Effective-writer resolution", and (b) the
  current concrete member set including hidden/closed members.
- [ ] 2.2 Implement the "healthy" classifier: exactly one effective writer AND that writer's name ends
  in `-<digits>`. Return a structured reason (no writer / multiple writers / non-numeric writer name)
  for use in Create-adoption error messages and Read warning diagnostics.
- [ ] 2.3 Add unit tests covering all `GetAlias` response shapes in the effective-writer requirement:
  explicit writer, sole implicit writer, sole explicit-false, multiple members with no writer,
  non-numeric writer name, hidden/closed members included in membership.

## 3. Naming validators

- [ ] 3.1 Add a `prefix` validator rejecting wildcard/comma/remote-cluster/date-math syntax and
  validating the full generated `"<prefix>-000001"` name (including total length) against
  Elasticsearch index-naming rules.
- [ ] 3.2 Add a `rollover_alias` validator using Elasticsearch alias-naming rules (not index-naming
  rules) and rejecting wildcard/comma/remote-cluster/date-math syntax.
- [ ] 3.3 Unit test both validators against the length/charset edge cases named in the delta spec.

## 4. `elasticstack_elasticsearch_rollover_index_series` resource

- [ ] 4.1 Create directory `internal/elasticsearch/index/rolloverindexseries/`.
- [ ] 4.2 Create `models.go` with a `Model` struct embedding `entitycore.ElasticsearchConnectionField`
  and the shared timeouts field, plus: `Prefix`, `RolloverAlias` (string, required); `DeletionProtection`
  (bool, default `true`); computed `ID`, `WriteIndex` (nullable string), `Indices` (set of string).
- [ ] 4.3 Create `schema.go`: `prefix`/`rollover_alias` as `RequiresReplace` strings using the
  validators from task 3; `deletion_protection` with `booldefault.StaticBool(true)`; `write_index` and
  `indices` as `Computed`; wire the standard entitycore timeouts attribute.
- [ ] 4.4 Create `create.go`:
  - Call `clients/elasticsearch.GetAlias` for `rollover_alias`.
  - If absent: call `PutIndex`/equivalent to `PUT /<prefix>-000001` with only the alias-definition
    body (`{"aliases": {"<rollover_alias>": {"is_write_index": true}}}`); on a name-collision error,
    return it verbatim without any fallback mutation.
  - If present and healthy (per task 2.2): perform no mutating call.
  - If present and unhealthy: return an actionable diagnostic naming the specific reason; no mutation.
  - After any Create path, re-read the alias and populate `id`, `write_index`, `indices` from the
    fresh read (never from the request body).
- [ ] 4.5 Create `read.go`:
  - Call `GetAlias`; if absent (and not merely an API error), call `resp.State.RemoveResource` and
    return — do not query `<prefix>-*` or any previously observed members.
  - If present, populate `indices` and `write_index`.
  - If unhealthy, append a warning diagnostic (do not call `RemoveResource`, do not error) explaining
    the specific reason and that the provider will not repair it.
  - Treat non-404 API errors as errors, not as absence.
- [ ] 4.6 Create `update.go`:
  - Persist `deletion_protection` and timeouts changes to state.
  - Re-read membership/writer for the computed attributes.
  - No Elasticsearch mutation call tied to these attribute changes; no placeholder API call.
- [ ] 4.7 Create `delete.go`:
  - Check prior-state `deletion_protection`; if `true`, return the actionable error from
    `specs/.../spec.md` "Deletion protection guards provider-initiated deletes" and make no mutation.
  - Discover current concrete membership via `GetAlias` (including hidden/closed).
  - If alias absent, succeed with no deletions.
  - Delete discovered concrete names via `clients/elasticsearch.DeleteIndex`, tolerating
    not-found (already deleted by ILM) as success; surface other failures.
  - Re-discover and repeat for any new members introduced by a rollover race, bounded by the
    `timeouts.delete` value (see task 1.1); on non-convergence, error and leave state intact for retry.
- [ ] 4.8 Create `resource.go` wiring the resource via `entitycore.NewElasticsearchResource[Model]`.
- [ ] 4.9 Implement `ImportState` using `entitycore.ParseCompositeID` to split off `<cluster_uuid>`,
  then split the remaining resource-ID portion on the first `/` into `rollover_alias` and `prefix`;
  reject malformed (not-three-part) import IDs with an actionable error.
- [ ] 4.10 Write unit tests in `rolloverindexseries_test.go` (or split per concern) covering:
  - Schema validation (naming validators wired correctly)
  - Create: missing-alias bootstrap, healthy-alias adoption (including out-of-prefix members),
    unhealthy-alias rejection (each reason), bootstrap-name collision without mutation
  - Read: healthy refresh, warning-only paths (no writer / non-numeric writer name), alias-absence
    state removal, API-error-is-not-absence
  - Update: state-only persistence of `deletion_protection`/timeouts, no mutating API call
  - Delete: protection block, concrete-name deletion (not alias/wildcard), hidden/closed members
    included, concurrently-deleted member tolerated, alias-absent no-op, out-of-prefix member deleted
  - Import: three-part ID parsing, malformed-ID rejection

## 5. Provider registration

- [ ] 5.1 Register `elasticstack_elasticsearch_rollover_index_series` in the Plugin Framework
  provider's resource list.
- [ ] 5.2 Add the new package import.

## 6. Acceptance tests

- [ ] 6.1 Create `internal/elasticsearch/index/rolloverindexseries/acc_test.go` with at least:
  - Bootstrap against a real ILM policy + index template (alias-only bootstrap inherits template
    settings); verify `write_index`/`indices` after apply.
  - Manual rollover via the ES API between applies; verify the next plan is a no-op for `write_index`
    (reflects the new writer) and does not try to restore the old writer or recreate `000001`.
  - Delete `000001` out of band (simulating ILM delete) while the alias remains on a newer
    generation; verify the next plan does not attempt to recreate `000001`.
  - Adoption: pre-create a healthy alias/index pair out of band, point a new resource config at it,
    verify Create performs no mutating call and adopts the existing member(s), including a
    deliberately out-of-prefix member name.
  - Adoption rejection: pre-create an alias with no effective writer (or a non-numeric writer name),
    verify Create fails without mutation, and verify Read on an already-adopted resource that becomes
    unhealthy produces a warning rather than an error.
  - Deletion protection: verify destroy fails while `deletion_protection = true`; apply
    `deletion_protection = false`; verify destroy then removes every current concrete member
    (including an out-of-prefix adopted member and, if feasible in CI, a hidden/closed member).
  - Import: bootstrap a series, roll it over at least once, import via the three-part ID, verify
    state matches a fresh Read.
  - If feasible per task 1.2, a shrink scenario verifying the shrunken replacement remains a member
    and is included in Delete's cleanup.
- [ ] 6.2 Add any required version-skip annotation per task 1.3's outcome.

## 7. Documentation

- [ ] 7.1 Create `templates/resources/elasticsearch_rollover_index_series.md.tmpl` documenting: the
  alias-is-the-boundary ownership model, that `deletion_protection` must be disabled via a prior apply
  before destroy, the three-part import ID format, and that index templates (not this resource) own
  settings/mappings and must not also define `rollover_alias` in their own `aliases` block.
  Explicitly document that automatic adoption extends this resource's delete scope to every
  pre-existing alias member, including ones that do not match `prefix`.
- [ ] 7.2 Add an example under `examples/resources/elasticstack_elasticsearch_rollover_index_series/`
  showing the resource alongside an `elasticstack_elasticsearch_index_lifecycle` and
  `elasticstack_elasticsearch_index_template` with a `depends_on` chain.
- [ ] 7.3 Run `make docs-generate` (or equivalent) to produce
  `docs/resources/elasticsearch_rollover_index_series.md`.

## 8. Validation

- [ ] 8.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-rollover-index-series --type change`.
- [ ] 8.2 Run `make build` once implementation lands (not part of this proposal-only change).
- [ ] 8.3 Run `go test ./internal/elasticsearch/index/rolloverindexseries/...` and
  `go test ./internal/clients/elasticsearch/...` (unit tests only) once implementation lands.
- [ ] 8.4 Run `make check-lint` once implementation lands.
- [ ] 8.5 Run targeted acceptance tests for the new resource against a live Elasticsearch cluster
  (requires `TF_ACC=1`) once implementation lands — deferred; not run as part of this proposal.
