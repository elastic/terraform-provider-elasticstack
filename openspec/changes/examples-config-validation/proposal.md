## Why

Examples under `examples/resources/` and `examples/data-sources/` feed generated provider documentation (including resource/data source-style pages, templates, and guides), but nothing currently verifies that each covered file can be planned against the provider. Issue [#2523](https://github.com/elastic/terraform-provider-elasticstack/issues/2523) is a recent instance of this: the `elasticsearch_ml_datafeed` example uses `delayed_data_check_config` as a block when the schema declares it as an attribute, so the documented snippet fails immediately when a user copies it into their configuration. This class of bug is mechanical, easy to introduce during refactors, and only ever caught by user reports.

A PlanOnly acceptance test for every example file closes that gap using the same provider-test harness already used by resource and data source acceptance tests. Planning each example catches schema mismatches and provider-side plan validation without adding the complexity of shelling out to `terraform validate` for every file.

## What Changes

- Add a Go acceptance test that, for every `*.tf` file under `examples/resources/` and `examples/data-sources/`, runs the file through a `terraform-plugin-testing` `PlanOnly` step against the in-process provider, with one subtest per file so failures attribute to the offending snippet.
- Establish a convention that each example `.tf` file is **self-contained**: it must not reference resources or data sources defined only in sibling files. Restructure `examples/resources/elasticstack_kibana_alerting_rule/` (currently the only directory that violates this) so each `.tf` file stands on its own.
- Skip `examples/cloud/` (uses the `ec` provider), `examples/provider/` (provider-config snippets), plus a minimal **enumerated per-file skip list** in the harness source for snippets that cannot be planned in isolation (for example `terraform_remote_state` multi-root workflows or **`hashicorp/time`** rotations that require CLI provider installation). Paths and rationales MUST be documented beside that list in code. New per-file skips require a code change — not sentinel comments alone.
- Run under normal acceptance-test prechecks so examples that plan data sources or provider-configured resources use the same live Elastic Stack environment as the existing acceptance suite.
- Fix every example surfaced as broken by the new test on first run, including the `delayed_data_check_config` bug from #2523 and any other latent schema mismatches.

## Capabilities

### New Capabilities

- `examples-validation`: a PlanOnly acceptance harness that verifies every covered example file can be planned against the provider.

### Modified Capabilities

None.

## Impact

- **Affected code**: new `examples/examples.go` (embed FS), new `internal/acctest/examples_plan_test.go` (or equivalent location), targeted updates to `.tf` files under `examples/resources/` to fix latent schema bugs and to make the `elasticstack_kibana_alerting_rule` examples self-contained.
- **Affected interfaces**: no provider schema changes. Documentation regeneration may produce diffs only for the example files actually changed by this work (alerting-rule restructuring and any examples whose schema bugs are fixed).
- **Backward compatibility**: no provider behaviour change for users who copy existing examples. Embedded `provider` and `elasticsearch_connection` blocks in examples remain valid input to the harness and are deliberately left in place to keep the diff minimal.
- **CI**: the new test runs in the same **`make testacc`** path (`TF_ACC` + Elastic Stack env) whenever **provider-impacting** changes trigger the acceptance matrix; PRs classified as OpenSpec-only may skip acceptance. PlanOnly may still touch data sources and plan-time validation against a live stack. Expected wall-clock impact is modest for ~115 example files; subtests use `t.Parallel()` plus a bounded semaphore (`maxConcurrentExamplesPlanHarness = 4`) so simultaneous Plan workloads do not overwhelm the muxed provider or cluster (see harness design notes).
