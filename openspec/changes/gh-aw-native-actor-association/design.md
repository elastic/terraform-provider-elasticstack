## Context

Every factory workflow (research, code, change, reproducer) performs actor trust gating through a
dedicated `check_actor_trust` pre-activation step. That step calls
`github.rest.repos.getCollaboratorPermissionLevel` — a live GitHub API call that threads
`actor_trusted == 'true'` guards through all subsequent conditional steps.

gh-aw v0.74.4+ (PR [#31836](https://github.com/github/gh-aw/pull/31836)) adds a native
`on.skip-author-associations` field that gates activation using the `author_association` already
present in the event payload. This compiles to a job-level `if` expression with no extra API call.

The issue explicitly asks to kill the hand-rolled check and use the native version (Approach A in
the research comment; Approach B — keeping JS as a secondary check — was rejected because it
contradicts the issue's intent and adds layered complexity without removing the cost).

## Goals / Non-Goals

**Goals:**
- Replace all four factory workflow actor-trust steps with the native `skip-author-associations` field.
- Delete the `check-actor-trust.js` runner and the shared library functions that back it.
- Prune the corresponding tests to keep the test suite aligned with the live code.
- Regenerate all four factory workflow lock files.

**Non-Goals:**
- Changing the trust policy itself (this is a mechanical replacement, not a policy revision).
- Migrating non-factory workflows that may have separate actor-trust patterns.
- Updating the `workflows.yml` test harness beyond removing now-deleted function tests.

## Decisions

### 1. Use `skip-author-associations` to skip `none`, `first_timer`, `first_time_contributor`, `contributor`

This preserves the current effective policy: `OWNER`, `MEMBER`, and `COLLABORATOR` associations are
allowed (collaborators have explicit write grants; org members have at least repo membership).
External actors and first-timers are skipped.

For the `change-factory` workflow, which has both an `issues` and an `issue_comment` / `slash_command`
trigger, the `skip-author-associations` block must cover both event types:

```yaml
skip-author-associations:
  issues: [none, first_timer, first_time_contributor, contributor]
  issue_comment: [none, first_timer, first_time_contributor, contributor]
```

All other factory workflows have only the `issues` trigger variant.

### 2. Hardcode `actor_trusted=true` in `normalize_context` for issue-event paths

After removal of `check_actor_trust`, the `normalize_context` step in each workflow can emit
`actor_trusted=true` unconditionally for the issue-event path, because the native gate has already
filtered untrusted actors before the job runs. The dispatch-intake path already hardcodes
`actor_trusted=true` today; this change makes both paths consistent.

The `finalize_gate` step reads `ACTOR_TRUSTED` from `normalize_context`; it will continue to work
correctly because it will always receive `actor_trusted=true` for events that pass the native gate.

### 3. Remove `check-actor-trust.js`, `factoryCheckActorTrust`, `factoryActorTrustWhenSenderMissing`

After the native gate handles actor filtering, the JS runner and its library functions are dead
code. Remove them entirely to avoid the appearance that the JS check is still active.

### 4. Prune only actor-trust tests; leave all other test coverage intact

Tests for `factoryQualifyTriggerEvent`, `factoryCheckDuplicatePR`, `factoryComputeGateReason`, and
other non-trust logic in `factory-issue-shared.test.mjs` remain untouched. Only the
`factoryCheckActorTrust` and `factoryActorTrustWhenSenderMissing` test cases are removed from
`factory-issue-shared.test.mjs`, and the corresponding actor-trust test cases are removed from
`code-factory-issue.test.mjs` and `change-factory-issue.test.mjs`.

### 5. Regenerate all four lock files

After source edits, run `gh aw compile` for each workflow to produce updated lock files that match
the new source. The lock files are committed alongside the source changes in the same pull request.

## Risks / Trade-offs

- [Minor semantic shift for org read-only members] Association-based gating trusts all `MEMBER`
  accounts regardless of their specific repository permission level. An org member with only `read`
  permission was previously blocked but would now pass. For a maintainer-operated repository this
  edge case is negligible.
- [Four workflow files + JS + tests = broad diff] The changes are mechanical and low-risk, but the
  breadth means all four lock files must be regenerated; omitting any one would leave a stale lock.

## Migration Plan

1. Add `skip-author-associations` to all four `.md` files; remove `check_actor_trust` step and
   downstream `actor_trusted` conditions.
2. Simplify `normalize_context` in each workflow to hardcode `actor_trusted=true` for issue-event
   paths.
3. Remove `factoryCheckActorTrust`, `factoryActorTrustWhenSenderMissing`, and their exports from
   `factory-issue-shared.js`.
4. Delete `check-actor-trust.js`.
5. Prune actor-trust tests from `factory-issue-shared.test.mjs`,
   `code-factory-issue.test.mjs`, and `change-factory-issue.test.mjs`.
6. Run `gh aw compile` for all four workflows to regenerate lock files.
7. Verify CI passes with `make check` (or equivalent) and commit.

## Open Questions

- Should `COLLABORATOR` remain allowed (i.e., kept out of the skip list)? The current system trusts
  outside collaborators with write permissions; skipping only `[none, first_timer,
  first_time_contributor, contributor]` preserves this, which appears correct.
- For `issues:labeled` events, does `skip-author-associations` check the labeler's or the issue
  opener's `author_association`? The gh-aw docs indicate it checks the **triggering user's**
  association — worth confirming for `labeled` events specifically so the semantic is correctly
  documented.
- After removing the `check_actor_trust` step, `finalize_gate`'s `computeGateReason` receives
  `actor_trusted=true` unconditionally for issue-event paths. Confirm the resulting `gate_reason`
  text used by downstream workflows is still correct.
