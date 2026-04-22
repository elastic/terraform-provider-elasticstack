## Context

Canonical requirements for this resource live in [`openspec/specs/elasticsearch-index-lifecycle/spec.md`](../../specs/elasticsearch-index-lifecycle/spec.md). Implementation lives in [`internal/elasticsearch/index/ilm/`](../../../internal/elasticsearch/index/ilm/). Elasticsearch policy expansion today expects each action as a **one-element slice** in the internal map (`expand.go`); the Framework layer can keep producing that shape from **object**-typed action attributes by wrapping in `[]any{m}` inside `attrValueToExpandRaw` (or equivalent).

## Goals / Non-Goals

**Goals:**

- Align Plugin Framework schema with **single nested block** semantics for phases and ILM actions so documentation and Terraform state match SDK “max one” intent.
- **Migrate existing state** from list-encoded nested values to object-encoded values without manual intervention.
- Enforce **required-when-present** action fields via **`objectvalidator.AlsoRequires`** after making those attributes optional.
- Preserve **readonly / freeze / unfollow** flatten semantics that depend on prior state (`priorHasDeclaredToggle`).

**Non-goals:**

- Changing **`elasticsearch_connection`** block type or connection resolution (REQ-009–REQ-010).
- Changing Elasticsearch API payloads beyond what is needed to preserve behavior after schema/type refactors.
- Reworking rollover “at least one condition” style rules unless already required by schema (rollover remains all-optional attributes at the Framework level today).

## Decisions

- **SingleNestedBlock** for `hot`, `warm`, `cold`, `frozen`, `delete` and for every ILM action block under those phases; **list nested block** only for `elasticsearch_connection`.
- **State upgrade**: Raw JSON map walk; unwrap `[]any` with `len >= 1` → first element for **known** keys only (phases at root; action keys per phase; include delete-phase inner `delete` action block name). Empty list → omit / null consistent with new state. **`len > 1`**: use first element (pragmatic recovery from invalid state).
- **AlsoRequires**: Follow [`internal/kibana/alertingrule/schema.go`](../../../internal/kibana/alertingrule/schema.go) pattern (`path.MatchRelative().AtName(...)`). Attach validators on **`SingleNestedBlock`** where the framework allows; otherwise on **`NestedBlockObject`** if any action stays list-wrapped during transition.
- **AlsoRequires field set**:

  | Action | Paths |
  |--------|--------|
  | `forcemerge` | `max_num_segments` |
  | `searchable_snapshot` | `snapshot_repository` |
  | `set_priority` | `priority` |
  | `wait_for_snapshot` | `policy` |
  | `downsample` | `fixed_interval` |

## Risks / Trade-offs

- **Acceptance test churn**: All `TestCheckResourceAttr` paths that assume list indices for phases/actions must be updated.
- **State upgrade bugs**: Missing or mistyped keys could leave stale list-shaped state; mitigate with unit tests and manual upgrade test against a v0 state fixture.
- **Empty action block**: Validation must fail with a clear diagnostic when AlsoRequires is violated.

## Migration Plan

1. Land implementation and delta spec under this change.
2. Run `make check-openspec` / `openspec validate` on the change (when CLI available).
3. After review, **sync** delta into `openspec/specs/elasticsearch-index-lifecycle/spec.md` (or archive per workflow).
4. Release note: mention state upgrade and doc shape for nested blocks.

## Open Questions

- None.
