## 1. `scripts/version-matrix` engine

- [ ] 1.1 Scaffold Go module `scripts/version-matrix/` (`main.go` + package files), mirroring the `scripts/changelog/` and `scripts/auto-approve/` layout (thin `main.go` entrypoint, testable package logic)
- [ ] 1.2 Implement GA-tag discovery: list `elastic/elasticsearch` git tags via the GitHub API (authenticated with the workflow token), filter to `^v(8|9)\.\d+\.\d+$`, group by minor, keep the max patch per minor
- [ ] 1.3 Implement SNAPSHOT-label lookup against `https://snapshots.elastic.co/latest/master.json`; fail the run (do not silently drop the entry) if the endpoint is unavailable or the response is unparsable
- [ ] 1.4 Implement the Docker image existence probe: for each newly computed GA patch not already in the pinned artifact, check that `docker.elastic.co/elasticsearch/elasticsearch:<version>` resolves; on failure, fall back to the previously pinned version for that minor instead of failing the run
- [ ] 1.5 Implement pinned-artifact read/diff/write against `.github/versions/acceptance-test-matrix.json` (sorted ascending, SNAPSHOT entry last)
- [ ] 1.6 Add unit tests: tag-filtering/latest-patch-per-minor logic, SNAPSHOT-label parsing, image-probe fallback behavior, diff/no-op detection — using fixture HTTP responses, no live network calls
- [ ] 1.7 Seed `.github/versions/acceptance-test-matrix.json` with the current (research-time-corrected) list so the first scheduled run computes a small, reviewable diff rather than rewriting the whole file

## 2. `version-matrix-generation` workflow

- [ ] 2.1 Add `.github/workflows/version-matrix-generation.yml`: `schedule` (daily, offset from `changelog-generation.yml`'s cron) + `workflow_dispatch`, `permissions: contents: write, pull-requests: write`
- [ ] 2.2 Checkout, `actions/setup-go`, run `go run ./scripts/version-matrix compute` (or equivalent subcommand) to produce the desired list and diff it against the pinned artifact
- [ ] 2.3 When unchanged: exit without further steps (no commit, no branch push, no PR touch)
- [ ] 2.4 When changed: commit the rewritten artifact as `github-actions[bot]`, push to stable branch `acceptance-test-version-matrix`, then push an empty commit re-authenticated with `GH_AW_CI_TRIGGER_TOKEN` to trigger downstream CI — mirroring `changelog-generation.yml`'s unreleased-mode push steps
- [ ] 2.5 Look up an existing open PR from `acceptance-test-version-matrix` → default branch; create it (with `no-changelog` label and an all-or-nothing PR body) when absent, or update its body when present
- [ ] 2.6 Add/extend a workflow-script unit test (or Go test, matching whichever layer owns the lookup/create/update logic) covering: no existing PR → create; existing PR → update; unchanged list → no-op leaves an existing PR untouched

## 3. `provider.yml` matrix sourcing

- [ ] 3.1 Add a preceding job (e.g. `load-matrix`) that checks out the repo, reads `.github/versions/acceptance-test-matrix.json`, and exposes its contents as a JSON-array job output
- [ ] 3.2 Change the `test` job's `strategy.matrix.version` to `${{ fromJson(needs.load-matrix.outputs.versions) }}`; keep `shard: [0, 1]` static; make `test` depend on the new job in addition to `build`/`classify`
- [ ] 3.3 Remove the `strategy.matrix.include` block entirely
- [ ] 3.4 Replace the `runs-on` expression with a version-range check against `matrix.version` (`startsWith(matrix.version, '8.0.') || ... || startsWith(matrix.version, '8.4.')` → `ubuntu-22.04`, else `ubuntu-latest`)
- [ ] 3.5 Replace the fleet-image pre-pull step's `if: matrix.fleetImage` condition and the `FLEET_IMAGE` env value with version-range checks (`8.0.`/`8.1.` → Docker Hub `elastic/elastic-agent`, else `docker.elastic.co/elastic-agent/elastic-agent`)
- [ ] 3.6 Replace the forced-synthetics `if:` condition's exact-patch list (`8.14.3`, `8.15.5`, `8.16.6`, `8.17.10`) with a minor-range check (`startsWith(matrix.version, '8.14.') || ... || startsWith(matrix.version, '8.17.')`)
- [ ] 3.7 Confirm no other step references `matrix.runner` or `matrix.fleetImage`

## 4. Change classification

- [ ] 4.1 Update `.github/scripts/workflows/lib/classify-changes.js` so `.github/versions/acceptance-test-matrix.json` is provider-impacting (i.e., excluded from the "any path under `.github/`" non-impacting carve-out)
- [ ] 4.2 Update/add table-driven unit tests for `classify-changes.js` covering: a PR that changes only the pinned versions artifact → `provider_changes=true`; a PR that changes other `.github/` paths (not `provider.yml`, not the versions artifact) → unaffected/`provider_changes=false` as before

## 5. Auto-approve `version-matrix` category

- [ ] 5.1 Add a `version-matrix` category to `scripts/auto-approve/evaluator.go`: selector = same-repo PR with head branch exactly `acceptance-test-version-matrix`; gates = every commit authored by `github-actions[bot]`, every changed file path exactly `.github/versions/acceptance-test-matrix.json`
- [ ] 5.2 Wire the category into global approval gates (including Provider Gate success) with no diff-threshold gate, following the existing `generated-changelog` category as the template
- [ ] 5.3 Add table-driven unit tests in `scripts/auto-approve/evaluator_test.go` covering: matching branch + bot commits + allowlisted file → approve; foreign commit author → no approve; extra changed file → no approve; wrong branch name → category does not match
- [ ] 5.4 Confirm no auto-merge call is added anywhere in `scripts/auto-approve/`

## 6. Spec/doc consistency

- [ ] 6.1 Confirm `openspec/specs/ci-build-lint-test/spec.md`'s Schema section (or prose) does not still imply a hand-maintained `strategy.matrix.version` list once this change is archived
- [ ] 6.2 Confirm no other spec or doc under `openspec/specs/` or `dev-docs/` references the removed `strategy.matrix.include` block or the exact-patch synthetics/runner conditions
