## 1. Shared lib — accept `issue_comment` in `factoryQualifyTriggerEvent`

- [ ] 1.1 In `factory-issue-shared.js`, add a branch at the top of `factoryQualifyTriggerEvent` that returns `event_eligible: true` when `eventName === 'issue_comment'`
- [ ] 1.2 Add a test case to `factory-issue-shared.test.mjs` asserting that `factoryQualifyTriggerEvent` returns `event_eligible: true` for `eventName === 'issue_comment'`
- [ ] 1.3 Confirm the existing `pull_request` ineligibility test still passes (gh-aw routes PR conversation comments under `pull_request_comment` rather than `issue_comment`, so no additional payload guard is required in the shared helper)

## 2. New pre-activation scripts

- [ ] 2.1 Create `scripts/capture_command_text.inline.js` — reads `context.payload.comment?.body`, strips the leading `/change-factory` token and surrounding whitespace, outputs `human_direction` (empty string when event is not `issue_comment`)
- [ ] 2.2 Create `scripts/notify_duplicate_blocked.inline.js` — posts one comment on `context.payload.issue.number` with the existing PR URL and retry instructions using `github.rest.issues.createComment`; script is a no-op when `process.env.DUPLICATE_PR_URL` is empty

## 3. `workflow.md.tmpl` — trigger, pre-activation steps, and outputs

- [ ] 3.1 Add `slash_command: { name: change-factory, events: [issue_comment] }` to the `on:` block alongside the existing `issues: [labeled]` trigger
- [ ] 3.2 Add `capture_command_text` step to `on.steps:` (after `qualify_trigger`, before `check_actor_trust`); include `x-script-include: scripts/capture_command_text.inline.js`; expose `human_direction` output
- [ ] 3.3 Add `notify_duplicate_blocked` step to `on.steps:` with `if:` condition `steps.qualify_trigger.outputs.event_eligible == 'true' && steps.check_actor_trust.outputs.actor_trusted == 'true' && steps.check_duplicate_pr.outputs.duplicate_pr_found == 'true'`; include `x-script-include: scripts/notify_duplicate_blocked.inline.js`; pass `DUPLICATE_PR_URL` and `ISSUE_NUMBER` env vars
- [ ] 3.4 Add `human_direction: ${{ steps.capture_command_text.outputs.human_direction }}` to `jobs.pre-activation.outputs`
- [ ] 3.5 Add `## Human direction` section to the agent prompt markdown, conditional on `${{ needs.pre_activation.outputs.human_direction }}` being non-empty, framing the text as the final say on all design decisions and stating it overrides the research comment's `### Recommendation`

## 4. Compile and verify

- [ ] 4.1 Run `make workflow-generate` (or equivalent) to regenerate `.github/workflows/change-factory-issue.md` from the template
- [ ] 4.2 Run `gh aw compile .github/workflows/change-factory-issue.md` to regenerate the lock file; confirm zero errors and zero warnings
- [ ] 4.3 Run `make check-workflows` and confirm it passes
- [ ] 4.4 Confirm the compiled lock file contains an `update_research_comment`-equivalent job (or the slash_command job block) with `issues: write` in its permissions
- [ ] 4.5 Run the lib unit tests (`node --test .github/workflows-src/lib/*.test.mjs`) and confirm all pass
