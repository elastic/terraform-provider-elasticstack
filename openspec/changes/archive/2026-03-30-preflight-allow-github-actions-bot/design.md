## Context

The preflight job already decides whether push-triggered CI should run when a branch also has an open pull request. Today the requirements and workflow logic recognize only the Copilot coding agent noreply email when evaluating whether all commits in a push were authored by an allowed bot identity.

## Goals / Non-Goals

**Goals:**

- Allow the preflight gate's bot-author exception to treat GitHub Actions-generated commits the same as Copilot-generated commits.
- Keep the existing `should_run` decision structure intact: no open PR remains sufficient, and non-`push` events keep their current behavior.
- Make the allowed-user list explicit in the spec and easy to mirror in workflow code.

**Non-Goals:**

- Changing how open pull requests are discovered.
- Expanding the exception to arbitrary bot accounts or domains.
- Changing downstream job gating, permissions, or ready-for-review behavior.

## Decisions

- **Allowed identity list**: Define the exception in terms of an explicit allowed-email set containing `198982749+Copilot@users.noreply.github.com` and `41898282+github-actions[bot]@users.noreply.github.com`. This keeps the rule concrete and testable in both spec and workflow code.
- **Author source**: Continue to reason from commit author emails in the push payload, because the current requirement is phrased in terms of who authored the commits rather than who pushed them.
- **Minimal behavioral delta**: Change only the author matching rule. The preflight gate still evaluates `push` events first by open-PR state and still skips downstream jobs when an open PR exists and at least one commit is not on the allowed list.

## Risks / Trade-offs

- **[Risk] Bot identities change again** -> Mitigation: keep the list centralized in the workflow logic and update the spec alongside any future identity change.
- **[Risk] Ambiguity over author vs committer** -> Mitigation: preserve the current authored-by framing so the rule remains consistent with existing requirements.
- **[Risk] Duplicate push and PR CI for allowed bot pushes** -> Mitigation: accepted as existing behavior for this exception; this change only broadens the allowed bot set by one known identity.

## Migration Plan

1. Add a delta spec for `ci-build-lint-test` that replaces the Copilot-only language with an explicit allowed-user list.
2. Update `.github/workflows/test.yml` so the preflight script accepts either allowed email.
3. Validate the change with OpenSpec checks and targeted workflow verification.

## Open Questions

- None.
