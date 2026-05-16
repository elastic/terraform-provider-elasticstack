## Context

The provider ships ~115 example `.tf` files under `examples/resources/` and `examples/data-sources/`. They are consumed two ways:

1. `terraform-plugin-docs` embeds them into the generated reference docs at `docs/resources/<name>.md` and `docs/data-sources/<name>.md`.
2. End users copy them into their own configurations as starting points.

Today there is no automated check that these snippets plan against the provider's actual schema and plan-time validation. Schema drift between the provider and its documented examples (e.g. a block-form attribute being changed to an object attribute) is silent until a user reports a broken copy-paste, as in issue #2523.

This change introduces a PlanOnly acceptance test that verifies every covered example against the in-process provider used by the existing acceptance suite. The design favours using the provider test harness already present in the repo over maintaining a separate `terraform validate` runner, provider installation path, and Terraform CLI orchestration.

## Goals / Non-Goals

**Goals**:
- Catch schema-mismatch and plan-time validation bugs in example `.tf` files (block vs. attribute, wrong attribute names, type errors, missing required attributes, invalid provider validators) before they reach users.
- Attribute failures to a specific example file so the fix is obvious.
- Run as part of the existing acceptance-test workflow using the same provider factories and prechecks as other acceptance coverage.
- Make example files self-contained so users can copy any single snippet and have it plan.
- Establish "examples must plan" as an enforceable convention going forward.

**Non-Goals**:
- Verify that examples can be applied successfully (different and much more expensive guarantee, already partially covered by per-resource acceptance tests).
- Lint examples for stylistic concerns (formatting, comment quality).
- Test combinations of examples that span multiple resource directories.

## Decisions

### 1. Use `PlanOnly` acceptance tests rather than `terraform validate`

`terraform-plugin-testing` already gives the provider a stable in-process execution path through `resource.Test`, `ProtoV6ProviderFactories`, `ConfigDirectory`, and `PlanOnly`. Using that path avoids the complexity of creating a separate `terraform validate` harness that must install or locate a locally built provider, manage `terraform init`, parse CLI diagnostics, and handle Terraform CLI warning noise.

PlanOnly also catches more of what users experience when they copy an example: provider schema checks, Plugin Framework validators, data source planning, and provider configuration all run through the same machinery as the rest of the acceptance suite. The trade-off is that this is acceptance coverage, not a stack-free unit test. Examples that include data sources or provider interactions may require the live Elasticsearch/Kibana endpoints already required by `acctest.PreCheck`.

The bug class motivating this change (#2523, block vs. attribute) is caught before planning can succeed, and additional plan-time validation failures are caught as well.

### 2. Drive the harness via `terraform-plugin-testing`

The harness will:
1. Embed the contents of `examples/resources/` and `examples/data-sources/` via `embed.FS` in `examples/examples.go`.
2. For each example `.tf` file, write it into an isolated per-subtest module directory under `testdata/<test name>/plan/` — where `<test name>` is the full `testing.T` name (`t.Name()`, including `TestAccExamples_planOnly` and `/`-separated subtest segments — see `acctest.NamedTestCaseDirectory("plan")`) — matching the repo’s acceptance-test directory pattern enforced by `acctestconfigdirlint`.
3. Run `resource.Test` with `ProtoV6ProviderFactories: acctest.Providers`, `ConfigDirectory: acctest.NamedTestCaseDirectory("plan")`, and `PlanOnly: true`.
4. Set `ExpectNonEmptyPlan: true` for every file under `examples/resources/`. For files under `examples/data-sources/`, set `ExpectNonEmptyPlan: true` only when HCL parsing of the root body finds a top-level `resource` or `output` block (supporting managed resources or outputs in the snippet); otherwise use `false` so read-only plans are not rejected solely for being empty. This aligns with `terraform-plugin-testing`, which compares `ExpectNonEmptyPlan` to both the non-refresh and refresh plans in a PlanOnly step.
5. Use `t.Run("<path-under-examples/>", ...)` so the failing file is named in the test output and matches `-run` filters.

### 3. One subtest per example file

The user explicitly preferred per-file failure attribution. To make per-file isolation work, every example `.tf` file must be **self-contained**: it cannot reference resources or data sources defined only in sibling files. This is also a usability improvement — readers of the generated docs see one snippet at a time and reasonably expect each to be runnable on its own.

Today, only one directory violates this: `examples/resources/elasticstack_kibana_alerting_rule/`. Its `resource-index-rule.tf` and `resource_rule_action_frequency.tf` both reference `elasticstack_kibana_action_connector.index_example` and `elasticstack_elasticsearch_data_stream.my_data_stream` defined in the directory's `resource.tf`. We will inline those dependencies into each file. The cost is some duplication in the generated docs; the benefit is that every documented snippet is independently usable.

### 4. Leave embedded provider configuration in examples alone

An earlier version of this design considered stripping every embedded `provider "elasticstack" { ... }` block and per-resource `elasticsearch_connection { ... }` block from existing examples. That cleanup is unnecessary for the PlanOnly harness: the acceptance-test provider configuration path already supplies the environment needed by examples, and examples that intentionally show explicit provider or connection blocks should continue to do so in the generated docs.

This change therefore leaves embedded provider and connection blocks in place. If a future change wants to standardise example docs (e.g. for stylistic consistency), it can do so as a separate, focused cleanup with its own justification.

### 5. Static skip-list for non-covered directories and enumerated per-file snippets

Two **directory prefixes** under `examples/` are not traversed by the embedded example trees (`examples/cloud/`, `examples/provider/`); the harness rejects them similarly by prefix (`skippedExamplePathPrefixes`). Out of scope patterns:

- **`examples/cloud/`**: uses the `ec` (Elastic Cloud) provider, not `elasticstack`.
- **`examples/provider/`**: partial provider-configuration snippets rather than runnable roots.

Beyond prefixes, **`planOnlySkippedEmbedPaths`** holds a **minimal** allowlist for individual `examples/resources/**/*.tf` and `examples/data-sources/**/*.tf` snippets that legitimately resist single-module Plugin Framework harness planning despite being checked into `examples/`. Each entry names the embed path and states why (see nearby comments in `examples_plan_test.go`). Current documented cases:

- **`data-sources/elasticstack_kibana_agentbuilder_agent/import.tf`** — consumes `terraform_remote_state` referencing another Terraform root; isolation would require scaffolding a second workspace.
- **`resources/elasticstack_elasticsearch_security_api_key/rotation.tf`** — installs **`hashicorp/time`** alongside `elasticstack`; the harness exposes only compiled `elasticstack` factories (`ProtoV6ProviderFactories`).
- **`data-sources/elasticstack_fleet_enrollment_tokens/data-source.tf`** — Fleet enrollment token reads depend on an agent **`policy_id` UUID that exists in that target stack**. Matrix acceptance stacks expose no common UUID; example IDs return HTTP 404 in CI despite valid provider schema.

Introducing a new enumerated skip **still** demands a reviewer-visible code edit plus rationale; sentinel comments alone are intentionally **not** parsed to avoid unmanaged sprawl.

We considered sentinel comments (`# acctest:skip reason=`) for richer UX, but enumerated inline constants keep policy grep-friendly and deterministic for CI auditors.

### 6. Fix every example the harness flags

Rather than landing the test in a partially-disabled state, this change fixes every example that the new harness reports as broken. We expect the `delayed_data_check_config` bug from #2523 plus a small number of latent issues — most likely a handful of block-vs-attribute or attribute-rename mistakes that have accumulated since the relevant resources were last touched.

### 7. Run with the acceptance-test suite

Because PlanOnly may evaluate data sources and provider plan validation that depends on live services, this test belongs in the acceptance-test suite and should use `acctest.PreCheck(t)`. CI will catch broken examples wherever acceptance tests run with the standard Elasticsearch/Kibana environment.

The harness uses the in-process provider factories from `acctest.Providers`, so it does not need `terraform init`, `dev_overrides`, a filesystem mirror, or a locally installed provider binary.

### 8. Bounded concurrency inside `TestAccExamples_planOnly`

Subtests remain parallel (`t.Parallel()`), but simultaneous PlanOnly runs across ~100+ Terraform roots are gated with a semaphore (currently **four** concurrent example plans — see `maxConcurrentExamplesPlanHarness` in `internal/acctest/examples_plan_test.go`). Full unbounded parallelism correlated with flaky refresh-phase failures complaining that the Elasticsearch client was unset despite `ELASTICSEARCH_ENDPOINTS`; a higher cap (16) still saw occasional recurrence on repeated **full-package** reruns (`-count` ≥ 2). The behaviour appears tied to concurrency stress across many independent `resource.Test`/Terraform runs against the muxed in-process provider, not to bad example HCL once Elasticsearch-only provider patterns are correct. Lower throughput is an acceptable trade-off for stable CI; raise the constant only with evidence that upstream fixes or faster stacks remove the flake.

## Risks / Trade-offs

- **PlanOnly may be slower or flakier than schema-only validation.** Some examples, especially data sources, may reach the live stack during planning. Mitigation: run under existing acceptance-test prechecks, keep each example isolated, and use normal acceptance-test CI expectations for service availability.
- **Self-contained example files inflate `alerting_rule` docs slightly.** Each of the three rule snippets will redefine the connector and data stream prerequisites. Mitigation: accept the duplication; it is a one-time cost confined to one resource and matches what users actually need to copy.
- **Latent broken examples may be more numerous than expected.** Mitigation: the WIP commit referenced by the task already proves the harness pattern works; we will run it locally during implementation to enumerate and fix all surfaced failures before the change is ready for review.
- **PlanOnly does not prove apply succeeds.** The test stops before mutating the stack. Mitigation: treat this test as a documentation-example floor; per-resource acceptance tests continue to cover apply, update, read, import, and destroy behaviour.
- **Data-source examples may need real backing objects.** Some data sources cannot plan successfully unless referenced objects exist. Mitigation: either make those examples self-contained by defining their prerequisites in the same file, or add narrowly documented skips only when a prerequisite cannot be created without leaving PlanOnly scope.

## Migration Plan

The work is naturally staged:

1. Add the embed FS in `examples/examples.go` and the PlanOnly acceptance harness in `internal/acctest/`.
2. Run the harness locally; collect the list of failing example files.
3. Restructure `examples/resources/elasticstack_kibana_alerting_rule/` for self-containment.
4. Fix `delayed_data_check_config` (#2523) and every other example surfaced as broken by step 2.
5. Regenerate provider docs only for the example files actually changed.
6. Land the harness, the targeted cleanups, and any regenerated docs in a single change.

No external data migration is required. No state-format changes. Embedded provider and connection blocks in existing examples remain valid input to the harness and are deliberately left untouched.

## Task 4 — documentation, CI alignment, and harness timing

**CI:** The workflow in `.github/workflows/test.yml` (“Matrix Acceptance Test” job) starts a Docker Elastic Stack, then runs `make testacc` with `TF_ACC: "1"` and the usual `ELASTICSEARCH_ENDPOINTS`, Elasticsearch credentials, and `KIBANA_ENDPOINT`/`KIBANA_PASSWORD` (same inputs as the rest of the acceptance suite). That target uses `gotestsum` with `--packages="./..."`. When **provider-impacting** changes trigger the matrix, `TestAccExamples_planOnly` in `internal/acctest/` runs with the full suite. Pull requests whose diffs are classified as **OpenSpec-only** may **skip** the acceptance job entirely (workflow change classification); in that case the examples harness does not run because the job is skipped, not because the test is excluded from the package list.

**Harness wall time:** On a localhost Elastic Stack, `go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1` measured **21.790s** Go package time on the final `ok` line and **23.62s** wall-clock elapsed (`time` real). Throughput is intentionally capped by **`maxConcurrentExamplesPlanHarness = 4`** (bounded semaphore), not by `t.Parallel()` alone (see Decision 8). Re-run or compare hosts with:

`time env TF_ACC=1 ELASTICSEARCH_ENDPOINTS=… KIBANA_ENDPOINT=… ELASTICSEARCH_USERNAME=… ELASTICSEARCH_PASSWORD=… go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1`

## Open Questions

- **Resource-vs-data-source plan expectations**: resolved in implementation — `ExpectNonEmptyPlan: true` for all `examples/resources/` files; for `examples/data-sources/`, `true` when the root HCL body declares a `resource` or `output` block, else `false`, matching `terraform-plugin-testing` PlanOnly semantics.
- **Test location**: place the new test alongside existing acceptance-test helpers in `internal/acctest/`, or in a new package if importing `examples` creates an undesirable dependency direction. Default plan: `internal/acctest/` with a clear filename like `examples_plan_test.go`.
- **Parallelism limit**: resolved — retain `t.Parallel()` with a bounded semaphore for concurrent PlanOnly executions (Decision 8). Tune `maxConcurrentExamplesPlanHarness` if stacks or mux behaviour warrant it.
