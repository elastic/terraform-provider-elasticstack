## Context

The preflight job already lists open PRs for the pushed branch and sets `should_run` for `push` vs `pull_request` / `workflow_dispatch`. The GitHub `push` webhook payload includes `commits` with `author.email` per commit.

## Goals / Non-Goals

**Goals:**

- For `push`, set `should_run=true` when **either** there is no open PR for the branch **or** every commit in the payload uses the Copilot coding agent author email (`198982749+Copilot@users.noreply.github.com`).
- When an open PR exists **and** any commit is not that Copilot email, skip the push workflow (dedupe with PR-driven CI).
- Leave non-`push` preflight rules unchanged.

**Non-Goals:**

- Defining behavior for tag pushes beyond what the workflow already filters (tags still ignored per workflow `on` config).
- Verifying commit signatures or GPG; email on the push payload is the source of truth for this gate.

## Decisions

- **Author field**: Use each commit’s `author.email` from `context.payload.commits` (not `committer`) so the rule matches “authored by” wording; document in implementation that amend/squash cases may differ if committer-only changes appear.
- **OR composition**: `should_run = noOpenPR || allCommitsHaveCopilotEmail` (same repository, same branch head as today’s PR listing).
- **Constant email**: Match exactly `198982749+Copilot@users.noreply.github.com` (case-sensitive as typical for emails in Git payloads).

## Risks / Trade-offs

- **[Risk] Copilot or GitHub changes the noreply email** → Mitigation: single constant in workflow; update spec and workflow together if the identity changes.
- **[Risk] Duplicate CI** → Mitigation: when a PR is open and commits are mixed, push CI skips; PR CI still runs. When a PR is open and all commits are Copilot, both may run—acceptable if rare or desired for agent workflows.
- **[Risk] Payload size / missing commits on huge pushes** → Mitigation: rare; if needed, fall back to Compare API (out of scope unless observed).

## Migration Plan

1. Land spec delta (this change) and sync to `openspec/specs/` when ready.
2. Update preflight script in `test.yml`.
3. Validate with `openspec validate` and a test push from Copilot vs local.

## Open Questions

- None for v1; confirm in staging that real Copilot pushes populate `author.email` as expected.
