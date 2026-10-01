## Why

The `OpenSpec verify (label)` agentic workflow (`.github/workflows/openspec-verify-label.md`) failed to activate for [PR #4948](https://github.com/elastic/terraform-provider-elasticstack/pull/4948) because gh-aw's per-workflow daily AI Credits guardrail (`max-daily-ai-credits`, default `5000`/24h rolling window) had already been exceeded (`5.2K` used against the `5K` threshold), per [run 35164355962](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/35164355962). The activation-time gate skipped the agent job entirely; the conclusion job filed issue #4964 with the standard failure report.

The workflow already disables the separate **per-run** guardrail (`max-ai-credits: -1`, see the comment at `.github/workflows/openspec-verify-label.md` lines 76-86) because the OpenRouter model slug it uses is absent from gh-aw's built-in AI-credits pricing table, so with that guardrail active every request is rejected outright. That existing comment explicitly notes "The daily guardrail (`max-daily-ai-credits`, default 5000/day) still applies" — this run's failure is exactly that still-active daily cap now blocking legitimate `verify-openspec` activity on a routine PR.

On the triggering issue, maintainer `@tobio` gave explicit direction: disable the daily guardrail outright (`max-daily-ai-credits: -1`), rather than raising it to a larger finite threshold as the generic failure-report guidance suggests.

## What Changes

- Add `max-daily-ai-credits: -1` to the top-level frontmatter of `.github/workflows/openspec-verify-label.md`, alongside the existing `max-ai-credits: -1`, fully disabling gh-aw's per-workflow daily AI-credits guardrail for this workflow.
- Update the existing explanatory comment (currently ending "The daily guardrail (`max-daily-ai-credits`, default 5000/day) still applies.") so it documents both overrides and why each is disabled, keeping the two related guardrail decisions co-located for future maintainers.
- Regenerate `.github/workflows/openspec-verify-label.lock.yml` via `gh aw compile` so the compiled lock file matches the new source (per REQ-001 of `ci-aw-openspec-verification`, source and lock must stay paired).
- No change to the workflow's trigger, permissions, safe-outputs, or verification/review/archive agent behavior — this is a guardrail-configuration change only.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-aw-openspec-verification`: add a requirement documenting that this workflow's daily AI-credits guardrail is disabled (`max-daily-ai-credits: -1`), so the workflow does not skip agent activation on account of accumulated 24-hour AI Credits usage.

## Impact

- `.github/workflows/openspec-verify-label.md` — frontmatter change (add `max-daily-ai-credits: -1`; update the adjacent guardrail-explanation comment).
- `.github/workflows/openspec-verify-label.lock.yml` — regenerated compiled artifact via `gh aw compile`.
- `openspec/specs/ci-aw-openspec-verification/spec.md` — sync the new requirement once implemented.
