## Context

`ci-changelog-generation` currently performs changelog synthesis late, from merged pull-request history, inside a GH AW workflow that gathers PR evidence and asks an agent to draft the release-note text. That architecture has proven fragile for this repo's needs: the strongest signal is often the PR title, customer-impact filtering is inferred too loosely from paths and labels, and the final bullets drift away from the tone and structure already established in `CHANGELOG.md`.

The requested pivot splits the problem into two workflows with different responsibilities:

1. A **PR-time agentic workflow** runs after `Build/Lint/Test` completes and ensures the pull request body contains a valid `## Changelog` section unless the PR is explicitly exempt with `no-changelog`.
2. A **scheduled/release deterministic workflow** reads merged PR bodies, parses the structured changelog contract, and rebuilds the target `CHANGELOG.md` section without asking an agent to summarize merged history.

This is cross-cutting because it affects workflow authoring format, trigger model, PR metadata handling, changelog parsing/validation helpers, and the canonical OpenSpec contract for CI/release automation.

## Goals / Non-Goals

**Goals:**

- Move changelog authorship to PR time so authors/reviewers provide the user-facing summary while the change context is fresh.
- Keep the normal changelog entry structured and easy to validate: `Customer impact` plus `Summary`.
- Allow an optional `### Breaking changes` subsection in PR bodies to remain free-form markdown, including paragraphs, lists, and fenced code blocks, while still being extractable deterministically by boundary.
- Trigger the PR-time workflow from `workflow_run` on `Build/Lint/Test` completion so authors have an opportunity to fill in the PR body before the workflow intervenes.
- Keep the release assembly workflow authoritative and deterministic: full-section regeneration from an authoritative merge range, no LLM summary of merged PRs, and only light normalization of output.
- Replace GH AW safe-output PR management in release assembly with explicit GitHub Actions logic for branch reuse, PR lookup, PR creation, and PR body refresh.

**Non-Goals:**

- Changing the existing `Build/Lint/Test` workflow's behavior beyond using its completion as the trigger source for the PR-time workflow.
- Replacing authoritative rebuilds with append-only changelog mutation on every merge.
- Parsing the internal semantics of breaking-change prose or rewriting breaking-change markdown into a more structured schema.
- Performing deep diff analysis during scheduled/release assembly.

## Decisions

1. **Split the problem into PR-time authoring and release-time assembly**  
   **Decision:** Add a new PR-time workflow capability for changelog authoring, and convert `ci-changelog-generation` into a deterministic assembly workflow.  
   **Rationale:** This keeps the agent where interpretation is useful and limits deterministic release assembly to parsing known structure.  
   **Alternative considered:** Keep one workflow and enrich the merged-PR evidence manifest with PR bodies/examples. Rejected because it still centralizes interpretation too late and keeps the release-time workflow responsible for both authorship and assembly.

2. **Use `workflow_run` on `Build/Lint/Test` completion for the PR-time workflow**  
   **Decision:** Trigger the PR-time workflow from `workflow_run` for the `Build/Lint/Test` workflow name and use deterministic steps to resolve the PR from the source run metadata.  
   **Rationale:** This gives humans time to add the changelog themselves, and it aligns with the intended required-check flow. It also allows the follow-up workflow to have the permissions needed to update the PR body.  
   **Alternative considered:** Trigger directly on `pull_request` events. Rejected because it races earlier in the authoring flow and is more likely to overwrite or duplicate human-authored changelog text before CI completes.

3. **Treat `workflow_run` as privileged metadata automation only**  
   **Decision:** The PR-time workflow SHALL avoid checking out or executing untrusted PR code. Deterministic gating, validation, and agent prompting will operate only on PR metadata (title, description, labels, body) and repository-authored examples/instructions.  
   **Rationale:** `workflow_run` carries elevated privileges; keeping it metadata-only avoids the standard pwn-request class of problems while still allowing PR body mutation.  
   **Alternative considered:** Allow checkout after deterministic gating for same-repository PRs. Rejected because checkout is unnecessary for this task and increases security risk without clear benefit.

4. **Keep the normal changelog summary structured but allow free-form breaking changes**  
   **Decision:** The PR-body contract will require structured `Customer impact` and `Summary` fields, but `### Breaking changes` will be optional free-form markdown preserved as a delimited block.  
   **Rationale:** Normal release notes need consistency and simple parsing, while breaking changes often require migration prose and code blocks that do not fit a tight schema.  
   **Alternative considered:** Make breaking changes a structured bullet list. Rejected because it would force unnatural formatting and lose important migration detail.

5. **The PR agent should fill a fixed template rather than free-write**  
   **Decision:** Constrain the PR-time agent to producing the `## Changelog` template fields instead of writing arbitrary release-note prose.  
   **Rationale:** The workflow's job is to populate a deterministic contract, not to create polished final changelog output. This also makes validator design simpler.  
   **Alternative considered:** Let the agent draft a loose changelog section and normalize it later. Rejected because it reintroduces the title-dump and style-drift problem.

6. **Scheduled/release assembly remains full-range and deterministic**  
   **Decision:** Preserve the existing authoritative-range rebuild model for both `## [Unreleased]` and release sections, but replace the agentic summary step with deterministic parsing of merged PR bodies and labels.  
   **Rationale:** Full rebuilds are easier to reason about and self-heal if a previous run missed or misrendered something.  
   **Alternative considered:** Replace the rebuild with merge-triggered append-only logic. Rejected as the primary model because it is more prone to drift, race conditions, and correction difficulties after merge.

7. **Simple normalization only during release assembly**  
   **Decision:** Assembly may normalize bullets, citations, whitespace, and placement of breaking-change blocks, but it SHALL NOT semantically rewrite the author-provided content.  
   **Rationale:** The PR contract is now the source of truth; deterministic assembly should render it consistently, not reinterpret it.  
   **Alternative considered:** Add a second agent or complex deterministic heuristics to polish output. Rejected because that recreates the same source-of-truth ambiguity the pivot is meant to remove.

8. **Use normal GitHub Actions PR management for deterministic assembly**  
   **Decision:** Scheduled/manual runs will update the `generated-changelog` branch, look up or create the singleton PR, and update its body if it already exists. Release runs will update only the triggering `prep-release-*` branch and use the PR number from event metadata when refreshing PR metadata.  
   **Rationale:** Once release assembly is deterministic, GH AW safe outputs are no longer appropriate or necessary.  
   **Alternative considered:** Keep the current GH AW wrapper just for PR creation/update. Rejected because it couples deterministic release assembly to an agentic runtime that is otherwise being removed.

## Risks / Trade-offs

- **[Risk] Authors may leave low-quality or stale changelog sections in PR bodies** → **Mitigation:** make the PR workflow a required check, validate structure deterministically, and let the agent only fill gaps rather than replace valid author input.
- **[Risk] `workflow_run` may be used unsafely** → **Mitigation:** keep the PR-time workflow metadata-only, with no checkout or execution of untrusted branch code.
- **[Risk] The free-form `### Breaking changes` block could be malformed or hard to place cleanly** → **Mitigation:** validate only that the block is present and non-empty when declared, and preserve it by heading boundaries rather than attempting semantic parsing.
- **[Risk] Scheduled/release assembly may have to handle mixed historical PRs that predate the new contract** → **Mitigation:** keep authoritative rebuilds, add explicit handling for missing contract cases, and use `no-changelog` plus deterministic fallbacks where needed during rollout.
- **[Risk] Moving PR update logic from GH AW to normal Actions may create branch/PR drift bugs** → **Mitigation:** centralize lookup/create/update logic in repository-authored scripts or helper steps, and cover singleton `generated-changelog` reuse plus release-PR refresh behavior with tests.

## Migration Plan

1. Define the new PR-time workflow capability and update `ci-changelog-generation` spec requirements to describe deterministic assembly rather than merged-history summarization.
2. Introduce repository-authored parser/validator helpers for the PR-body changelog contract, including optional breaking-change block extraction.
3. Implement the PR-time agentic workflow triggered from `workflow_run` on `Build/Lint/Test`, with deterministic PR resolution, skip logic, validation, and PR body update behavior.
4. Refactor the changelog-generation workflow source and compiled outputs to remove the merged-history agent phase, replace GH AW PR update safe outputs, and rebuild changelog sections from parsed merged PR metadata.
5. Validate the resulting workflows and OpenSpec artifacts, then make the PR-time workflow a required check after maintainers confirm its ergonomics.

## Open Questions

- Whether the PR-time workflow should treat an empty `## Changelog` section as “missing” and always invoke the agent, or whether some partial-but-invalid shapes should hard-fail instead of being auto-filled.
- Whether rollout needs a temporary deterministic fallback for merged PRs that predate the new PR-body contract but still fall inside the authoritative release range.
- Whether a merge-triggered convenience rerun is worth adding in the first implementation, or whether scheduled/manual plus release-preparation triggers are sufficient until the deterministic assembly stabilizes.
