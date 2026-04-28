## 1. PlanOnly acceptance harness

- [x] 1.1 Add `examples/examples.go` exporting `embed.FS` instances for `examples/resources` and `examples/data-sources`
- [x] 1.2 Add a Go acceptance test (e.g. `internal/acctest/examples_plan_test.go`) that walks both embedded filesystems and runs each `.tf` file as a subtest named after its path under `examples/` (such as `resources/...` or `data-sources/...`)
- [x] 1.3 In each subtest, write the example into an isolated config directory under `testdata/<test name>/plan/` (`<test name>` = full `t.Name()` including subtest path; via `acctest.NamedTestCaseDirectory("plan")`) and run `resource.Test` with `ProtoV6ProviderFactories: acctest.Providers`, `ConfigDirectory` pointing at that directory, and `PlanOnly: true`
- [x] 1.4 Use the existing acceptance-test precheck (`acctest.PreCheck(t)`) so the harness runs only when the standard live Elastic Stack environment is configured
- [x] 1.5 Skip `examples/cloud/` and `examples/provider/` via a static path skip-list documented in the harness
- [x] 1.6 Set `ExpectNonEmptyPlan: true` for all `examples/resources/` examples; for `examples/data-sources/`, set `ExpectNonEmptyPlan: true` when the root HCL body declares a top-level `resource` or `output` block, otherwise `false` (read-only / empty-friendly), matching `terraform-plugin-testing` PlanOnly checks
- [x] 1.7 Ensure plan diagnostics and `resource.Test` failure output include the offending example path clearly enough to fix failures from CI logs
- [x] 1.8 Mark subtests as `t.Parallel()` while respecting the normal `go test -parallel` cap to keep wall-clock reasonable without overwhelming the live stack

## 2. Example cleanup

- [ ] 2.1 Restructure `examples/resources/elasticstack_kibana_alerting_rule/` so `resource.tf`, `resource-index-rule.tf`, and `resource_rule_action_frequency.tf` each define their own connector and data-stream prerequisites and plan independently
- [ ] 2.2 Audit other multi-file example directories (e.g. `elasticstack_fleet_output/`, `elasticstack_fleet_integration/`, `elasticstack_kibana_security_role/`) and inline cross-file dependencies if any exist

## 3. Fix example bugs surfaced by the harness

- [ ] 3.1 Fix `examples/resources/elasticstack_elasticsearch_ml_datafeed/resource.tf` so `delayed_data_check_config` is configured as an attribute (not a block), per issue #2523
- [ ] 3.2 Run the new PlanOnly harness locally with the acceptance-test environment configured and enumerate every failing example file
- [ ] 3.3 Fix each surfaced failure (block-vs-attribute mistakes, renamed attributes, missing required fields, etc.) and verify the harness passes for every example
- [ ] 3.4 If any failure represents a genuine schema regression rather than a stale example, file a follow-up issue and resolve in scope or in a follow-on change before merging

## 4. Documentation and CI integration

- [ ] 4.1 Regenerate provider docs (`make docs-generate` or repo equivalent) so `docs/resources/*.md` and `docs/data-sources/*.md` reflect the cleaned example files
- [ ] 4.2 Confirm the new test is picked up by the acceptance-test CI path, using the standard `TF_ACC=1` and Elastic Stack environment variables
- [ ] 4.3 Measure local execution time for the harness and document it in the change's design notes if it exceeds a few minutes; tune parallelism if needed
- [ ] 4.4 Add a brief contributor note (in `dev-docs/high-level/development-workflow.md` or equivalent) explaining that every example `.tf` file must plan successfully in the PlanOnly acceptance harness and be self-contained (no cross-file references within the same example directory)

## 5. Verification

- [ ] 5.1 Run `make build` to ensure the harness compiles
- [ ] 5.2 Run the targeted acceptance test for the harness (for example, `TF_ACC=1 go test ./internal/acctest/... -run '^TestAccExamples_planOnly$'`) and confirm every example subtest passes
- [ ] 5.3 Run `make check-openspec` (or `make check-lint`) to confirm the OpenSpec artifacts are valid
- [ ] 5.4 Re-verify that issue #2523's reproduction (`delayed_data_check_config { ... }` as a block) is no longer present in `examples/resources/elasticstack_elasticsearch_ml_datafeed/`
