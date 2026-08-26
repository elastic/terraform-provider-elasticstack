## Context

`elasticstack_elasticsearch_index_lifecycle` models each ILM phase as a Plugin Framework `SingleNestedBlock`. The `rollover` action within the `hot` phase is already `Optional` at the schema level (`schema_actions.go:blockRollover`), and `expand.go`'s `ilmActionRollover` case already omits the action from the API payload when the user has not configured it. So the write path is correct today.

The read path is where the bug lives. `flattenPhase` (`flatten.go`) iterates over the actions map returned by Elasticsearch and, for any action name it does not special-case, falls into a `default:` branch that writes the raw action map into the phase unconditionally:

```go
default:
    phase[actionName] = []any{action}
```

When Elasticsearch returns `"rollover": {}` for a hot phase — observed in practice even when the user's PUT request omitted rollover entirely — `action` is an empty `map[string]any{}`. That gets converted by `phaseDataToObjectAttrs`/`anyToAttr` (`value_conv.go`) into a `types.Object` whose attributes are all null, which is a distinct Terraform value from `types.ObjectNull(rolloverObjectType().AttrTypes)`. Terraform Core treats "object with all-null attributes" and "null object" as different values, so every `terraform plan` after refresh reports a change on `hot.rollover`, even though the user's configuration and the API payload agree that no rollover exists.

This is the same class of bug fixed for `allocate` in the existing REQ-034 (omit replica/shard defaults when not configured), and for the toggle actions (`readonly`, `freeze`, `unfollow`) via REQ-029's `priorHasDeclaredToggle` guard.

## Goals / Non-Goals

**Goals:**
- Allow a `hot` phase with no `rollover {}` block to remain diff-free across repeated plans, matching index types (e.g. `lookup`) where rollover is not applicable.
- Reuse the existing prior-state-declaration guard pattern rather than introducing a new mechanism.
- Make no schema changes: `rollover` remains `Optional`, no new attributes, no schema version bump.

**Non-Goals:**
- Changing `rollover` from `Optional` to `Required`/`Computed` — the schema is already correct.
- Adding a `skip_rollover` (or similarly named) explicit attribute. This was considered and rejected: the ILM API has no such concept, the fix is purely a flatten-path normalization, and an explicit attribute would require a schema/state version bump for no functional gain.
- Changing how `warm`, `cold`, or `delete` phases handle their actions (they do not support `rollover`).
- Changes to `elasticstack_elasticsearch_data_stream_lifecycle`.

## Decisions

- **Add an explicit `rollover` case in `flattenPhase`.** When the action name is `rollover` and the returned action is empty (`len(action) == 0`), treat it as absent (skip writing it into `phase`) unless the prior state already declared a non-null `rollover` object for this phase. This exactly mirrors the existing `priorHasDeclaredToggle` check already used for `readonly`/`freeze`/`unfollow`, generalized (or duplicated as a sibling helper) to accept `rollover` as the action name.
- **Preserve explicitly-configured empty `rollover {}` blocks.** If a user's prior state already has a non-null `rollover` object (e.g. they declared `rollover {}` with no conditions, or any conditions), the guard keeps writing `phase[actionName] = []any{action}` so we don't silently drop a user-declared block. This mirrors the conservative behavior already used for toggle actions.
- **No change to `expand.go`.** The write path already omits `rollover` from the API payload correctly when unconfigured; only the read path needs the fix.

## Open questions

- Does Elasticsearch **always** return `rollover: {}` for a hot phase that was PUT without rollover, or only for data-stream-backed indices / certain ES versions? Confirming this against the live acceptance test environment would validate that this normalization resolves the reporter's exact scenario. If ES does NOT inject a default rollover, the perpetual diff may instead be an artefact of an older SDK-based state (pre-PF migration) — in which case the same fix still applies but the acceptance test is more important.
- Should an `import` of an existing policy with `rollover: {}` (explicitly-configured empty rollover) produce a non-null rollover in state, or null? If the consensus is "null" (i.e., an empty rollover is always treated as absent), the prior-state guard can be simplified/removed.
- Are there non-data-stream use cases where `rollover: {}` (no conditions) in a hot phase is intentional and meaningful? If yes, the prior-state guard must be retained as designed above.

## Risks / Trade-offs

- [Risk] Import of a policy with an intentionally-empty `rollover {}` (no conditions) loses that declaration in state, since import has no prior state to consult. Mitigation: this is a pre-existing ambiguity noted in the open questions above; an unconditional `rollover {}` with no trigger conditions is not a meaningful real-world ILM configuration, so the impact is expected to be near-zero. Revisit if acceptance testing surfaces a real scenario.
- [Risk] The exact conditions under which Elasticsearch returns `rollover: {}` for a hot phase without rollover are not fully confirmed against a live cluster. Mitigation: the acceptance test added by this change exercises the real scenario end-to-end against the configured test stack, which will surface any mismatch between this design's assumption and actual API behavior.
