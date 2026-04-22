## Context

The `openspec-verify-label` workflow is already structured around GitHub Agentic Workflows with deterministic pre-activation gates and GitHub-native review and archive safe outputs. The implemented workflow now uses the `claude` engine and routes Anthropic-compatible model traffic through the Elastic LiteLLM gateway instead of relying on a GitHub-hosted default model path.

This change is intentionally narrow: it affects only the verification workflow and only the model-inference path. The repository wants the OpenSpec artifacts to reflect the implemented `claude` engine contract, the LiteLLM endpoint at `https://elastic.litellm-prod.ai`, and the workflow model selector `llm-gateway/gpt-5.4`.

## Goals / Non-Goals

**Goals:**
- Keep `openspec-verify-label` on the implemented `claude` engine.
- Route verification-model traffic through the Elastic LiteLLM proxy using the Claude engine's Anthropic-compatible environment settings.
- Keep the workflow model selection at `llm-gateway/gpt-5.4`.
- Make the required endpoint, secret-backed auth handoff, and AWF network allowlist explicit in the OpenSpec contract.
- Preserve existing GitHub review, archive, push, and label-cleanup behavior.

**Non-Goals:**
- Migrating other GH AW workflows, including `schema-coverage-rotation`, in the same change.
- Changing review disposition rules, archive gating, or the deterministic pull-request classification logic.
- Reworking the workflow around a different engine contract than the implemented `claude` path.
- Removing GitHub connectivity from the run or changing GitHub-native review/archive operations.

## Decisions

### 1. Express LiteLLM routing through the Claude engine contract

The workflow should keep the implemented `claude` engine configuration and route inference through LiteLLM using Anthropic-compatible environment variables. The intended shape is:

- `engine.id: claude`
- `engine.model: llm-gateway/gpt-5.4`
- `engine.env.ANTHROPIC_BASE_URL: https://elastic.litellm-prod.ai`
- `engine.env.ANTHROPIC_API_KEY` sourced from a GitHub Actions secret-backed expression

Why:
- It matches the authored workflow source and compiled workflow artifacts already in the repository.
- The Claude engine expects Anthropic-style configuration, which LiteLLM can proxy through the configured base URL.
- It keeps the repository-facing model alias (`llm-gateway/gpt-5.4`) stable while allowing backend routing to evolve behind the gateway.

Alternatives considered:
- Preserve an outdated engine design in the spec despite the implementation: rejected because it leaves the change internally inconsistent and unverifiable.
- Switch to a different engine contract: rejected because the implemented workflow has already standardized on `claude`.

### 2. Treat the LiteLLM host as part of the AWF review-environment contract

The workflow should extend its existing AWF network policy to allow `elastic.litellm-prod.ai` alongside the current review-environment dependencies (`defaults`, `node`, and `go`).

Why:
- The current verification workflow already declares an explicit AWF allowlist.
- Routing inference through LiteLLM will fail unless the review job can reach that hostname.
- Encoding the host in the spec keeps the workflow observable and testable rather than burying the dependency in implementation detail.

Alternative considered:
- Rely on an implicit or broader network allowance: rejected because it weakens the current least-privilege posture and makes the routing dependency harder to audit.

### 3. Preserve GitHub-native workflow behavior and auth boundaries

This change should reroute only model inference. It should not alter the workflow's GitHub tools, safe outputs, or deterministic pre-activation steps. The workflow may continue to use GitHub authentication and GitHub-managed review/archive operations while sourcing LiteLLM auth separately through the Claude engine environment.

Why:
- Review comments, review submission, label removal, and archive push are GitHub operations, not model-provider operations.
- Keeping those paths unchanged minimizes migration risk.
- Narrowing the change to engine configuration avoids unnecessary churn in the rest of the workflow contract.

Alternative considered:
- Expand the migration into a broader workflow redesign: rejected because the user request is limited to aligning the existing verify workflow with its implemented Claude routing.

## Risks / Trade-offs

- [LiteLLM endpoint availability becomes part of verify-openspec reliability] -> Mitigation: keep the change limited to one workflow first and make the dependency explicit in both spec and workflow frontmatter.
- [A secret-backed auth contract adds one more operational prerequisite] -> Mitigation: require secret-backed configuration in the workflow contract and document the expected `ANTHROPIC_API_KEY` handoff in the change artifacts.
- [The configured model alias may not match the final LiteLLM catalog forever] -> Mitigation: encode the current model string in the workflow contract now; if the router naming changes later, capture that as a follow-up change.
- [Claude engine behavior through LiteLLM could differ from direct Anthropic defaults] -> Mitigation: keep the contract explicit, limit the rollout to one workflow, and validate on representative `verify-openspec` runs.

## Migration Plan

1. Update `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` to keep the implemented `claude` engine configuration, document the Anthropic-compatible environment variables, and allow the LiteLLM host in `network.allowed`.
2. Recompile `.github/workflows/openspec-verify-label.md` and `.github/workflows/openspec-verify-label.lock.yml`.
3. Sync the `ci-aw-openspec-verification` canonical spec with the approved Claude-routing and network requirements.
4. Validate the change artifacts and run or inspect a representative `verify-openspec` workflow execution that uses the Claude-via-LiteLLM configuration.

## Open Questions

- None for the current implementation alignment. Secret naming can vary as long as `ANTHROPIC_API_KEY` is sourced from a GitHub Actions secret expression rather than a checked-in literal.
