## Why

All four factory workflows (research, code, change, reproducer) gate actor trust through a
dedicated `check_actor_trust` pre-activation step that calls
`github.rest.repos.getCollaboratorPermissionLevel` at runtime. gh-aw v0.74.4+ introduces a native
`on.skip-author-associations` frontmatter field that gates activation at the job level using the
`author_association` already present in the GitHub event payload — at zero API call cost.

The hand-rolled JS trust check, its supporting library functions, and the tests for them are
unnecessary maintenance burden once the native field is available.

## What Changes

- Add `skip-author-associations` to the `on:` block of all four factory workflow source files
  (`.github/workflows/research-factory-issue.md`, `code-factory-issue.md`,
  `change-factory-issue.md`, `reproducer-factory-issue.md`), skipping activations from actors with
  associations `none`, `first_timer`, `first_time_contributor`, or `contributor`.
- Remove the `check_actor_trust` step and all downstream conditions that gate on
  `steps.check_actor_trust.outputs.actor_trusted == 'true'` from each workflow.
- Simplify `normalize_context` in each workflow to hardcode `actor_trusted=true` for
  issue-event paths (the native gate guarantees trust for any event that reaches the step).
- Delete `.github/scripts/workflows/lib/factory-runners/check-actor-trust.js`.
- Remove `factoryCheckActorTrust`, `factoryActorTrustWhenSenderMissing`, and their exports from
  `.github/scripts/workflows/lib/factory-issue-shared.js`.
- Prune the corresponding tests from `factory-issue-shared.test.mjs`,
  `code-factory-issue.test.mjs`, and `change-factory-issue.test.mjs`.
- Regenerate lock files for all four workflows via `gh aw compile`.

## Capabilities

### Modified Capabilities
- `ci-factory-actor-trust-gate`: replaces the runtime JS permission check with the declarative
  gh-aw `skip-author-associations` frontmatter field across all four factory workflows

## Impact

- Eliminates one GitHub API call per eligible factory event
- Removes ~60 lines of JS and test code
- Trust policy becomes visible in workflow YAML source rather than hidden in a JS helper
- Minor semantic shift: org members with `read`-only permission were previously blocked by the
  permission check and would now be allowed by association-based gating; this edge case is
  negligible for a maintainer-operated org repo
- All four factory workflow lock files must be regenerated after the source changes
