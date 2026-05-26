## Context

Every factory workflow (research, code, change, reproducer) performs actor trust gating through a
dedicated `check_actor_trust` pre-activation step. That step calls
`github.rest.repos.getCollaboratorPermissionLevel` — a live GitHub API call that threads
`actor_trusted == 'true'` guards through all subsequent conditional steps.

gh-aw already enforces the same trust policy by default. For `issues` triggers, the gh-aw compiler
injects a `check_membership` step with `GH_AW_REQUIRED_ROLES: "admin,maintainer,write"`. This
checks actual repository permission levels via the GitHub API — the same semantic as the old JS
check. The original JS implementation was therefore redundant, layered on top of what gh-aw already
does.

The issue explicitly asks to kill the hand-rolled check and rely on the native version.

## Goals / Non-Goals

**Goals:**
- Remove the redundant hand-rolled JS actor-trust check from all four factory workflows.
- Rely on gh-aw's built-in `check_membership` / `roles` default for human actor trust gating.
- Explicitly allow `github-actions[bot]` on `code-factory` via `on.bots` so workflow dispatch from
  `semantic-function-refactor` (and similar producers) is not blocked.
- Delete the `check-actor-trust.js` runner and the shared library functions that back it.
- Prune the corresponding tests to keep the test suite aligned with the live code.
- Regenerate all four factory workflow lock files.

**Non-Goals:**
- Changing the trust policy itself (this is a mechanical replacement, not a policy revision).
- Migrating non-factory workflows that may have separate actor-trust patterns.
- Updating the `workflows.yml` test harness beyond removing now-deleted function tests.

## Decisions

### 1. Rely on gh-aw implicit `roles` default (`admin,maintainer,write`)

The gh-aw compiler already injects `check_membership` with `GH_AW_REQUIRED_ROLES:
"admin,maintainer,write"` for `issues` triggers. This is functionally equivalent to the old JS
policy: it allows anyone with write-level repository access and blocks anyone without it.

No explicit `roles` frontmatter is declared because the default is already correct.

### 2. Do NOT use `skip-author-associations`

`skip-author-associations` gates on the `author_association` field in the event payload. GitHub's
documentation warns that this field can be unreliable in webhook payloads. Additionally, it's
redundant with `check_membership` which checks actual repository permissions. Using both layers
adds complexity without improving security.

### 3. Add `on.bots: [github-actions[bot]]` to `code-factory`

The `code-factory` workflow is explicitly triggered by `semantic-function-refactor` (and
potentially other producer workflows) via `workflow_dispatch`. Those dispatches run as
`github-actions[bot]`, which the default `roles` check might block because bots don't have a
conventional repository permission level.

Adding `on.bots: [github-actions[bot]]` makes bot dispatch handling visible in the source and
ensures producer-to-consumer workflow chains continue to function.

### 4. Hardcode `actor_trusted=true` in `normalize_context` for issue-event paths

After removal of `check_actor_trust`, the `normalize_context` step in each workflow can emit
`actor_trusted=true` unconditionally for the issue-event path, because the role gate has already
filtered untrusted actors before the job runs. The dispatch-intake path already hardcodes
`actor_trusted=true` today; this change makes both paths consistent.

The `finalize_gate` step reads `ACTOR_TRUSTED` from `normalize_context`; it will continue to work
correctly because it will always receive `actor_trusted=true` for events that pass the role gate.

### 5. Remove `check-actor-trust.js`, `factoryCheckActorTrust`, `factoryActorTrustWhenSenderMissing`

After the role gate handles actor filtering, the JS runner and its library functions are dead
code. Remove them entirely to avoid the appearance that the JS check is still active.

### 6. Prune only actor-trust tests; leave all other test coverage intact

Tests for `factoryQualifyTriggerEvent`, `factoryCheckDuplicatePR`, `factoryComputeGateReason`, and
other non-trust logic in `factory-issue-shared.test.mjs` remain untouched. Only the
`factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` test cases are removed from
`factory-issue-shared.test.mjs`, and the corresponding actor-trust test cases are removed from
`code-factory-issue.test.mjs` and `change-factory-issue.test.mjs`.

### 7. Regenerate all four lock files

After source edits, run `gh aw compile` for each workflow to produce updated lock files that match
the new source. The lock files are committed alongside the source changes in the same pull request.

## Risks / Trade-offs

- [No semantic shift] The effective trust boundary is unchanged. The gh-aw `check_membership`
  already enforces write/maintain/admin roles — exactly what the old JS check did. The only
  difference is that the check is now compiled by gh-aw rather than hand-rolled in JavaScript.
- [Four workflow files + JS + tests = broad diff] The changes are mechanical and low-risk, but the
  breadth means all four lock files must be regenerated; omitting any one would leave a stale lock.

## Migration Plan

1. Remove `skip-author-associations` from all four `.md` files; remove `check_actor_trust` step and
   downstream `actor_trusted` conditions.
2. Add `bots: [github-actions[bot]]` to `code-factory-issue.md`.
3. Simplify `normalize_context` in each workflow to hardcode `actor_trusted=true` for issue-event
   paths.
4. Remove `factoryCheckActorTrust`, `factoryActorTrustWhenSenderMissing`, and their exports from
   `factory-issue-shared.js`.
5. Delete `check-actor-trust.js`.
6. Prune actor-trust tests from `factory-issue-shared.test.mjs`,
   `code-factory-issue.test.mjs`, and `change-factory-issue.test.mjs`.
7. Run `gh aw compile` for all four workflows to regenerate lock files.
8. Verify CI passes with `make check` (or equivalent) and commit.

## Open Questions

- After removing the `check_actor_trust` step, `finalize_gate`'s `computeGateReason` receives
  `actor_trusted=true` unconditionally for issue-event paths. Confirm the resulting `gate_reason`
  text used by downstream workflows is still correct.
