## ADDED Requirements

### Requirement: Every example file SHALL plan against the provider (REQ-001)

For every `*.tf` file under `examples/resources/` and `examples/data-sources/` that is not in the skip-list defined by this capability, the project SHALL provide an automated acceptance test that runs that file in isolation through a `terraform-plugin-testing` `PlanOnly` step against the in-process `elasticstack` provider.

The test SHALL fail when planning reports any diagnostic of severity `error` for a covered file. The test SHALL surface the offending file path and plan diagnostic text in the failure output.

#### Scenario: Plan-conformant example passes

- **GIVEN** an example file whose configuration matches the current `elasticstack` provider schema and plan-time validation rules
- **WHEN** the PlanOnly harness runs against that file
- **THEN** the corresponding subtest SHALL pass with no error diagnostics

#### Scenario: Block-vs-attribute mismatch is rejected

- **GIVEN** an example file that configures a schema attribute using HCL block syntax (or vice versa) — for example `delayed_data_check_config { ... }` where the schema declares an attribute
- **WHEN** the PlanOnly harness runs against that file
- **THEN** the corresponding subtest SHALL fail and the failure message SHALL include the planning diagnostic identifying the offending construct

#### Scenario: Unknown attribute name is rejected

- **GIVEN** an example file that uses an attribute name that does not exist on the referenced resource or data source
- **WHEN** the PlanOnly harness runs against that file
- **THEN** the corresponding subtest SHALL fail with the unknown-attribute diagnostic from planning

### Requirement: Example failures SHALL be attributed to a single file (REQ-002)

The PlanOnly harness SHALL run each covered `.tf` file in its own subtest, named so that a `go test -run` filter can target an individual example. The subtest name SHALL include the file's path relative to the `examples/` root.

#### Scenario: Subtest name identifies the example

- **WHEN** an example at `examples/resources/elasticstack_elasticsearch_ml_datafeed/resource.tf` fails planning
- **THEN** the failing subtest name SHALL include `elasticsearch_ml_datafeed/resource.tf` (or the equivalent relative path) so the failing example is identifiable from CI output alone

#### Scenario: Failures do not cascade across files

- **GIVEN** two example files in the same directory, only one of which is broken
- **WHEN** the PlanOnly harness runs both
- **THEN** only the subtest for the broken file SHALL fail; the subtest for the well-formed file SHALL pass

### Requirement: Example files SHALL be self-contained (REQ-003)

Each `*.tf` file under `examples/resources/` and `examples/data-sources/` SHALL plan in isolation. An example file SHALL NOT depend on resources, data sources, locals, or variables that are defined only in sibling files within the same directory.

This requirement is enforced by REQ-001 and REQ-002: the harness plans each file independently, so any cross-file reference produces an unresolved-reference error.

#### Scenario: Cross-file references are rejected

- **GIVEN** an example file that references `elasticstack_kibana_action_connector.shared` defined only in a sibling file
- **WHEN** the PlanOnly harness runs against that file in isolation
- **THEN** the subtest SHALL fail with an unresolved-reference diagnostic

#### Scenario: Inlined dependencies pass

- **GIVEN** an example file that references its own definitions of every resource and data source it depends on
- **WHEN** the PlanOnly harness runs against it in isolation
- **THEN** the subtest SHALL pass

### Requirement: The harness SHALL use acceptance-test provider execution (REQ-004)

The PlanOnly harness SHALL use `terraform-plugin-testing` with `resource.Test`, `ProtoV6ProviderFactories: acctest.Providers`, and `PlanOnly: true`. It SHALL NOT shell out to `terraform validate`, manage a Terraform CLI provider installation, or require `dev_overrides`.

`terraform-plugin-testing` applies `ExpectNonEmptyPlan` to PlanOnly steps against both the non-refresh plan and the follow-up refresh plan: a mismatch fails the step. The harness therefore chooses `ExpectNonEmptyPlan` so that success is allowed whenever the plan can legitimately be non-empty or empty under that model.

For covered files under `examples/resources/`, the harness SHALL set `ExpectNonEmptyPlan: true` (resource-documentation examples normally plan creates).

For covered files under `examples/data-sources/`, the harness SHALL set `ExpectNonEmptyPlan: true` when the root module body, as parsed by HCL, declares a top-level `resource` or `output` block (supporting managed resources or outputs in the same file). Otherwise it SHALL set `ExpectNonEmptyPlan: false` so read-only data-source examples whose plans are empty are not rejected solely for that reason.

#### Scenario: Resource example may produce creates

- **GIVEN** a resource example that plans one or more managed resources
- **WHEN** the PlanOnly harness runs against that file
- **THEN** the subtest SHALL pass if planning succeeds and `ExpectNonEmptyPlan` is satisfied (non-empty plans allowed)

#### Scenario: Data-source example with only reads may produce an empty plan

- **GIVEN** a data-source example whose root module declares no `resource` or `output` blocks (only data sources, provider configuration, etc.)
- **WHEN** the PlanOnly harness runs against that file
- **THEN** the subtest SHALL pass if planning succeeds and the plan is empty, with `ExpectNonEmptyPlan: false`

#### Scenario: Data-source example with supporting resources or outputs may produce a non-empty plan

- **GIVEN** a data-source example whose root module declares at least one top-level `resource` or `output` block in the same file
- **WHEN** the PlanOnly harness runs against that file
- **THEN** the subtest SHALL pass if planning succeeds and `ExpectNonEmptyPlan: true` is satisfied (non-empty plans allowed)

### Requirement: The harness SHALL skip non-covered example directories (REQ-005)

The PlanOnly harness SHALL exclude `examples/cloud/` and `examples/provider/` from coverage. The skip-list SHALL be expressed as repository-relative path prefixes in the harness source code and SHALL be documented inline.

The harness SHALL NOT exclude any other directory by default. New skips SHALL require explicit code changes to the harness, with a documented justification, rather than being controlled by data files or environment variables.

#### Scenario: `examples/cloud/` is not planned

- **WHEN** the PlanOnly harness walks `examples/`
- **THEN** files under `examples/cloud/` SHALL NOT produce subtests

#### Scenario: `examples/provider/` is not planned

- **WHEN** the PlanOnly harness walks `examples/`
- **THEN** files under `examples/provider/` SHALL NOT produce subtests

#### Scenario: Other directories are not implicitly skipped

- **GIVEN** a new `*.tf` file added under any path beneath `examples/resources/` or `examples/data-sources/`
- **WHEN** the PlanOnly harness next runs
- **THEN** that file SHALL be covered by a subtest without further configuration

### Requirement: The harness SHALL run as acceptance coverage (REQ-006)

The PlanOnly harness SHALL run under the existing acceptance-test workflow. It SHALL require the same acceptance-test preconditions as other provider acceptance tests, including `TF_ACC=1` and the live Elastic Stack environment variables validated by `acctest.PreCheck(t)`.

The harness SHALL NOT apply resources, update resources, import resources, or destroy resources. It SHALL stop after a successful plan for each covered example.

#### Scenario: Test requires acceptance environment

- **GIVEN** a host with no Elastic stack reachable and no acceptance-test environment variables set
- **WHEN** the examples PlanOnly acceptance test is invoked
- **THEN** the harness SHALL fail or skip according to the standard acceptance-test precheck behavior

#### Scenario: Test runs without mutating resources

- **GIVEN** a CI run with `TF_ACC=1` and a live stack configured
- **WHEN** the examples PlanOnly acceptance test runs
- **THEN** the harness SHALL execute exactly once per covered example and SHALL NOT apply, update, import, or destroy managed resources
