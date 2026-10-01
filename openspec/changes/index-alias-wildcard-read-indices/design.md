## Context

See `proposal.md` for motivation and `specs/elasticsearch-index-alias/spec.md` for requirements.

Current behavior (`internal/elasticsearch/index/alias`):

- `read_indices` is a `set` of nested objects whose identity is the configured `name`. `populateFromAPI` builds one element per concrete index returned by Get Alias, so a wildcard in config never correlates with read-back state.
- `update.go` keys current entries by concrete index name and planned entries by the configured string, so wildcards would cause remove/add churn on every update.
- `delete.go` builds `remove` actions from state element names, which would be selector strings.
- The resource is built on `entitycore.ElasticsearchResource[tfModel]`, which does not implement `ModifyPlan` today; the alias resource will add its own `ModifyPlan`.

## Goals / Non-Goals

**Goals:**
- Wildcard and multi-target expressions in `read_indices` apply cleanly and converge to an empty plan.
- Membership drift (a matching target not attached, or an attached target not configured) shows up in the plan.
- No remove/add churn for already attached targets.

**Non-Goals:**
- Continuous attachment of future indices outside of plan/apply.
- Drift detection for filter/routing/hidden changes on an attached target when membership is unchanged.
- Wildcards in `write_index.name`.

## Decisions

### Deviation from the implementation research recommendation

The implementation research comment recommended Approach A (reject wildcard and multi-target names at validation time). The maintainer comment on the issue explicitly decides otherwise: wildcards and multi-target expressions are supported in `read_indices`, while `write_index.name` stays single-target. This change follows the maintainer decision. Approach A's validator survives only for `write_index.name`. Approach B (collapse resolved indices back onto the pattern) is also not used; the `concrete_indices` attribute replaces that mechanism and avoids relying on a prior-state reference at read time.

### State model: `concrete_indices` inside each `read_indices` element

Each `read_indices` element gains `concrete_indices` (computed `set(string)`). `name` keeps the user's expression; `concrete_indices` is the set of concrete targets currently attached to the alias for that expression. Because `read_indices` is a set, elements remain identified by their configured values (`name` plus settings); there is no positional index.

- **Plan (`ModifyPlan`)**: resolve every `read_indices[*].name` and compare the desired targets with state. When creation or an in-place update is required, leave the affected computed `concrete_indices` values unknown. Terraform therefore shows the resource update without constraining the final membership to a stale plan-time set. When no update is needed, preserve the current concrete membership in the plan.
- **Apply (create/update)**: re-resolve the expressions, compute actions against the live alias response, then read back. Because the affected computed values were unknown, a target created between plan and apply can be attached and recorded without producing an inconsistent result.
- **Read**: resolve the configured expression, fetch the alias's actual associations, and store the intersection as that element's `concrete_indices`. Storing the whole resolved set would hide a missing attachment and prevent Terraform from planning it.
- **Unconfigured members**: alias members not covered by any configured expression are emitted as singleton `read_indices` entries (`name` equals the concrete index, `concrete_indices` equals `{name}`), preserving the current full-ownership semantics so the next plan removes them.
- **Write index**: the alias's write index remains modeled only by `write_index`; it is excluded from `read_indices` coverage when classifying members.

The implementation must not assume that a `SetNestedAttribute` ignores its computed child when correlating elements. Its planning behavior must be verified with framework and acceptance tests for partial membership, unknown planned membership, and the final post-apply state.

### Virtual state for an empty desired alias

Elasticsearch has no persistent alias object without an associated target. When there is no `write_index` and every `read_indices` expression resolves to an empty set, create and update retain a virtual Terraform resource state instead of attempting to create an empty alias. The state retains the configured expressions and empty `concrete_indices` sets.

On read, a missing alias is retained only when the prior state is virtual and the desired membership is still empty. A missing alias with a write index or any resolved read target remains normal not-found drift and removes the resource from state. When a later plan resolves a target, the virtual resource receives an in-place update that creates the alias association. Deleting virtual state makes no Update Aliases API call.

### Selector resolution

- Expressions are resolved with the Elasticsearch Resolve Index API (`GET /_resolve/index/<expr>`) or an equivalent typed-client call, with `expand_wildcards=all` so open, closed and hidden targets are included, and `allow_no_indices=true` so expressions with no match return an empty result instead of an error.
- Only resolved `indices` and `data_streams` entries are accepted. If the expression resolves to aliases, or contains remote-cluster (`cluster:index`) targets, the provider returns an error diagnostic naming the offending target, since neither can be a member of this alias.
- A data stream target is attached by its data stream name, matching what Update Aliases accepts.
- Resolution runs through a single helper used by plan, create, update and read so all four see identical semantics.

### Action computation

For create and update:

1. Resolve each `read_indices` expression to concrete targets, producing one `IndexConfig` per (target, element).
2. Merge by target. If two expressions resolve to the same target with equal settings, keep one config. If settings differ, return a configuration error identifying the target and both expressions. A target that is also the `write_index` is a configuration error (extending REQ-009 to resolved targets).
3. Compare against the live Get Alias response keyed by concrete name: `remove` members not in the desired set, `add` missing members or members whose settings differ, skip identical ones.
4. Submit all actions in one atomic Update Aliases call; skip the call when there are no actions.

Delete uses the live Get Alias response to build `remove` actions for every current member, rather than state element names.

### Validation

- `write_index.name` rejects multi-target selector syntax: `*`, `?`, `,`, a leading `-` (exclusion), and `_all`. It must not reject currently valid single-target names, so the check is syntax-based on those tokens only (no API call during `ValidateConfig`; values unknown at validate time are skipped).
- REQ-009 (write index not also a read index) stays at config time on literal equality; after resolution it is re-checked against resolved targets during plan and apply.
- `read_indices[*].name` must be non-empty; no other syntax restrictions are added since Elasticsearch validates expressions.

### Schema and state compatibility

`concrete_indices` is `Computed` `set(string)` inside the existing `read_indices` nested object. Existing state lacks this field; the schema change is additive, so the framework fills the missing attribute with null on first load and the first refresh and plan must treat null as "unknown membership" rather than an error. Decide in implementation whether a schema `Version` bump with a `StateUpgrader` is needed or whether null-tolerant read plus plan-time population is enough; either way, an existing-state acceptance test (apply with the previous provider version, then upgrade) must pass with a clean follow-up plan. Import continues to use passthrough id; read after import must produce `read_indices` entries consistent with the alias's live members (singletons, since no configured expression exists yet), and a subsequent plan against wildcard config shows the expected in-place update.

### Documentation

Update the `read_indices.name` and `write_index.name` descriptions, add a `concrete_indices` description, regenerate docs, and add a changelog entry noting wildcard support and the new computed attribute.

## Risks / Trade-offs

- **Plan-vs-apply consistency**: Terraform requires post-apply values to match planned known values. Create and changing updates therefore plan affected computed memberships as unknown, and acceptance tests verify a target created between plan and apply is attached without an inconsistent result.
- **Set element correlation**: `concrete_indices` is a computed child of a `read_indices` set element. Unknown-on-change planning avoids constraining its final value, but framework and acceptance tests must prove that partial memberships and final read-back correlate correctly.
- **Virtual empty alias state**: The provider represents a configured but targetless alias even though Elasticsearch has no alias object. Read must retain this state only while membership remains empty; otherwise it would mask real deletion drift.
- **Plan-time API calls**: `ModifyPlan` now issues a Resolve Index call. When the cluster is unreachable at plan time the plan fails with a clear diagnostic; unknown `name` values skip resolution and leave `concrete_indices` unknown.
- **Narrow drift detection**: out-of-band filter/routing/hidden changes on an attached target do not trigger a plan on their own. This is a documented limitation.
- **Large expansions**: `_all` or broad wildcards can produce many actions in one atomic request; accepted for now.

## Migration Plan

Additive schema change. Existing configurations with concrete `read_indices` names keep working: their `concrete_indices` is the single name. No config changes are required. Existing wildcard users who currently hit the apply error get a working resource.

## Open questions

- Implementation research open questions, copied verbatim for traceability. Resolved by the maintainer comment on the issue, summarized after each:
  - Is wildcard support in `read_indices.name` an intentional, supported use case, or only incidental behavior that happened to work at the API level? The answer decides between A and B. *(Resolved: supported.)*
  - Does anyone rely on today's behavior (alias created despite the apply error) such that rejecting wildcards in a minor release would be considered breaking? Is a changelog "breaking change" note acceptable? *(Moot: wildcards are not rejected for `read_indices`.)*
  - Should the validator also reject comma-separated lists, `_all`, exclusions (`-index`) and date-math expressions, or only `*` and `?`? *(Resolved for `write_index.name`: reject `*`, `?`, comma lists, exclusions, `_all`. Date-math handling in `write_index.name` is not specified; the implementer should confirm whether date-math should be rejected or passed through.)*
  - If B is chosen: what should plan and import do when config holds a pattern but state holds the expanded set? *(Superseded by the `concrete_indices` design.)*
  - Should a pattern-based alias be re-evaluated on update, so newly created matching indices get attached, or remain a one-time expansion? *(Resolved: re-evaluated on every plan and apply.)*
- Is a `StateUpgrader`/schema version bump required for adding `concrete_indices`, or is null-tolerant handling sufficient?
- Should date-math expressions (for example `<logs-{now/d}>`) be rejected in `write_index.name`?
