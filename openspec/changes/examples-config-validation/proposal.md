## Why

Examples under `examples/resources/` and `examples/data-sources/` are surfaced verbatim in the generated provider docs, but nothing currently verifies that they parse against the provider schema. Issue [#2523](https://github.com/elastic/terraform-provider-elasticstack/issues/2523) is a recent instance of this: the `elasticsearch_ml_datafeed` example uses `delayed_data_check_config` as a block when the schema declares it as an attribute, so the documented snippet fails immediately when a user copies it into their configuration. This class of bug is mechanical, easy to introduce during refactors, and only ever caught by user reports.

A test that runs `terraform validate` against every example file closes that gap and turns the schema itself into the lint rule for examples going forward.

## What Changes

- Add a Go test that, for every `*.tf` file under `examples/resources/` and `examples/data-sources/`, runs the file through `terraform init` + `terraform validate` against the locally built provider, with one subtest per file so failures attribute to the offending snippet.
- Establish a convention that each example `.tf` file is **self-contained**: it must not reference resources or data sources defined only in sibling files. Restructure `examples/resources/elasticstack_kibana_alerting_rule/` (currently the only directory that violates this) so each `.tf` file stands on its own.
- Skip `examples/cloud/` (uses the `ec` provider) and `examples/provider/` (provider-config snippets, not standalone configs) from the harness via a static path skip-list.
- Fix every example surfaced as broken by the new test on first run, including the `delayed_data_check_config` bug from #2523 and any other latent schema mismatches.

## Capabilities

### New Capabilities

- `examples-validation`: a harness that validates every example file against the provider schema, run as part of the standard `go test` suite.

### Modified Capabilities

None.

## Impact

- **Affected code**: new `examples/examples.go` (embed FS), new `internal/acctest/examples_validate_test.go` (or equivalent location), targeted updates to `.tf` files under `examples/resources/` to fix latent schema bugs and to make the `elasticstack_kibana_alerting_rule` examples self-contained.
- **Affected interfaces**: no provider schema changes. Documentation regeneration may produce diffs only for the example files actually changed by this work (alerting-rule restructuring and any examples whose schema bugs are fixed).
- **Backward compatibility**: no behaviour change for users who copy existing examples. Embedded `provider` and `elasticsearch_connection` blocks in examples remain valid input to the harness — `terraform validate` accepts them — so they are deliberately left in place to keep the diff minimal.
- **CI**: the new test runs without a live Elastic stack (`terraform validate` does not call data sources or refresh state) and integrates into the existing `go test` invocation. Expected wall-clock impact is single-digit minutes for ~115 example files; can be parallelised via `t.Parallel()` since each subtest uses its own working directory.
