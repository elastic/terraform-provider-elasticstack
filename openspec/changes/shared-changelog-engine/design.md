## Context

The current changelog automation mixes two concerns in one workflow: scheduled maintenance of the `## [Unreleased]` section and release-specific regeneration of a concrete `## [x.y.z] - <date>` section. Release mode is currently activated from `pull_request_target` events on `prep-release-*` branches, which makes final release changelog generation depend on PR event delivery and timing rather than on the release-preparation workflow that actually owns the release branch contents.

The repository already contains deterministic changelog parsing and rendering logic spread across inline `actions/github-script` steps, small shared JavaScript helpers under `.github/workflows-src/lib/`, and a small `scripts/changelog-generation` Go entrypoint. The target design should preserve deterministic assembly and existing PR-body changelog-contract rules, while moving release-mode invocation to an explicit synchronous step in release preparation.

## Goals / Non-Goals

**Goals:**
- Introduce a shared repository-authored changelog engine that can be invoked from multiple workflows with explicit mode inputs.
- Make `prep-release.yml` responsible for invoking release-mode changelog generation before it creates or updates the release PR.
- Keep scheduled unreleased generation in `changelog-generation.yml` and add an explicit manual `workflow_dispatch` release-mode fallback in the same workflow.
- Keep changelog assembly fail-fast: invalid or missing PR changelog contracts in the authoritative range must fail release preparation and manual release regeneration.
- Keep workflow orchestration responsibilities separate from changelog assembly responsibilities.

**Non-Goals:**
- Changing the PR-body changelog contract itself.
- Changing the singleton `generated-changelog` branch/PR model for unreleased maintenance.
- Introducing agentic changelog synthesis or LLM-authored release notes.
- Generalizing the engine into a cross-repository tool.

## Decisions

### Use an explicit shared changelog engine instead of event-inferred release mode
The changelog engine will be invoked with explicit workflow inputs that select `release` or `unreleased` mode. Release-mode behavior will no longer be inferred from `pull_request_target` event metadata.

This makes release preparation deterministic and enables a manual release fallback without needing synthetic PR events. It also removes the need for the changelog engine to interpret GitHub event shape as its primary control surface.

**Alternatives considered:**
- Keep `pull_request_target` as the release trigger: rejected because release preparation should not depend on downstream PR-event timing.
- Use `workflow_run` or dispatch from `prep-release.yml`: rejected because explicit synchronous invocation inside release preparation is simpler and provides fail-fast semantics.

### The shared engine resolves merged PRs through the GitHub API using the workflow token
The engine will own authoritative-range discovery and merged-PR resolution, including GitHub API lookups needed to map commits in the compare range to merged pull requests and to retrieve their bodies, labels, and other required metadata. Workflows will provide authenticated environment context through the built-in workflow token.

This keeps the core changelog assembly path self-contained and reusable across release and unreleased modes. It also avoids splitting the authoritative changelog assembly pipeline between workflow glue and engine internals.

**Alternatives considered:**
- Have workflows gather merged PR metadata and pass a manifest into the engine: rejected because it leaves key release-note semantics distributed across multiple workflows.
- Use only local git metadata without GitHub API resolution: rejected because PR labels and bodies are authoritative inputs to changelog assembly.

### Workflows remain responsible for checkout, commit/push, and PR management
The shared engine will mutate `CHANGELOG.md` in the checked-out worktree and emit structured outputs such as compare range, target version, and whether user-facing changes were rendered. Workflows will continue to own branch checkout, commit creation, push destination, PR create/reuse logic, `no-changelog` labeling, and PR body refresh.

This keeps the engine focused on deterministic content generation and preserves straightforward workflow ownership of repository mutations beyond the changelog file itself.

**Alternatives considered:**
- Move PR creation/editing into the engine: rejected because it would blur content generation with branch/PR orchestration and make testing/retries more complex.

### Release preparation produces a single release-preparation commit
`prep-release.yml` will combine the version bump and final release changelog update into a single deterministic release-preparation commit before pushing the branch.

This matches the desired operator experience: the workflow should leave behind a ready-to-review release PR whose content already reflects the final version bump and release changelog section.

**Alternatives considered:**
- Separate commits for version bump and changelog update: rejected because it adds orchestration complexity without meaningful review benefits in a squash-merge workflow.

### Manual release fallback lives in the existing changelog-generation workflow via dispatch inputs
`changelog-generation.yml` will continue to serve scheduled unreleased maintenance, and its `workflow_dispatch` entrypoint will accept explicit inputs for release-mode execution (including target version) so maintainers can manually regenerate a release section when needed.

This preserves a single operational place for changelog regeneration without keeping automatic release-mode triggers.

**Alternatives considered:**
- Create a separate manual recovery workflow: rejected because the same engine and workflow can support both scheduled unreleased runs and explicit release-mode recovery cleanly.

## Risks / Trade-offs

- **Engine extraction touches multiple existing code paths** → Mitigation: preserve current parsing/rendering semantics and move logic behind stable tests before simplifying workflow glue.
- **Using the workflow token for GitHub API resolution couples the engine to Actions execution context** → Mitigation: make the token/env contract explicit and keep workflow-side authentication simple and standard.
- **Removing `pull_request_target` eliminates automatic release-mode retries on PR activity** → Mitigation: release preparation becomes the primary synchronous path, and `workflow_dispatch` provides explicit manual recovery.
- **Single-commit release preparation changes commit shape** → Mitigation: document the new deterministic commit behavior in the release-preparation spec and workflow output.

## Migration Plan

1. Extract or consolidate the deterministic changelog engine behind a reusable repository-authored script interface.
2. Update `prep-release.yml` to invoke the engine in release mode after applying the version bump and before creating/updating the PR.
3. Update `changelog-generation.yml` and its template to remove `pull_request_target`, add explicit `workflow_dispatch` inputs for release mode, and invoke the shared engine in unreleased or release mode accordingly.
4. Preserve or adapt PR-management helper logic so unreleased mode still maintains the singleton `generated-changelog` PR and release mode can refresh release PR metadata when manually dispatched.
5. Regenerate the compiled workflow YAML and validate behavior through existing and new tests.

## Open Questions

- None at proposal time.
