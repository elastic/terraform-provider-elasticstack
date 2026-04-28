## Context

The provider ships ~115 example `.tf` files under `examples/resources/` and `examples/data-sources/`. They are consumed two ways:

1. `terraform-plugin-docs` embeds them into the generated reference docs at `docs/resources/<name>.md` and `docs/data-sources/<name>.md`.
2. End users copy them into their own configurations as starting points.

Today there is no automated check that these snippets parse against the provider's actual schema. Schema drift between the provider and its documented examples (e.g. a block-form attribute being changed to an object attribute) is silent until a user reports a broken copy-paste, as in issue #2523.

This change introduces a single Go test that validates every example against the locally built provider. The design favours keeping the test cheap, fast, and correct over catching every possible bug class — `terraform validate` is the right tool for "does this snippet match the schema", and going further (PlanOnly, full Apply) trades dramatically more execution cost and infrastructure dependency for a small marginal gain in coverage.

## Goals / Non-Goals

**Goals**:
- Catch schema-mismatch bugs in example `.tf` files (block vs. attribute, wrong attribute names, type errors, missing required attributes) before they reach users.
- Attribute failures to a specific example file so the fix is obvious.
- Run as part of the normal `go test ./...` suite without requiring a live Elastic stack.
- Make example files self-contained so users can copy any single snippet and have it parse.
- Establish "examples must validate" as an enforceable convention going forward.

**Non-Goals**:
- Verify that examples produce a valid plan (would force a third of data-source examples to depend on a live stack).
- Verify that examples can be applied successfully (different and much more expensive guarantee, already partially covered by per-resource acceptance tests).
- Lint examples for stylistic concerns (formatting, comment quality).
- Test combinations of examples that span multiple resource directories.

## Decisions

### 1. Use `terraform validate` rather than `PlanOnly`

`terraform validate` parses the configuration and checks it against the provider schema, including block-vs-attribute, attribute names, types, required attributes, and intra-config references. It does **not** call data sources or refresh state, so it has no live-backend dependency.

`PlanOnly` via `terraform-plugin-testing` would additionally evaluate data sources and resolve computed cross-resource references. In this repo, only a small minority of resource examples reference data sources at all, and most data-source examples are local-compute (`elasticstack_elasticsearch_ingest_processor_*`). The marginal coverage `PlanOnly` would add over `validate` is small — and would require a live Elastic and Kibana stack for a meaningful subset of examples, pushing the test out of unit-test territory.

The bug class motivating this change (#2523, block vs. attribute) is squarely in `validate`'s wheelhouse.

### 2. Drive the harness via `terraform-exec`, not `terraform-plugin-testing`

`terraform-plugin-testing` does not expose a "validate-only" step; its lifecycle starts at plan. Using `terraform-exec` (`tfexec.NewTerraform(...).Init(...).Validate(...)`) keeps the harness aligned with what is actually being tested and avoids dragging in lifecycle behaviour the test does not want.

The harness will:
1. Embed the contents of `examples/resources/` and `examples/data-sources/` via `embed.FS` in `examples/examples.go`.
2. For each example `.tf` file, write it to a temporary directory together with a generated `terraform.tf` that pins the locally built provider via `dev_overrides` (or the existing acceptance-test provider factory if cleaner — see open question 1).
3. Run `terraform init` and `terraform validate`, surfacing structured diagnostics if validation fails.
4. Use `t.Run("<relative-path>", ...)` so the failing file is named in the test output and matches `-run` filters.

### 3. One subtest per example file

The user explicitly preferred per-file failure attribution. To make per-file isolation work, every example `.tf` file must be **self-contained**: it cannot reference resources or data sources defined only in sibling files. This is also a usability improvement — readers of the generated docs see one snippet at a time and reasonably expect each to be runnable on its own.

Today, only one directory violates this: `examples/resources/elasticstack_kibana_alerting_rule/`. Its `resource-index-rule.tf` and `resource_rule_action_frequency.tf` both reference `elasticstack_kibana_action_connector.index_example` and `elasticstack_elasticsearch_data_stream.my_data_stream` defined in the directory's `resource.tf`. We will inline those dependencies into each file. The cost is some duplication in the generated docs; the benefit is that every documented snippet is independently usable.

### 4. Leave embedded provider configuration in examples alone

An earlier version of this design considered stripping every embedded `provider "elasticstack" { ... }` block and per-resource `elasticsearch_connection { ... }` block from existing examples. That cleanup is unnecessary for the validate-based harness: `terraform validate` accepts those blocks as schema-conformant input and does not require the harness to inject its own provider configuration. The blocks are also stylistic choices in the rendered docs (some users prefer the explicit setup, others prefer it implicit), and stripping them would produce a large, behaviourally-irrelevant docs diff.

This change therefore leaves embedded provider and connection blocks in place. The harness itself does not write a provider configuration into the per-test working directory; it only writes the `terraform { required_providers { ... } }` pin needed for `terraform init` to discover the locally built provider. If a future change wants to standardise example docs (e.g. for stylistic consistency), it can do so as a separate, focused cleanup with its own justification.

### 5. Static skip-list for non-validatable directories

Two directories under `examples/` cannot be validated by this harness:

- `examples/cloud/`: uses the `ec` (Elastic Cloud) provider, not `elasticstack`. Out of scope for this test.
- `examples/provider/`: contains snippets demonstrating provider configuration itself. They are intentionally partial and not standalone configurations.

We will hard-code these paths in the harness skip-list. We considered an in-file sentinel comment (`# acctest:skip reason=...`) for richer per-file metadata, but with "fix everything" as the policy (see decision 6) the only legitimate skips are these two structural cases, and a static list keeps the mechanism trivial to audit.

### 6. Fix every example the harness flags

Rather than landing the test in a partially-disabled state, this change fixes every example that the new harness reports as broken. We expect the `delayed_data_check_config` bug from #2523 plus a small number of latent issues — most likely a handful of block-vs-attribute or attribute-rename mistakes that have accumulated since the relevant resources were last touched.

### 7. Run alongside the regular `go test` suite

Because `terraform validate` does not need a live stack, this test does not need the `TF_ACC=1` gating used for true acceptance tests. It can be a standard `*_test.go` test in `internal/acctest/` (or a new lightweight package) that runs in `make test`. CI will catch broken examples on every PR, not only on full acceptance runs.

The harness still depends on the locally built provider being available to the `terraform` CLI. We will use the same `dev_overrides`-style configuration the existing development workflow already supports.

## Risks / Trade-offs

- **`terraform init` per subtest is slow.** With ~115 examples and a fresh init per directory, naive sequential execution may add several minutes to `go test`. Mitigation: run subtests in parallel (`t.Parallel()`), which is safe because each writes to its own tempdir; share the provider plugin cache across subtests via `TF_PLUGIN_CACHE_DIR`. Measure before optimizing further.
- **Self-contained example files inflate `alerting_rule` docs slightly.** Each of the three rule snippets will redefine the connector and data stream prerequisites. Mitigation: accept the duplication; it is a one-time cost confined to one resource and matches what users actually need to copy.
- **Latent broken examples may be more numerous than expected.** Mitigation: the WIP commit referenced by the task already proves the harness pattern works; we will run it locally during implementation to enumerate and fix all surfaced failures before the change is ready for review.
- **`terraform validate` may not catch every schema bug Plugin Framework providers can express.** Some validations only run during plan (e.g. `Validators` on attributes that compare across blocks). Mitigation: treat this test as a floor, not a ceiling. Per-resource acceptance tests continue to cover apply-time behaviour.
- **Provider `Configure` may still be invoked by `terraform validate` in some Terraform versions.** Mitigation: confirm during implementation that the elasticstack provider's `Configure` does not require live connectivity (most providers defer client construction). If it does, document the requirement and consider running with `TF_SKIP_PROVIDER_VERIFY` or split the harness into stack-required and stack-free buckets.

## Migration Plan

The work is naturally staged:

1. Add the embed FS in `examples/examples.go` and the validation harness in `internal/acctest/`.
2. Run the harness locally; collect the list of failing example files.
3. Restructure `examples/resources/elasticstack_kibana_alerting_rule/` for self-containment.
4. Fix `delayed_data_check_config` (#2523) and every other example surfaced as broken by step 2.
5. Regenerate provider docs only for the example files actually changed.
6. Land the harness, the targeted cleanups, and any regenerated docs in a single change.

No external data migration is required. No state-format changes. Embedded provider and connection blocks in existing examples remain valid input to the harness and are deliberately left untouched.

## Open Questions

- **Provider availability mechanism**: should the harness use `dev_overrides` (matching the local development workflow) or build the provider once at test setup and inject it via a filesystem mirror? `dev_overrides` is simpler but emits a stderr warning each invocation; a filesystem mirror is cleaner but more setup. Default plan: `dev_overrides`, revisit if the warnings make output hard to read.
- **Test location**: place the new test alongside existing acceptance tests in `internal/acctest/` for consistency, or in a new `internal/examples/` package to signal it is a different class of test (no `TF_ACC=1` gating, no live stack). Default plan: `internal/acctest/` with a clear filename like `examples_validate_test.go`; revisit if the lighter dependencies make a separate package worth the split.
- **Parallelism limit**: how aggressively to parallelise. Default plan: full `t.Parallel()` with the standard `go test -parallel` cap; tune if init contention shows up under high parallelism.
