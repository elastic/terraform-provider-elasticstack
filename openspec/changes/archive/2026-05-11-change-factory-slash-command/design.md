## Context

The `change-factory-issue` workflow currently activates only via `issues: [labeled]`. It delegates proposal authoring to an agent whose design decisions are guided by the implementation-research comment's `### Recommendation`. There is no mechanism for a human to override that recommendation at trigger time — they must either rely on the undocumented edit-the-research-comment path or post a comment and hope the agent interprets it as an "explicit contradiction".

The `slash_command:` trigger in gh-aw fires on `issue_comment` events and automatically provides the sanitised comment text (everything after the command name) as `steps.sanitized.outputs.text` in the activation context. Combining this with the existing `issues: [labeled]` trigger is explicitly supported by gh-aw when the issues trigger is label-only.

The shared `factoryQualifyTriggerEvent` helper currently hard-gates on `eventName !== 'issues'`, returning ineligible for any other event. This is the only barrier to supporting slash-command entry for both `change-factory` and `code-factory` without per-workflow special-casing.

## Goals / Non-Goals

**Goals:**
- Allow maintainers to post `/change-factory <freeform text>` on an issue comment to trigger `change-factory` with their text as an authoritative design directive.
- Make the human direction available to the proposal agent as the final word, overriding the research recommendation.
- Post an explanatory comment when the duplicate-PR gate fires (applies to both label and slash command triggers).
- Fix the shared `factoryQualifyTriggerEvent` helper at the right level so `code-factory` benefits automatically when it adds a slash command trigger in future.

**Non-Goals:**
- Structured approach selection syntax (e.g. `approach:id`) — freeform text is sufficient.
- A slash command for `code-factory` — that is a separate change.
- Behaviour changes to `research-factory`.

## Decisions

### D1: Accept `issue_comment` in `factoryQualifyTriggerEvent` (shared lib, not per-workflow)

**Decision**: Add a branch at the top of `factoryQualifyTriggerEvent` in `factory-issue-shared.js` that returns `event_eligible: true` when `eventName === 'issue_comment'`, before the existing `eventName !== 'issues'` guard. No payload-level `issue.pull_request` guard is needed.

**Rationale**: The qualification logic for slash-command triggers is event-level, not factory-specific — any factory workflow that adds a `slash_command:` trigger will want the same pass-through. Handling it once in the shared lib is correct and avoids drift. The existing `pull_request` rejection test remains valid.

Critically, gh-aw's `slash_command:` trigger uses `events: [issue_comment]` as a routing filter that corresponds to gh-aw's own `issue_comment` event name, which is **distinct** from `pull_request_comment`. Unlike the raw GitHub Actions `issue_comment` webhook (which fires for both issue and PR comments), gh-aw routes pull request conversation comments under its `pull_request_comment` event. A workflow declaring `events: [issue_comment]` therefore never receives PR-comment payloads at all — no `github.event.issue.pull_request` guard is required in application code. The `factoryQualifyTriggerEvent` unconditional pass-through for `issue_comment` is safe precisely because gh-aw's routing acts as the first filter.

**Alternatives considered**: Handle in each workflow's `qualify_trigger.inline.js` before calling the shared function. Rejected — duplicates logic across factories and obscures where the contract is defined. Add a `github.event.issue.pull_request` guard in `factoryQualifyTriggerEvent`. Rejected — unnecessary given gh-aw's routing model, and would introduce coupling to GitHub's raw payload shape that bypasses the framework abstraction.

### D2: Capture command text in a dedicated pre-activation step

**Decision**: Add a `capture_command_text` step in `on.steps:` that reads `context.payload.comment?.body`, strips the leading `/change-factory` token and surrounding whitespace, and writes the remainder to a `human_direction` output. When the event is not `issue_comment` (i.e. the label path), the step outputs an empty string.

**Rationale**: Keeps command-text extraction as a deterministic pre-activation step with clear output semantics. The agent receives it as a prompt variable, not raw payload data.

### D3: `human_direction` is free-form; the agent is instructed to treat it as final say

**Decision**: No parsing, no structured extraction. The full text after `/change-factory` is passed verbatim to the agent under a `## Human direction` prompt section that calls it "the final say on all design decisions" and says it overrides the research recommendation.

**Rationale**: Users will write things like "use approach B", "skip the plugin SDK approach, use the framework approach", or "ignore the research and just use X". Freeform with strong prompt framing is more resilient to natural language variation than any structured format.

### D4: Duplicate-blocked comment posted from pre-activation via `on.steps`

**Decision**: Add a `notify_duplicate_blocked` step in `on.steps:` that runs when `steps.check_duplicate_pr.outputs.duplicate_pr_found == 'true'` and the event was eligible and actor trusted. It posts a single comment on the issue using `actions/github-script` with the `GITHUB_TOKEN` (which already has `issues: write` from `on.permissions:`).

**Rationale**: The blocked comment is a deterministic side-effect, not agent output — it belongs in pre-activation. The `on.permissions: issues: write` is already present for label removal; no new permission grant needed.

**Comment text**: `⚠️ **change-factory skipped** — [#PR](url) is already open for this issue.\nClose or convert it to a draft, then retry.`

This applies to both the label and slash-command paths — the duplicate gate fires regardless of trigger source.

### D5: `remove_trigger_label` is a no-op when triggered by slash command

**Decision**: No code change. The existing `remove_trigger_label` script calls `github.rest.issues.removeLabel` and handles 404 as success. When triggered via slash command, the `change-factory` label is not on the issue, so the 404 path fires silently. Acceptable.

**Alternative considered**: Condition the step on `eventName == 'issues'`. This would be marginally cleaner but adds complexity with no observable difference.

## Risks / Trade-offs

- **Risk**: A slash command fires on a `change-factory`-labelled issue that already triggered a label run earlier in the same workflow run → two agent runs race. **Mitigation**: The duplicate-PR gate blocks the second run; if the first run completes fast enough, the second sees a PR and posts the blocked comment instead.
- **Risk**: Human direction text contains prompt injection. **Mitigation**: `steps.sanitized.outputs.text` is the gh-aw sanitised form of the comment body (gh-aw strips the command prefix and applies standard sanitisation). The text is injected into the prompt in a clearly-labelled section with explicit override semantics — no structural escaping is added beyond what gh-aw provides.
- **Risk**: `factoryQualifyTriggerEvent` change makes `code-factory` accept `issue_comment` events from unrelated workflows. **Mitigation**: `code-factory` has no `issue_comment` trigger today; the change is inert until a slash command trigger is added. Actor trust check still applies.

## Open Questions

None — design is complete.
