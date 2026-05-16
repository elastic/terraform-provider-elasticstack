## Context

The repository has a `code-factory` issue-intake workflow that currently activates from `issues.opened` and `issues.labeled` events when an issue carries the `code-factory` label. Three other repository-authored analysis workflows (`semantic-function-refactor`, `schema-coverage-rotation`, and `flaky-test-catcher`) use safe outputs to create follow-up issues and currently attach the `code-factory` label so those issues can be implemented automatically.

That trigger strategy fails in practice because the issues are created by GitHub Actions safe-output processing using `GITHUB_TOKEN`, which does not trigger downstream issue-event workflows. The producer workflows therefore create actionable issues, but no implementation workflow run follows. We now know the producer workflows already emit `/tmp/gh-aw/temporary-id-map.json`, whose entries contain the real created issue numbers and repository slugs for every created issue. That artifact provides a deterministic handoff surface after safe-output processing has completed.

The repository also needs to preserve the manual maintainer path where adding `code-factory` to an existing issue should still trigger the implementation workflow. The design therefore needs two entry modes: manual issue-event intake and internal workflow dispatch intake. In both cases, `code-factory` must continue to implement exactly one issue per run and keep duplicate linked-PR suppression deterministic.

## Goals / Non-Goals

**Goals:**
- Preserve manual `code-factory` label-based issue intake for maintainers.
- Replace producer-side reliance on issue-label trigger side effects with explicit deterministic dispatch after issue creation.
- Support true fan-out so one producer run can dispatch `code-factory` once per created issue.
- Keep the issue itself as the canonical source of implementation scope by fetching live issue data in dispatch mode.
- Reuse the existing duplicate-PR guardrail so redispatches or partial retries do not create duplicate work.
- Remove `code-factory` from producer-created issue labels so label semantics return to “manual intake trigger” rather than overloaded metadata.

**Non-Goals:**
- Changing `code-factory` from a single-issue-per-run workflow into a batch worker.
- Introducing cross-repository dispatch; all dispatches remain within the current repository.
- Replacing the existing duplicate-PR linkage contract (`code-factory/issue-<n>` plus `Closes #<n>`).
- Reworking unrelated producer workflow issue-selection logic, prioritization logic, or issue-slot caps.

## Decisions

### 1. Use `workflow_dispatch` as the automated `code-factory` entrypoint
`code-factory` will add a `workflow_dispatch` trigger with typed inputs for at least `issue_number`, `issue_repo`, and optional provenance such as `source_workflow`. This keeps the workflow invocable from deterministic repository jobs after issue creation.

**Why this choice:**
- `workflow_call` is a poor fit for this handoff because the producer workflows can create up to three issues and need true fan-out. The gh-aw `call-workflow` output path does not naturally support “dispatch once per created issue” after safe-output processing.
- `workflow_dispatch` creates independent downstream runs, which is operationally clearer and matches the single-issue-per-run contract.

**Alternatives considered:**
- **`workflow_call` / safe-output call-workflow**: rejected because it does not model post-create N-way fan-out cleanly and depends on information the agent does not have during reasoning.
- **Keep relying on `code-factory` labels on created issues**: rejected because `GITHUB_TOKEN`-created issue events do not trigger the downstream workflow.

### 2. Dispatch from a deterministic post-safe_outputs job, not from agent-emitted safe outputs
Each producer workflow will add a deterministic job after `safe_outputs` that downloads or reads the safe-output artifacts, parses `temporary-id-map.json`, and dispatches one `code-factory` run per created issue.

**Why this choice:**
- The real issue numbers exist only after safe-output processing has completed.
- The temporary ID map is already structured as the authoritative mapping from temporary issue IDs to `{ repo, number }` and therefore supports deterministic fan-out without heuristic searches.
- Keeping dispatch logic outside the agent avoids coupling the agent prompt to post-processing mechanics and preserves least privilege for the reasoning phase.

**Alternatives considered:**
- **Agent emits dispatch outputs directly**: rejected because the agent does not know the final issue numbers yet.
- **Search GitHub after safe outputs using title/labels/time windows**: rejected because it is less deterministic than using the temporary ID map.

### 3. Remove `code-factory` from producer-created issue labels
The semantic refactor, schema coverage rotation, and flaky test catcher workflows will stop adding `code-factory` to the labels they apply to newly created issues.

**Why this choice:**
- It restores a clean meaning for the `code-factory` label: manual maintainer-triggered intake.
- It avoids a confusing partial semantic where some `code-factory`-related issues are label-triggered and others are dispatch-triggered.
- The new dispatch mechanism no longer needs the label for automation.

**Alternatives considered:**
- **Keep `code-factory` as a metadata label on producer-created issues**: rejected to avoid overloading the label with two incompatible meanings and two different lifecycle behaviors.

### 4. Normalize intake context inside `code-factory`
`code-factory` will resolve a normalized intake context during pre-activation, producing outputs such as issue number, issue title, issue body, intake mode, and gate reason. The implementation prompt and downstream deterministic steps will consume these normalized outputs instead of directly referencing `github.event.issue.*`.

**Why this choice:**
- `workflow_dispatch` runs do not have `github.event.issue` payloads.
- A normalized intake model lets the workflow preserve one implementation prompt and one downstream duplicate-PR contract across both entry modes.
- This keeps the workflow’s “issue is the sole source of truth” rule intact while still allowing multiple entry mechanisms.

**Alternatives considered:**
- **Branch the entire workflow into separate manual and dispatch implementations**: rejected because it would duplicate prompt and guardrail logic.
- **Trust issue title/body passed through dispatch inputs**: rejected because the live issue should remain authoritative, and dispatch inputs only need to identify which issue to load.

### 5. Split deterministic gating by entry mode while preserving duplicate-PR suppression
Manual issue-event runs will continue to use event qualification, actor-trust checks, duplicate-PR checks, and trigger-label removal. Dispatch-triggered runs will skip label qualification and actor-trust logic, validate dispatch inputs, fetch the live issue, and then apply the same duplicate-PR suppression used by manual runs.

**Why this choice:**
- The trust boundary is different for internal dispatch than for label-triggered issue events.
- Reusing duplicate-PR suppression ensures retries, partial failures, or repeated dispatches do not open duplicate implementation PRs.
- Keeping trigger-label removal only in manual mode matches the new semantics where dispatch-triggered producer issues no longer carry the `code-factory` label.

**Alternatives considered:**
- **Reusing actor-trust checks in dispatch mode**: rejected because `workflow_dispatch` is intentionally a repository-authored internal handoff, not a user-applied label event.

## Risks / Trade-offs

- **Dispatch job parses the wrong artifact shape** → Mitigation: treat `temporary-id-map.json` as the required contract, add focused tests for map parsing, and fail clearly when the file is missing or malformed.
- **Producer workflow reruns redispatch already-dispatched issues** → Mitigation: rely on the existing duplicate linked-PR guardrail in `code-factory`, and ensure dispatch mode still checks for canonical linked PRs before agent activation.
- **Manual and dispatch entry modes drift behaviorally** → Mitigation: normalize intake outputs and keep shared duplicate-PR logic and prompt contract in one workflow path.
- **Issue changes between creation and dispatch** → Mitigation: this is an acceptable trade-off because the live issue should be authoritative; dispatch mode intentionally fetches the current title/body.
- **Dispatch introduces more workflow runs to observe** → Mitigation: accept this as a feature of true fan-out; separate runs provide clearer per-issue traceability and retry behavior.
