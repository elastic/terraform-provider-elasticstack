# Design: Require `searchable_snapshot` in the ILM `frozen` phase

## Context

The ILM resource already models `searchable_snapshot.snapshot_repository` as required when a `searchable_snapshot` block is present, but it does not currently express the stronger rule that the `frozen` phase itself must include `searchable_snapshot`.

That leaves the provider in this state:

```text
frozen block present
        │
        ├─ without searchable_snapshot ──▶ provider accepts config
        │                                  Elasticsearch rejects apply
        │
        └─ with searchable_snapshot ─────▶ valid
```

The bug is partly documentation, but the underlying problem is schema/validation fidelity. If the schema continues to treat `frozen.searchable_snapshot` as optional, generated docs will continue to be ambiguous and users will keep discovering the rule too late.

## Goals

1. Make the Terraform schema express that `searchable_snapshot` is mandatory inside `frozen`.
2. Fail invalid `frozen` configurations before any ILM API call.
3. Align generated docs, acceptance tests, and OpenSpec requirements with Elasticsearch behavior.

## Non-Goals

- Changing `searchable_snapshot` behavior in `hot` or `cold`.
- Changing the existing "required when block is present" rule for `searchable_snapshot.snapshot_repository`.
- Altering ILM API expansion or read-back behavior for already valid policies.

## Decisions

### Model the `frozen.searchable_snapshot` block as required

The simplest and most accurate fix is to make the `searchable_snapshot` nested block required within `phaseFrozenBlock()`.

That gives the provider two benefits at once:

- schema-driven validation rejects `frozen {}` without the action block
- generated docs describe the nested block as required

The existing object-level `AlsoRequires` validator on `snapshot_repository` remains in place so the provider still enforces:

```text
frozen.searchable_snapshot present
        │
        └─ snapshot_repository missing ──▶ invalid
```

### Keep an explicit validation scenario in the requirements

Even with schema-driven enforcement, the canonical requirements should state the rule directly. The OpenSpec delta should capture that:

- `frozen` is the only phase whose sole supported action is mandatory when the phase is declared
- omission is rejected before any Elasticsearch API call

### Cover the behavior with validation-focused tests

Testing should cover both sides:

- a valid `frozen` phase with `searchable_snapshot`
- an invalid `frozen` phase without `searchable_snapshot`, asserting a Terraform validation error

This complements the existing acceptance test that already exercises a valid frozen phase.

## Risks and Trade-offs

| Risk | Mitigation |
|------|------------|
| The stricter schema could surprise users who previously relied on apply-time API failure | This only rejects configurations Elasticsearch already considers invalid, and it improves feedback timing |
| Generated docs may still need regeneration after the schema fix | Include documentation regeneration explicitly in the task list |

## Migration and State

No state migration is required. Existing valid `frozen` configurations continue to work unchanged. Existing invalid configurations become plan-time errors instead of apply-time errors.
