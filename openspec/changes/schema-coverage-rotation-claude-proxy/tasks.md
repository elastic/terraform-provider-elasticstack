## 1. Update the schema-coverage rotation workflow contract

- [ ] 1.1 Update `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` so the workflow switches from `engine.id: copilot` to `engine.id: claude` and configures `ANTHROPIC_BASE_URL` for `https://elastic.litellm-prod.ai/`.
- [ ] 1.2 Ensure any configured `ANTHROPIC_API_KEY` value is sourced from a GitHub Actions secret-backed expression, add `tools.timeout: 300`, and extend `network.allowed` with `elastic.litellm-prod.ai`.
- [ ] 1.3 Regenerate `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml` from the authored workflow source without changing unrelated workflow behavior.

## 2. Align tests and requirements

- [ ] 2.1 Update workflow generation tests under `.github/workflows-src/lib/` to assert the Claude engine, Anthropic proxy environment, explicit tool timeout, and LiteLLM host allowlist.
- [ ] 2.2 Sync the canonical `ci-schema-coverage-rotation-engine` and `ci-schema-coverage-rotation-toolchain` specs with the approved workflow contract from this change.

## 3. Validate the migration

- [ ] 3.1 Validate the OpenSpec artifacts with `npx openspec validate schema-coverage-rotation-claude-proxy --type change` or an equivalent repository OpenSpec check.
- [ ] 3.2 Run workflow generation and the relevant workflow tests (for example `make workflow-generate` and `make workflow-test`) to confirm the authored and generated artifacts stay in sync.
- [ ] 3.3 Run or inspect a representative `schema-coverage-rotation` workflow execution that uses the Claude + LiteLLM configuration to confirm the worker reaches `https://elastic.litellm-prod.ai/` successfully.
