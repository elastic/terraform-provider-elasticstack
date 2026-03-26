## Context

The Copilot auto-approve gate compares `additions + deletions` from the GitHub pull request payload against a fixed constant (`maxEditedLines` in `scripts/auto-approve/evaluator.go`). The canonical requirement is REQ-009 in `openspec/specs/ci-pr-auto-approve/spec.md`. Today both are set to 300.

## Goals / Non-Goals

**Goals:**

- Raise the threshold to 1000 in spec, code, and tests so behavior stays single-sourced and verifiable.

**Non-Goals:**

- Making the limit configurable via environment or config file.
- Changing Dependabot behavior or other Copilot gates (authors, file allowlist).

## Decisions

- **Use the same constant and `fmt.Sprintf` reason string as today** — keep `maxEditedLines = 1000` and `edited lines must be < %d` so operators and tests only see the number change. No new observability fields.
- **Boundary semantics unchanged** — approval requires `additions + deletions < maxEditedLines` (strict less-than), so 1000 total edits still fails when the limit is 1000.

## Risks / Trade-offs

- **Larger auto-approved Copilot PRs** → Mitigation: file allowlist (`*_test.go`, `*.tf`) and author gates unchanged; this only widens the size window for those paths.

## Migration Plan

- Deploy script change together with workflow usage; no data migration. After implementation, sync or archive the change so `openspec/specs/ci-pr-auto-approve/spec.md` matches the delta.

## Open Questions

- None.
