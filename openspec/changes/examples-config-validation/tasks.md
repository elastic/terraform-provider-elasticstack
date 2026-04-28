## 1. Validation harness

- [ ] 1.1 Add `examples/examples.go` exporting `embed.FS` instances for `examples/resources` and `examples/data-sources`
- [ ] 1.2 Add a Go test (e.g. `internal/acctest/examples_validate_test.go`) that walks both embedded filesystems and runs each `.tf` file as a subtest named after its repo-relative path
- [ ] 1.3 In each subtest, write the example contents plus a generated provider-pin file (`terraform { required_providers { ... } }`) into a per-test tempdir, then invoke `terraform init` and `terraform validate` via `terraform-exec`
- [ ] 1.4 Configure the harness to discover the locally built provider via `dev_overrides` (or filesystem mirror if `dev_overrides` proves noisy)
- [ ] 1.5 Skip `examples/cloud/` and `examples/provider/` via a static path skip-list documented in the harness
- [ ] 1.6 Surface `terraform validate` diagnostics in the test failure message so the offending file, line, and message are visible without re-running locally
- [ ] 1.7 Mark subtests as `t.Parallel()` and share a `TF_PLUGIN_CACHE_DIR` across subtests to keep wall-clock reasonable

## 2. Example cleanup

- [ ] 2.1 Restructure `examples/resources/elasticstack_kibana_alerting_rule/` so `resource.tf`, `resource-index-rule.tf`, and `resource_rule_action_frequency.tf` each define their own connector and data-stream prerequisites and validate independently
- [ ] 2.2 Audit other multi-file example directories (e.g. `elasticstack_fleet_output/`, `elasticstack_fleet_integration/`, `elasticstack_kibana_security_role/`) and inline cross-file dependencies if any exist

## 3. Fix example bugs surfaced by the harness

- [ ] 3.1 Fix `examples/resources/elasticstack_elasticsearch_ml_datafeed/resource.tf` so `delayed_data_check_config` is configured as an attribute (not a block), per issue #2523
- [ ] 3.2 Run the new harness locally and enumerate every failing example file
- [ ] 3.3 Fix each surfaced failure (block-vs-attribute mistakes, renamed attributes, missing required fields, etc.) and verify the harness passes for every example
- [ ] 3.4 If any failure represents a genuine schema regression rather than a stale example, file a follow-up issue and resolve in scope or in a follow-on change before merging

## 4. Documentation and CI integration

- [ ] 4.1 Regenerate provider docs (`make docs-generate` or repo equivalent) so `docs/resources/*.md` and `docs/data-sources/*.md` reflect the cleaned example files
- [ ] 4.2 Confirm the new test is picked up by the standard `go test ./...` invocation used in CI, with no extra environment variables required
- [ ] 4.3 Measure local execution time for the harness and document it in the change's design notes if it exceeds a few minutes; tune parallelism if needed
- [ ] 4.4 Add a brief contributor note (in `dev-docs/high-level/development-workflow.md` or equivalent) explaining that every example `.tf` file must validate against the provider schema and be self-contained (no cross-file references within the same example directory)

## 5. Verification

- [ ] 5.1 Run `make build` to ensure the harness compiles
- [ ] 5.2 Run `go test ./internal/acctest/...` (or whichever package hosts the harness) and confirm every example subtest passes
- [ ] 5.3 Run `make check-openspec` (or `make check-lint`) to confirm the OpenSpec artifacts are valid
- [ ] 5.4 Re-verify that issue #2523's reproduction (`delayed_data_check_config { ... }` as a block) is no longer present in `examples/resources/elasticstack_elasticsearch_ml_datafeed/`
