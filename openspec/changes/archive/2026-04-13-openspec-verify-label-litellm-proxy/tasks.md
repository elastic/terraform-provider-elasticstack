## 1. Update the verify workflow routing

- [x] 1.1 Update `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` so the verification job uses `engine.id: claude`, sets `engine.model` to `llm-gateway/gpt-5.4`, and configures Anthropic-compatible environment variables for `https://elastic.litellm-prod.ai`.
- [x] 1.2 Extend the workflow's AWF network policy so `network.allowed` includes `elastic.litellm-prod.ai` alongside the existing review-environment dependencies.
- [x] 1.3 Regenerate `.github/workflows/openspec-verify-label.md` and `.github/workflows/openspec-verify-label.lock.yml` from the authored workflow source.

## 2. Align requirements and workflow contract

- [x] 2.1 Sync the `ci-aw-openspec-verification` canonical spec with the approved Claude-routing and review-environment network requirements from this delta spec.
- [x] 2.2 Ensure the workflow sources any `ANTHROPIC_API_KEY` value from a GitHub Actions secret-backed expression rather than a checked-in literal.

## 3. Validate the migration

- [x] 3.1 Validate the new change and the updated specs with `npx openspec validate --changes openspec-verify-label-litellm-proxy` or an equivalent repository OpenSpec check.
- [x] 3.2 Run or inspect a representative `verify-openspec` workflow execution on a branch that includes the LiteLLM configuration to confirm the review job can reach `https://elastic.litellm-prod.ai` and starts with model `llm-gateway/gpt-5.4`.
