## Why

Examples under `examples/resources/` and `examples/data-sources/` are surfaced verbatim in the generated provider docs, but nothing currently verifies that they parse against the provider schema. Issue [#2523](https://github.com/elastic/terraform-provider-elasticstack/issues/2523) is a recent instance of this: the `elasticsearch_ml_datafeed` example uses `delayed_data_check_config` as a block when the schema declares it as an attribute, so the documented snippet fails immediately when a user copies it into their configuration. This class of bug is mechanical, easy to introduce during refactors, and only ever caught by user reports.

A test that runs `terraform validate` against every example file closes that gap and turns the schema itself into the lint rule for examples going forward.

## What Changes

- Add a Go test that, for every `*.tf` file under `examples/resources/` and `examples/data-sources/`, runs the file through `terraform init` + `terraform validate` against the locally built provider, with one subtest per file so failures attribute to the offending snippet.
- Establish a convention that each example `.tf` file is **self-contained**: it must not reference resources or data sources defined only in sibling files. Restructure `examples/resources/elasticstack_kibana_alerting_rule/` (currently the only directory that violates this) so each `.tf` file stands on its own.
- Strip embedded `provider "elasticstack" {}` blocks and per-resource `elasticsearch_connection {}` overrides from existing example files. The validation harness injects provider configuration; duplicate provider blocks in a snippet would conflict with that and are also confusing for documentation readers.
- Skip `examples/cloud/` (uses the `ec` provider) and `examples/provider/` (provider-config snippets, not standalone configs) from the harness via a static path skip-list.
- Fix every example surfaced as broken by the new test on first run, including the `delayed_data_check_config` bug from #2523 and any other latent schema mismatches.

## Capabilities

### New Capabilities

- `examples-validation`: a harness that validates every example file against the provider schema, run as part of the standard `go test` suite.

### Modified Capabilities

None.

## Impact

- **Affected code**: new `examples/examples.go` (embed FS), new `internal/acctest/examples_validate_test.go` (or equivalent location), updates to most `.tf` files under `examples/resources/` and `examples/data-sources/` to remove provider blocks and fix latent schema bugs.
- **Affected interfaces**: no provider schema changes. Documentation regeneration may produce diffs because example contents change.
- **Backward compatibility**: stripped provider blocks change the rendered docs. Users who copy an example will now also need to provide their own provider configuration (which they typically already do), so this is a docs-quality improvement rather than a regression.
- **CI**: the new test runs without a live Elastic stack (`terraform validate` does not call data sources or refresh state) and integrates into the existing `go test` invocation. Expected wall-clock impact is single-digit minutes for ~115 example files; can be parallelised via `t.Parallel()` since each subtest uses its own working directory.
