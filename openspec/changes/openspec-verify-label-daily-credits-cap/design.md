## Context

`.github/workflows/openspec-verify-label.md` is a GitHub Agentic Workflow (gh-aw) that verifies pull requests against an active OpenSpec change and, on approval, archives it. gh-aw enforces two independent AI-credits guardrails:

- `max-ai-credits` — a per-run cap.
- `max-daily-ai-credits` — a per-workflow cap aggregated across all runs in a trailing 24-hour window (default `5000`), evaluated at activation time. Exceeding it causes the activation job to skip the agent job for subsequent runs until the window rolls off; it does not cancel an in-progress run.

The workflow already sets `max-ai-credits: -1` (line 86) because the OpenRouter model slug `anthropic/claude-opus-5` it uses is absent from gh-aw's AWF api-proxy's built-in AI-credits pricing table, so with the per-run guardrail active every request is rejected with HTTP 400 (`unknown_model_ai_credits`); see [github/gh-aw#47365](https://github.com/github/gh-aw/issues/47365) and pending fix [github/gh-aw#47571](https://github.com/github/gh-aw/pull/47571). The comment documenting that override explicitly notes the daily guardrail "still applies" — and issue #4964 is that daily guardrail now blocking a routine PR (`5.2K` used against a `5K` threshold).

Maintainer `@tobio` gave explicit direction on the issue: disable the daily guardrail (`max-daily-ai-credits: -1`), rather than raise it to a larger finite value.

## Goals / Non-Goals

**Goals:**
- Disable (not merely raise) the daily AI-credits guardrail for this one workflow, per explicit maintainer direction.
- Keep the change mechanical and scoped to guardrail configuration; no change to verification logic, triggers, permissions, or safe-outputs.
- Keep the per-run and per-day guardrail overrides documented together so a future maintainer investigating cost controls for this workflow sees both decisions and their rationale in one place.

**Non-Goals:**
- Changing guardrails for any other agentic workflow file in this repository.
- Introducing an alternate cost-control mechanism (e.g. raising to a finite ceiling like `20K`) — the maintainer direction was an explicit disable, not a larger cap.
- Revisiting or fixing the underlying OpenRouter pricing-table gap that motivated the existing `max-ai-credits: -1` override; that is tracked upstream in gh-aw.

## Decisions

- **Decision 1: Set `max-daily-ai-credits: -1` at the top level of `.github/workflows/openspec-verify-label.md` frontmatter, alongside the existing `max-ai-credits: -1`.** Rationale: this is the documented gh-aw mechanism to fully disable the per-workflow daily guardrail (per the failure report's "How to disable this guardrail" section), and matches the maintainer's explicit instruction on the issue.
  - *Alternative considered:* raise the threshold to a large but finite value (e.g. `max-daily-ai-credits: 20K`) per the failure report's generic "How to raise the daily limit" guidance. Rejected because the maintainer's direction was explicit disable, not a higher cap.
- **Decision 2: Extend the existing guardrail-explanation comment rather than add a second, separate comment block.** Rationale: the current comment already explains why the per-run cap is disabled and explicitly flags that the daily cap "still applies" — updating it in place keeps both guardrail decisions co-located and avoids a stale, now-incorrect statement lingering in the source.
- **Decision 3: Regenerate `.github/workflows/openspec-verify-label.lock.yml` via `gh aw compile` as an implementation task, not part of this proposal PR.** Rationale: this OpenSpec change-factory run is proposal-only (no `gh aw compile` invocation, no workflow edits in this PR); REQ-001 of `ci-aw-openspec-verification` already requires the compiled lock file to stay paired with its source, so regeneration is tracked as a follow-up implementation task.

## Open questions

- Should this fully-disabled daily guardrail be revisited with a large-but-finite ceiling if `verify-openspec` PR volume grows substantially, rather than remaining disabled indefinitely? Left for maintainers to reassess if usage patterns change; out of scope for this change.
