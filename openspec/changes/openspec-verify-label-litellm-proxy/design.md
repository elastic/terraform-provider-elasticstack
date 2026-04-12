## Context

The `openspec-verify-label` workflow is already structured around GitHub Agentic Workflows with deterministic pre-activation gates, a Copilot engine, and GitHub-native review and archive safe outputs. Today, the workflow frontmatter still selects a GitHub-hosted Copilot model directly, which means the repository cannot centralize model routing, observability, or provider policy through the Elastic LiteLLM gateway.

This change is intentionally narrow: it affects only the verification workflow and only the model-inference path. The repository wants to preserve the existing `copilot` engine contract while routing Copilot CLI through the OpenAI-compatible LiteLLM endpoint `https://elastic.litellm-prod.ai/v1` and updating the workflow model selector to `llm-gateway/gpt-5.4`.

## Goals / Non-Goals

**Goals:**
- Keep `openspec-verify-label` on the `copilot` engine rather than redesigning the workflow around a different GH AW engine.
- Route verification-model traffic through the Elastic LiteLLM proxy using documented Copilot CLI BYOK provider settings.
- Change the workflow model selection to `llm-gateway/gpt-5.4`.
- Make the required provider endpoint, secret-backed auth handoff, and AWF network allowlist explicit in the OpenSpec contract.
- Preserve existing GitHub review, archive, push, and label-cleanup behavior.

**Non-Goals:**
- Migrating other GH AW workflows, including `schema-coverage-rotation`, in the same change.
- Changing review disposition rules, archive gating, or the deterministic pull-request classification logic.
- Switching the workflow to `codex`, `claude`, or another non-Copilot engine.
- Introducing `COPILOT_OFFLINE` mode or otherwise attempting to remove all GitHub connectivity from the run.

## Decisions

### 1. Use Copilot CLI BYOK provider settings instead of switching engines

The workflow should keep `engine.id: copilot` and route inference through LiteLLM by setting Copilot CLI BYOK provider configuration in `engine.env`. The intended shape is:

- `engine.id: copilot`
- `engine.model: llm-gateway/gpt-5.4`
- `engine.env.COPILOT_PROVIDER_TYPE: openai`
- `engine.env.COPILOT_PROVIDER_BASE_URL: https://elastic.litellm-prod.ai/v1`
- `engine.env.COPILOT_PROVIDER_API_KEY` sourced from a GitHub Actions secret-backed expression

Why:
- It preserves the workflow's existing Copilot-specific execution path and GH AW semantics.
- Copilot CLI documents BYOK provider configuration for OpenAI-compatible endpoints, which matches LiteLLM cleanly.
- It separates the stable repository-facing model alias (`llm-gateway/gpt-5.4`) from whatever provider routing LiteLLM performs behind the endpoint.

Alternatives considered:
- Switch the workflow engine to `codex` or another OpenAI-native engine: rejected because the requested migration explicitly keeps Copilot as the engine.
- Use a Copilot base-URL proxy override instead of BYOK provider settings: rejected because the BYOK provider path is the documented OpenAI-compatible configuration and makes provider auth explicit.

### 2. Treat the LiteLLM host as part of the AWF review-environment contract

The workflow should extend its existing AWF network policy to allow `elastic.litellm-prod.ai` alongside the current review-environment dependencies (`defaults`, `node`, and `go`).

Why:
- The current verification workflow already declares an explicit AWF allowlist.
- Routing inference through LiteLLM will fail unless the review job can reach that hostname.
- Encoding the host in the spec keeps the workflow observable and testable rather than burying the dependency in implementation detail.

Alternative considered:
- Rely on an implicit or broader network allowance: rejected because it weakens the current least-privilege posture and makes the routing dependency harder to audit.

### 3. Preserve GitHub-native workflow behavior and auth boundaries

This change should reroute only model inference. It should not alter the workflow's GitHub tools, safe outputs, or deterministic pre-activation steps, and it should not require offline mode. The workflow may continue to use GitHub authentication and GitHub-managed review/archive operations while sourcing LiteLLM provider auth separately.

Why:
- Review comments, review submission, label removal, and archive push are GitHub operations, not model-provider operations.
- Keeping those paths unchanged minimizes migration risk.
- Avoiding offline mode reduces the chance of accidentally breaking existing Copilot or GH AW behavior that still expects GitHub connectivity.

Alternative considered:
- Enable `COPILOT_OFFLINE` and try to fully isolate the run to LiteLLM: rejected because this workflow still depends on GitHub-native capabilities and the user request only requires routing the model path.

## Risks / Trade-offs

- [LiteLLM endpoint availability becomes part of verify-openspec reliability] -> Mitigation: keep the change limited to one workflow first and make the dependency explicit in both spec and workflow frontmatter.
- [A secret-backed provider auth contract adds one more operational prerequisite] -> Mitigation: require secret-backed configuration in the workflow contract and document the expected provider env handoff in the change artifacts.
- [The configured model alias may not match the final LiteLLM catalog forever] -> Mitigation: encode the requested model string in the workflow contract now; if the router naming changes later, capture that as a follow-up change.
- [Copilot-specific behavior could still differ slightly when using BYOK routing] -> Mitigation: keep the engine unchanged, use the documented BYOK provider configuration, and validate on a representative verify-openspec run before considering broader rollout.

## Migration Plan

1. Update `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` to set the new model and provider environment variables, and to allow the LiteLLM host in `network.allowed`.
2. Recompile `.github/workflows/openspec-verify-label.md` and `.github/workflows/openspec-verify-label.lock.yml`.
3. Sync the `ci-aw-openspec-verification` canonical spec with the approved routing and network requirements.
4. Validate the change artifacts and run or inspect a representative `verify-openspec` workflow execution that uses the LiteLLM-backed configuration.

## Open Questions

- None for the requested migration. Secret naming can be chosen during implementation as long as the provider API key is sourced from a GitHub Actions secret expression rather than a checked-in literal.
