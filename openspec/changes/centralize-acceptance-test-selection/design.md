## Context

See proposal.md for motivation. The current test matrix has a fixed shard axis,
and each version-and-shard job independently determines its targeted package
list. That repeats selection work and creates empty jobs for small selections.
The selector's existing threshold already defines when a selected package set
is large enough to split.

## Goals / Non-Goals

**Goals:**

- Establish one preparation boundary that supplies both the pinned version
  matrix and a targeted pull-request shard plan.
- Keep selection and sharding policy in the selector so all consumers use the
  same plan.
- Avoid costly test-job setup when the central plan proves that a job has no
  packages.
- Preserve fixed, full-suite, two-shard execution outside pull requests.

**Non-Goals:**

- Rebalance shards by historical execution duration.
- Change the 30-package threshold, acceptance-test version list, or version
  coverage policy.
- Change provider-gate semantics for provider changes.
- Add a new service, persistent state, or workflow job.

## Decisions

### Extend matrix preparation

The existing matrix-loading job becomes the preparation boundary. It continues
to load pinned versions and version flags and, for pull requests only, also
computes the targeted shard plan once. This reuses an existing prerequisite and
avoids a separate job that would duplicate checkout and Go setup.

For non-pull-request events, preparation emits the existing fixed shard indices
and does not invoke targeted selection. The downstream full-suite path therefore
remains unchanged.

### Make the selector the shard-plan authority

`targeted-testacc` emits compact JSON by default, including the complete sorted
package selection, ordered shard assignments, a `has_packages` indicator, and
selection rationale. `--dry-run` uses the same JSON contract. The selector owns
the existing package-count threshold and round-robin assignment, and no longer
accepts a shard index.

The workflow preparation step derives GitHub Actions-friendly values from the
plan: numeric shard indices, package arrays indexed by shard, and the
has-packages result. This adapter must not recreate selection or sharding
rules. The local Make target reads the same plan and chooses the requested
array for its local shard index.

### Build the PR matrix from populated assignments

The acceptance workflow keeps versions as one matrix dimension and supplies
only prepared numeric shard indices as the other. A selection below 30 packages
therefore creates one test job per version; a selection at or above 30 creates
two populated test jobs per version.

For a zero-package pull-request selection, preparation emits one empty package
array and one shard index. The resulting job succeeds without disk cleanup,
checkout, dependency setup, stack lifecycle, diagnostics, teardown, or
acceptance-test execution. Keeping these no-op jobs preserves the provider
gate's existing expectation that provider changes do not unexpectedly skip the
test job.

### Preserve fail-closed planning

Preparation fetches the pull-request base once. If that fetch fails, it invokes
the selector without an explicit base so the selector uses its existing
conservative fallback. If selection, JSON decoding, or plan validation fails,
preparation fails rather than treating the result as an empty plan.

Plan validation enforces that zero selected packages yields exactly one empty
assignment, while every nonempty selection is covered exactly once by nonempty
assignments. A requested shard count greater than the package count is capped
to populated assignments.

### Test at policy and workflow boundaries

Selector unit tests cover JSON output, rationale, threshold behavior, invalid
shard counts, complete package coverage, no-empty-shard behavior, and dry-run
output. Workflow-script tests cover PR base resolution and fallback, selector
failure, non-PR fixed-shard results, and zero/small/large plans. Workflow
structure tests verify that preparation owns selection and that no-op jobs gate
all costly setup. Existing targeted acceptance tests remain the live-stack
validation path.

The JSON output change lands atomically with all known consumers. Reverting the
change restores the existing fixed two-shard workflow; no persistent migration
or external rollback action is required.

## Risks / Trade-offs

- [Risk] The compact plan is passed through GitHub Actions job outputs, which
  have size limits. -> Mitigation: package import paths and two shard arrays are
  expected to remain well below the output limit; preparation fails if it cannot
  publish a valid plan rather than truncating it.
- [Risk] GitHub Actions expression indexing could select the wrong package
  array. -> Mitigation: use numeric shard indices derived from plan order and
  test zero, small, and large plans in workflow-structure tests.
- [Risk] Removing plaintext output can break local consumers. -> Mitigation:
  update every repository consumer in the same change and verify local Make
  targets against the JSON contract.
- [Risk] A no-op matrix job still consumes minimal runner scheduling time. ->
  Mitigation: preserve it only to maintain the existing gate contract; omit
  every expensive job step.
