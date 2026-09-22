## 1. Update workflow frontmatter

- [x] 1.1 Add `max-daily-ai-credits: -1` to the top-level frontmatter of `.github/workflows/openspec-verify-label.md`, alongside the existing `max-ai-credits: -1`.
- [x] 1.2 Update the explanatory comment above `max-ai-credits: -1` (currently ending "The daily guardrail (max-daily-ai-credits, default 5000/day) still applies.") so it documents both the per-run and per-day overrides and why each is disabled.

## 2. Regenerate compiled workflow

- [x] 2.1 Run `gh aw compile` to regenerate `.github/workflows/openspec-verify-label.lock.yml`.
- [x] 2.2 Confirm the regenerated lock file contains no unrelated diff beyond the guardrail change.

## 3. Spec sync and validation

- [x] 3.1 Sync the `ADDED Requirements` delta in this change's `specs/ci-aw-openspec-verification/spec.md` into `openspec/specs/ci-aw-openspec-verification/spec.md` once the workflow frontmatter change is implemented.
- [x] 3.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate openspec-verify-label-daily-credits-cap --type change` and resolve any reported issues.
