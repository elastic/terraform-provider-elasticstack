## Why

Running the full acceptance test suite on every PR is expensive and slow — the suite spans 101 test packages across 20+ Elastic Stack versions and takes up to 35 minutes per shard. Most PR changes affect only a small number of resources, and the relevant tests can be identified automatically from the diff. Adding `make targeted-testacc` lets developers (and CI on PRs) run only the tests that matter for a given branch, while the full suite remains authoritative on pushes to `main` and in the GitHub merge queue (when enabled).

## What Changes

- New Go tool at `scripts/targeted-testacc/` that computes the minimal set of acceptance test packages for the current branch diff.
- New `make targeted-testacc` and `make targeted-testacc-dry-run` Makefile targets.
- `provider.yml` CI workflow updated so PRs run `make targeted-testacc` instead of `make testacc`, with step-level shard gating to skip the Elastic Stack startup entirely when a shard has no packages to test.
- `provider.yml` CI workflow adds `merge_group` as a trigger (for future merge-queue enablement) and runs the full suite for those events.
- `provider.yml` CI workflow unchanged for `push` to `main` and `workflow_dispatch` — those always run the full suite.

## Capabilities

### New Capabilities

- `selective-acceptance-tests`: A Go tool and Makefile targets that identify the minimal set of acceptance test packages for a branch diff via two complementary phases — (1) Go reverse-dependency analysis to find packages that import changed code, and (2) TF entity name grep to find test suites that consume the affected Terraform resources in their testdata. Shard-count is determined dynamically by the tool based on the number of selected packages.

### Modified Capabilities

- `ci-provider-acceptance-tests`: The `provider.yml` acceptance test job gains a `compute-packages` step before stack startup. For PRs, this step runs the targeting tool and gates all expensive downstream steps (fleet pull, stack start, stack wait, API key, fleet setup, test run) on whether the shard has packages to test. For non-PR events (push to main, workflow_dispatch, merge_group), the step signals `has_packages=true` unconditionally and the test step falls back to `make testacc` (full suite, unchanged behaviour).

## Impact

- New Go source under `scripts/targeted-testacc/` (same module, no new module or `go tool` entry required).
- `Makefile`: two new targets, two new optional variables (`TARGETED_TESTACC_BASE`, `TARGETED_TESTACC_VERBOSE`).
- `.github/workflows/provider.yml`: `merge_group` trigger added; `compute-packages` step added; expensive steps gain `if:` conditions; test step switches between `targeted-testacc` and `testacc` based on event type and package availability. Static `shard: [0, 1]` matrix is preserved.
- No changes to existing test code, resource code, or any spec under `openspec/specs/`.
