## Why

When `research-factory` produces two implementation approaches, there is currently no structured way for a human to select one before triggering `change-factory`. The only path is an undocumented edit to the research comment, and even then `change-factory` follows its own recommendation logic. A `/change-factory <freeform direction>` slash command gives maintainers a direct, discoverable mechanism to pass instructions to the proposal agent as the final word on design decisions — approach selection, scope adjustments, or any other override.

## What Changes

- Add a `slash_command: change-factory` trigger to the `change-factory-issue` workflow (events: `issue_comment` only), alongside the existing `issues: [labeled]` trigger.
- Extend `factoryQualifyTriggerEvent` in `factory-issue-shared.js` to accept `issue_comment` as an eligible event name, making the behaviour available to both `change-factory` and `code-factory` without per-workflow special-casing.
- Add a `capture_command_text` pre-activation step that extracts the text following `/change-factory` from the triggering comment body and exposes it as a `human_direction` output.
- Add a `notify_duplicate_blocked` pre-activation step that posts an explanatory issue comment when the duplicate-PR gate fires, telling the human to close the existing PR before retrying.
- Add a `human_direction` section to the `change-factory` agent prompt that, when non-empty, instructs the agent to treat the human's text as the final say on all design decisions, overriding the research comment's `### Recommendation`.

## Capabilities

### New Capabilities

- `ci-change-factory-slash-command`: Slash command trigger for `change-factory` issue intake — freeform human direction passed as an authoritative override to the proposal agent.

### Modified Capabilities

- `ci-change-factory-issue-intake`: Adds the slash-command trigger path, `human_direction` agent input, duplicate-blocked comment behaviour, and the `issue_comment` eligibility extension to the shared qualify-trigger helper.

## Impact

- `.github/workflows-src/lib/factory-issue-shared.js` — one new branch in `factoryQualifyTriggerEvent` accepting `issue_comment`; corresponding test case added in `factory-issue-shared.test.mjs`.
- `.github/workflows-src/change-factory-issue/workflow.md.tmpl` — new trigger, two new pre-activation steps, new pre-activation output, updated agent prompt section.
- `.github/workflows-src/change-factory-issue/scripts/` — two new inline scripts (`capture_command_text.inline.js`, `notify_duplicate_blocked.inline.js`).
- Generated artifacts re-compiled: `.github/workflows/change-factory-issue.md`, `.github/workflows/change-factory-issue.lock.yml`.
- No provider Go code, no acceptance tests, no Elastic Stack changes.
