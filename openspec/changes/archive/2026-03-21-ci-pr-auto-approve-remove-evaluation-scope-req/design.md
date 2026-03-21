## Context

Canonical behavior for `scripts/auto-approve/` is defined in `openspec/specs/ci-pr-auto-approve/spec.md`. REQ-001 normatively required ignoring drafts; the script implements that today. This change removes the draft filter in code and removes REQ-001 from the spec so both stay aligned.

## Goals / Non-Goals

**Goals:**

- Remove REQ-001 from the canonical spec and renumber remaining requirements so IDs stay sequential.
- Remove draft-PR exclusion from the evaluator and update tests accordingly.
- Keep all other requirement text and scenarios unchanged except for the `(REQ-NNN)` labels in requirement titles.

**Non-Goals:**

- Changing CI workflow triggers solely for this story (unless a workflow explicitly duplicates draft filtering and must be updated for consistency—evaluate case by case).
- Altering category or gate rules unrelated to draft scope (e.g. Copilot/Dependabot rules stay as specified).

## Decisions

- **Renumber after removal**: After deleting REQ-001, shift `REQ-002`…`REQ-014` down to `REQ-001`…`REQ-013` so references stay dense and predictable.
- **Delta spec shape**: Use `REMOVED Requirements` for the dropped requirement and `MODIFIED Requirements` for each retained requirement with full copied blocks and updated titles (per OpenSpec archive rules).

## Risks / Trade-offs

- **Stale references** — Docs or comments citing “REQ-001” as evaluation scope may need manual updates after archive. **Mitigation**: Note the renumbering in proposal impact; search for `REQ-00` when applying.

## Migration Plan

1. Land the change delta under `openspec/changes/.../specs/ci-pr-auto-approve/spec.md`.
2. Run `make check-openspec` (or `openspec validate`) before merge.
3. On archive, the canonical `openspec/specs/ci-pr-auto-approve/spec.md` is updated from the delta.

## Open Questions

- None.
