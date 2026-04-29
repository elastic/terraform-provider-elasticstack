## 1. Documentation and CI Matrix

- [ ] 1.1 Update `README.md` so the documented minimum supported Elastic Stack version is `8.0` or higher.
- [ ] 1.2 Remove the `7.17.13` entry from `.github/workflows-src/test/workflow.yml.tmpl`.
- [ ] 1.3 Regenerate `.github/workflows/test.yml` with `make workflow-generate`.
- [ ] 1.4 Verify the generated workflow acceptance matrix contains no Elastic Stack 7.x entries.

## 2. Docker Workflow Support Floor

- [ ] 2.1 Update `Makefile` Fleet image fallback logic so it matches `8.0.%` and `8.1.%`, but not `7.17.%`.
- [ ] 2.2 Update comments and current OpenSpec wording for Makefile workflow behavior so older-version fallback language no longer mentions 7.17.

## 3. Remove Redundant Pre-8.0 Runtime Gates

- [ ] 3.1 Remove the transform minimum feature gate for Elasticsearch versions below `7.2.0`.
- [ ] 3.2 Always pass transform API operation timeouts for supported versions; remove the `7.17.0` timeout branch.
- [ ] 3.3 Remove transform setting gates whose minimum versions are below `8.0.0`, while preserving gates for `8.1.0`, `8.4.0`, `8.5.0`, and `8.8.0`.
- [ ] 3.4 Always decode configured transform `metadata` on create and update for supported versions.
- [ ] 3.5 Remove the ILM `allocate.total_shards_per_node` `7.16.0` compatibility gate while preserving later 8.x ILM gates.
- [ ] 3.6 Review acceptance tests with explicit 7.x-only skips or minimums and remove or update only those that are redundant under the 8.0+ support floor.

## 4. Generated Documentation and Stale References

- [ ] 4.1 Update live schema descriptions that advertise 7.15, 7.16, or 7.x support boundaries where those boundaries are no longer relevant.
- [ ] 4.2 Regenerate Terraform docs with `make docs-generate`.
- [ ] 4.3 Search current docs, workflow sources, Makefile, live OpenSpec specs, and implementation code for remaining current-support 7.x references; leave historical changelog entries and archived OpenSpec changes unchanged.

## 5. Verification

- [ ] 5.1 Run `openspec validate remove-7x-support --strict`.
- [ ] 5.2 Run `make check-workflows`.
- [ ] 5.3 Run focused Go tests for changed transform, ILM, and version-related acceptance helper code where unit coverage exists.
- [ ] 5.4 Run `make build`.
