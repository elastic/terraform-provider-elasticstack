## Why

`elasticstack_elasticsearch_index_alias.read_indices[*].name` accepts a wildcard such as `traces-apm*` at the Elasticsearch API level. The `add` action succeeds and the alias is attached to every matching index, but the provider then reads the alias back as one `read_indices` element per concrete index. The planned element (`name = "traces-apm*"`) correlates with none of them, so Terraform fails the apply with `Provider produced inconsistent result after apply` even though the alias was created and works. Subsequent plans and updates would also churn, because current and planned entries are keyed by different strings.

Source: [elastic/terraform-provider-elasticstack#5027](https://github.com/elastic/terraform-provider-elasticstack/issues/5027). A maintainer comment on the issue records the agreed direction: support wildcard and Elasticsearch multi-target expressions in `read_indices`, and keep `write_index.name` limited to a single target.

## What Changes

- `read_indices[*].name` accepts any Elasticsearch multi-target expression (`*`, `?`, comma-separated lists, `-` exclusions, `_all`). The user's original expression stays in `name`.
- Add a computed `concrete_indices` set of strings inside each `read_indices` element. It holds the concrete indices and data streams currently attached to the alias for that expression.
- The provider resolves expressions (with `expand_wildcards=all` and `allow_no_indices=true`) during plan (`ModifyPlan`), create, read and update:
  - Plan shows an in-place update when a resolved target is not yet attached to the alias.
  - Create and update attach the alias to every currently resolved target, re-resolving at apply time so targets created between plan and apply are included.
  - Read stores, per element, the intersection of the resolved targets and the alias's actual members.
  - Delete removes the alias from its live concrete members instead of from selector strings or stale state.
- An expression with no current match is allowed; later plans re-resolve it and attach newly matching targets.
- Alias members not covered by any configured `read_indices` expression remain drift: read surfaces them as concrete singleton `read_indices` entries so the next plan removes them.
- Resolved targets that are aliases or remote-cluster targets are rejected with a clear diagnostic. Overlapping expressions with identical settings are deduplicated; conflicting settings on the same target are a configuration error.
- `write_index.name` is validated to reject multi-target syntax (wildcards, commas, exclusions, `_all`) while continuing to accept currently valid single-target names.
- Update diffing compares against the live alias response per concrete target rather than selector strings, avoiding remove/add churn.
- Update the schema descriptions, generated docs, and add unit and acceptance tests.

Non-goals:

- No automatic attachment of the alias to indices created after the last apply; this happens on the next plan/apply (use index templates for continuous attachment).
- No detection of out-of-band filter, routing or hidden-setting changes on an already attached target when membership is unchanged. Such changes are applied whenever another update is required.
- No changes to alias identity or import format.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `elasticsearch-index-alias`: `read_indices` gains selector-expression semantics and a computed `concrete_indices` attribute; validation, create, update, read and delete requirements change accordingly; `write_index.name` gains single-target validation.

## Impact

- `internal/elasticsearch/index/alias/` (schema, models, plan modification, create/update/read/delete), possibly a new Resolve Index / index-expression helper in `internal/clients/elasticsearch/`.
- Resource schema change (new nested computed attribute) requires a state-compatibility strategy for existing state that lacks `concrete_indices`.
- Regenerated docs for `elasticstack_elasticsearch_index_alias` and a changelog entry.
- Acceptance tests that create multiple indices, including hidden and closed ones.
