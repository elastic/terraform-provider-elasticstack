## Context

The repository already runs two issue-driven agentic workflows:

- `change-factory` consumes a labeled issue and produces an OpenSpec proposal PR; it treats the issue title and body as authoritative and explicitly avoids any back-and-forth comment exploration loop.
- `code-factory` consumes a labeled issue (or a `workflow_dispatch` from another workflow) and produces an implementation PR.

Both share the deterministic intake plumbing in `.github/workflows-src/lib/factory-issue-shared.js`: trigger qualification, actor trust, duplicate-PR detection, label removal, and gate-reason finalization. They both wire up the `elastic-docs` MCP server, the litellm Anthropic proxy, and `status-comment: true`.

What the existing workflows can't do is *enrich* an underspecified issue. `change-factory`'s ambiguity escape hatch is "post one clarifying comment, emit `noop`, stop." That moves no work forward, captures no investigation, and forces the next iteration to start cold.

`research-factory` fills the gap. It is structurally a third sibling of change-factory and code-factory but with a fundamentally different output medium: instead of a pull request, its artifact is a gated section appended to the triggering issue's body. That medium choice is the keystone decision driving most of the rest of the design.

## Goals / Non-Goals

**Goals:**

- Make durable, visible progress on underspecified issues across one or more research loops.
- Produce a machine-and-human-readable research artifact that `change-factory` can adopt verbatim as the spine of an OpenSpec proposal.
- Compare at least two approaches per research run, grounded in Elastic documentation and the actual repository code.
- Re-runnable with full awareness of new comments and the prior research output, without bespoke state storage.
- Reuse the existing factory-intake plumbing so a future `repro-factory` is a small additional change rather than another fork.
- Constrain blast radius: read-only against the repo, time-boxed, single-session-per-issue concurrency.

**Non-Goals:**

- The issue classifier that auto-labels issues `change-factory` / `research-factory` / `repro-factory` (separate future change).
- The `repro-factory` workflow itself.
- A schema-level lint that mechanically validates the gated block; the prompt-driven contract is sufficient for v1.
- Persisting state outside the issue (no sidecar files, no hidden JSON, no separate state comment).
- Any modification to `code-factory`.
- Multi-issue coordination, cross-repository research, or research that mutates files in `main`.

## Decisions

### D1. The artifact lives in the issue body, not in a PR or sidecar

**Decision:** The agent's only durable output is a single gated section in the triggering issue's body, delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->`.

**Why:** `change-factory` already treats the issue title and body as the sole authoritative source. Putting research in the body — rather than in a comment, a PR, or a sidecar file — means `change-factory` consumes it through its existing input channel. No new plumbing in the consumer; the only consumer-side change is interpretive ("if a research block is present, prefer it").

**Alternatives considered:**

- *Research PR with a markdown report.* Rejected: change-factory would not see it; it would also force a branch + PR lifecycle the workflow doesn't otherwise need.
- *State file in the repo (`research/issue-NNN.md`).* Rejected: not visible to humans without leaving GitHub, and not ingested by change-factory.
- *Tracking comment that the agent edits each run.* Rejected: GitHub comments aren't part of the issue body, so change-factory wouldn't read them; also, deciding which of N comments is "the" research adds parsing fragility.

### D2. `update_issue` with `operation: replace`, agent reconstructs the body

**Decision:** The agent uses the `safe-outputs.update-issue` capability of the gh-aw framework with `operation: replace`. The agent receives the current issue body, strips any prior `<!-- implementation-research:start -->`…`<!-- implementation-research:end -->` block, regenerates the block, and emits the recomposed body. Outside-block content is preserved byte-for-byte.

**Why:** The framework's `replace-island` operation uses workflow-run-id-keyed markers that produce a *new* island per run rather than updating a stable one. That breaks the "make progress, don't accumulate" requirement. `replace` plus agent-side reconstruction gives us human-readable, stable markers under our own naming control, with the tradeoff that we trust the agent (which already has the full body in context) to preserve outside-block content.

**Alternatives considered:**

- *`replace-island` with framework markers.* Rejected: per-run markers, ugly auto-generated identifiers, no control over human-readable section heading.
- *`append` with deduplication step.* Rejected: would require a separate Action step to rewrite the body, and double-execution race conditions get nasty.
- *Append-only with a marker scan in a second job.* Same problems as above with more moving parts.

### D3. Stateless re-derivation; no persistent state outside the issue

**Decision:** Each run reads the entire current issue state — body (including the prior block) and all human-authored comments in chronological order — and synthesizes a fresh block. No hidden JSON metadata, no sidecar comment, no run-to-run handoff beyond what is in the issue itself.

**Why:** State files drift. JSON-in-HTML-comment blobs are ugly and ratchet themselves into prompt complexity. The set "original body + comments + current block" is sufficient: answered questions disappear naturally next round, new ones emerge naturally, and the human can audit the conversation by reading the issue.

**Alternatives considered:**

- *Hidden JSON state (`<!--research-state:{…}-->`).* Rejected: parsing fragility, no human readability, drift if humans edit.
- *Separate "research-state" comment the agent rewrites.* Rejected: comment proliferation, race with status comment, unclear ownership.
- *Dedicated state branch.* Rejected: massive overhead for the value.

### D4. Free will inside the block; edits are read as input, not preserved verbatim

**Decision:** The block contract states explicitly: edits inside the block are **read as input on the next run** but are **not preserved verbatim**. To influence the next run durably, users should comment on the issue or edit content outside the block. Users may still edit inside the block; the agent will treat their edits as one more piece of evidence.

**Why:** We cannot mechanically prevent humans from editing the block, and we shouldn't try. Three escape hatches are valid: (a) edit inside the block, (b) post a comment, (c) edit outside the block. Promising verbatim preservation of (a) requires the agent to detect human edits, which is fragile. Treating the prior block as draft input — not authoritative agent output — gives us a single, simple rule the prompt can express cleanly.

**Alternatives considered:**

- *Strict ownership: agent overwrites without reading prior block.* Rejected: throws away what may be the best signal of how to revise.
- *Detect-and-preserve human edits.* Rejected: fragile, brittle prompt logic, unclear what "human edit" even means after multiple loops.

### D5. Structural mandate of ≥2 approaches via output schema

**Decision:** The block schema requires:

- A `### Approaches considered` H3 with two or more `#### ` H4 children (one per approach).
- A `### Recommendation` H3 naming one approach as the chosen spine.
- A `### Open questions` H3 (possibly empty bullet list).
- A `### Problem framing` H3 above approaches.
- A `### References` H3 below.
- A provenance header (timestamp, run link, edit notice).

The prompt enforces these structurally. v1 has no automated linter; we may add one later if drift becomes a problem.

**Why:** Embedding "compare ≥2 approaches" in the *output structure* is more robust than a prose instruction. A future linter can mechanically reject malformed blocks. `change-factory`'s consumer prompt can also rely on these section names being stable.

**Alternatives considered:**

- *Free-form prose research.* Rejected: change-factory then has to guess where the recommendation lives.
- *Add a CI lint step in v1.* Deferred: not worth the implementation cost yet; the prompt + the human review already catch most drift.

### D6. 25-minute self-budget, 35-minute hard kill

**Decision:** The agent is told its budget is 25 minutes and that it should stop researching at minute 22 to leave 3 minutes for emitting `update_issue`. The job's `timeout-minutes` is 35, giving a 10-minute buffer for setup (checkout, `npm ci`), MCP latency, and safe-output post-processing.

**Why:** A single `timeout-minutes` is a guillotine — if the agent overshoots even slightly, the run dies and nothing is written. Two layers (prompt budget + hard kill) means worst-case is "agent emits a partial-but-valid block at minute 24" rather than "nothing is written." The prompt also instructs: *prefer a partial-but-valid block with explicit unanswered questions over `noop`*.

**Alternatives considered:**

- *Hard `timeout-minutes: 25`.* Rejected: too tight. Setup alone can eat several minutes.
- *No prompt-side budget.* Rejected: agents are bad at managing wall-clock without explicit instruction.

### D7. Per-issue concurrency, queue rather than cancel

**Decision:**

```yaml
concurrency:
  group: research-factory-issue-${{ github.event.issue.number || inputs.issue_number }}
  cancel-in-progress: false
```

**Why:** A research session in flight has invested compute and is about to write a useful block; killing it loses that work and may leave the issue half-updated. With `cancel-in-progress: false`, GitHub queues new triggers (collapsing superseded ones). The queued run picks up the freshly-updated state, including whatever the first run wrote, so it is strictly better-informed.

**Alternatives considered:**

- *`cancel-in-progress: true`.* Rejected: discards in-flight work, racy on `update_issue` writes.
- *No concurrency control.* Rejected: two parallel runs racing on the same `update_issue` payload would clobber each other.

### D8. Triggers: label + workflow_dispatch (mirroring code-factory)

**Decision:** Subscribe to `issues.opened`, `issues.labeled`, and `workflow_dispatch` with an `issue_number` input (and optional `source_workflow` for traceability). Apply the existing intake gate pattern (`determine_intake_mode` → `qualify_trigger`/`validate_dispatch_inputs` → `check_actor_trust` (issue-event only) → `fetch_live_issue` (dispatch only) → `normalize_context`).

**Why:** Reuses the proven shape from `code-factory`. The dispatch path is what the future issue classifier will use to chain workflows. Issue-event trust check applies only to the label-applied path; dispatch bypasses trust because it is internally authored. Comment fetching is a new step layered on top of the existing pattern.

**Alternatives considered:**

- *Label-only.* Rejected: blocks the future classifier integration.
- *`repository_dispatch` instead of `workflow_dispatch`.* Rejected: cross-repo permissions get more complex; `workflow_dispatch` is the consistent convention here.

### D9. Comment history as agent input

**Decision:** A new pre-activation step fetches all comments on the triggering issue, filters to human-authored ones (excluding `github-actions[bot]`, `dependabot[bot]`, and the workflow's own status comments), and exposes them as a normalized output that flows into the agent prompt alongside the issue body. Implemented as a shared helper in `lib/factory-issue-shared.js` so a future `repro-factory` can reuse it.

**Why:** The agent needs the conversation, not just the body. Pagination matters (issues with long discussions). Bot/status-comment filtering matters (we don't want the agent reading its own status comments as user signal). Centralizing this in the shared lib avoids duplicated implementations across factories.

**Alternatives considered:**

- *Let the agent fetch comments via the GitHub MCP tool.* Rejected: pre-activation deterministic capture is more auditable, and the agent shouldn't burn its 25-minute budget on plumbing.
- *Pass only the most recent N comments.* Rejected: arbitrary cutoff; full history is bounded by the GitHub issue size limit anyway.

### D10. No research safe-output PR; only `update-issue` and the framework status comment

**Decision:** `safe-outputs:` enables `update-issue` (with `body: enabled`) and `noop`. Status updates are conveyed via the framework's `status-comment: true`. No `add-comment` is configured; the agent cannot post free-form comments.

**Why:** Limiting outputs limits blast radius. The status comment already shows when the run started/completed and links to the run; the issue-body diff shows what changed. A separate `add-comment` would add notification spam without surfacing any signal that isn't already in the body diff.

**Alternatives considered:**

- *Enable `add-comment` for a "what changed this run" summary.* Rejected for v1: notification spam, duplication with body diff. Reconsider if maintainers find scanning the body diff insufficient.
- *Custom status-comment content.* Investigated; the gh-aw framework owns the activation/completion comment, and we accept its default content.

### D11. change-factory becomes research-block-aware, fallback unchanged

**Decision:** Modify the change-factory prompt to: (a) detect the `<!-- implementation-research:start --> … <!-- implementation-research:end -->` block in the issue body; (b) when present, treat the block as the **exclusive** authoritative scope (the `Recommendation` drives the proposal spine, `Open questions` go into `design.md`, `Approaches considered` informs context but is not re-explored); (c) when absent, retain today's "title and body authoritative" behavior unchanged.

**Why:** Issues that already arrived well-specified shouldn't be forced through research. The fallback preserves the current happy path; the awareness adds a strictly-better path when a block is present. The "exclusive when present" framing avoids ambiguity about whether the agent should also reason about content outside the block — outside-block content (the original problem statement, any subsequent human edits) is read for context, but the recommendation and open questions inside the block win.

**Alternatives considered:**

- *Always merge in-block + out-of-block reasoning.* Rejected: invites change-factory to override the recommendation, defeating the purpose of the research pass.
- *Make change-factory require a research block for all issues.* Rejected: forces unnecessary friction on simple, well-specified issues.

## Risks / Trade-offs

- **Risk:** Agent corrupts the original issue body during `replace`. → **Mitigation:** prompt explicitly instructs preservation of all content outside the markers byte-for-byte; the original body is also captured into a normalized output during pre-activation, and the agent prompt echoes it as the authoritative pre-block content.
- **Risk:** Agent accumulates blocks over runs (e.g., agent forgets to strip the prior block). → **Mitigation:** prompt is explicit on "exactly one block, regenerated each run"; we also trust D2's `replace` semantics — there is no append fallback that could double-write.
- **Risk:** Two triggers race and the second produces a stale block. → **Mitigation:** D7's per-issue concurrency with `cancel-in-progress: false` queues the second run, which then re-reads current state including whatever the first run wrote.
- **Risk:** Time-out kills the run before any block is written. → **Mitigation:** D6's two-layer budget; agent is instructed to write a partial block at minute 22 rather than time out; `noop` is explicitly less preferred than a partial block.
- **Risk:** A user edits the block expecting it to persist, then is frustrated when the next run rewrites it. → **Mitigation:** an explicit notice block at the top of every research section explains the social contract; we also accept this is a known UX cost of D4.
- **Risk:** Issues with very long comment threads exceed the agent context window. → **Mitigation:** comment fetching paginates with a sensible cap (e.g. last 200 human comments); if more exist, the prompt notes truncation and asks the agent to focus on the most recent threads. Documented as a v1 limitation.
- **Risk:** Research-factory and change-factory labels both applied. → **Mitigation:** research-factory removes only its own label; change-factory remains independently triggerable. The two workflows can run in sequence or independently. Promotion from research to change is a human (or future classifier) action.
- **Risk:** A user asks research-factory to write code by accident. → **Mitigation:** the prompt and the spec (`Workflow remains research-only`) explicitly forbid implementation; safe-outputs do not include `create-pull-request` or any code-writing tool; hard isolation rather than soft prompt rules.
- **Trade-off:** We trust the agent to do byte-correct body reconstruction. → Accepted because the agent has the full pre-block body in context and the prompt is explicit; failure modes are visible in the issue diff, recoverable with a single re-run.
- **Trade-off:** No mechanical schema validation of the block in v1. → Accepted because the prompt-driven contract has worked in change-factory and code-factory; if drift becomes a problem, a small markdown linter is a follow-up change.
