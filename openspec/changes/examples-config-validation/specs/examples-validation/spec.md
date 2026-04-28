## ADDED Requirements

### Requirement: Every example file SHALL validate against the provider schema (REQ-001)

For every `*.tf` file under `examples/resources/` and `examples/data-sources/` that is not in the skip-list defined by this capability, the project SHALL provide an automated test that runs `terraform init` and `terraform validate` against the locally built `elasticstack` provider for that file in isolation.

The test SHALL fail when `terraform validate` reports any diagnostic of severity `error` for a covered file. The test SHALL surface the offending file path, validate diagnostic text, and source location in the failure message.

#### Scenario: Schema-conformant example passes

- **GIVEN** an example file whose configuration matches the current `elasticstack` provider schema
- **WHEN** the validation harness runs against that file
- **THEN** the corresponding subtest SHALL pass with no error diagnostics

#### Scenario: Block-vs-attribute mismatch is rejected

- **GIVEN** an example file that configures a schema attribute using HCL block syntax (or vice versa) — for example `delayed_data_check_config { ... }` where the schema declares an attribute
- **WHEN** the validation harness runs against that file
- **THEN** the corresponding subtest SHALL fail and the failure message SHALL include the validate diagnostic identifying the offending construct

#### Scenario: Unknown attribute name is rejected

- **GIVEN** an example file that uses an attribute name that does not exist on the referenced resource or data source
- **WHEN** the validation harness runs against that file
- **THEN** the corresponding subtest SHALL fail with the unknown-attribute diagnostic from `terraform validate`

### Requirement: Example failures SHALL be attributed to a single file (REQ-002)

The validation harness SHALL run each covered `.tf` file in its own subtest, named so that a `go test -run` filter can target an individual example. The subtest name SHALL include the file's path relative to the `examples/` root.

#### Scenario: Subtest name identifies the example

- **WHEN** an example at `examples/resources/elasticstack_elasticsearch_ml_datafeed/resource.tf` fails validation
- **THEN** the failing subtest name SHALL include `elasticsearch_ml_datafeed/resource.tf` (or the equivalent relative path) so the failing example is identifiable from CI output alone

#### Scenario: Failures do not cascade across files

- **GIVEN** two example files in the same directory, only one of which is broken
- **WHEN** the harness runs both
- **THEN** only the subtest for the broken file SHALL fail; the subtest for the well-formed file SHALL pass

### Requirement: Example files SHALL be self-contained (REQ-003)

Each `*.tf` file under `examples/resources/` and `examples/data-sources/` SHALL be valid in isolation. An example file SHALL NOT depend on resources, data sources, locals, or variables that are defined only in sibling files within the same directory.

This requirement is enforced by REQ-001 and REQ-002: the harness validates each file independently, so any cross-file reference produces an unresolved-reference error.

#### Scenario: Cross-file references are rejected

- **GIVEN** an example file that references `elasticstack_kibana_action_connector.shared` defined only in a sibling file
- **WHEN** the validation harness runs against that file in isolation
- **THEN** the subtest SHALL fail with an unresolved-reference diagnostic

#### Scenario: Inlined dependencies pass

- **GIVEN** an example file that references its own definitions of every resource and data source it depends on
- **WHEN** the validation harness runs against it in isolation
- **THEN** the subtest SHALL pass

### Requirement: The harness SHALL skip non-validatable example directories (REQ-005)

The validation harness SHALL exclude `examples/cloud/` and `examples/provider/` from coverage. The skip-list SHALL be expressed as repository-relative path prefixes in the harness source code and SHALL be documented inline.

The harness SHALL NOT exclude any other directory by default. New skips SHALL require explicit code changes to the harness, with a documented justification, rather than being controlled by data files or environment variables.

#### Scenario: `examples/cloud/` is not validated

- **WHEN** the validation harness walks `examples/`
- **THEN** files under `examples/cloud/` SHALL NOT produce subtests

#### Scenario: `examples/provider/` is not validated

- **WHEN** the validation harness walks `examples/`
- **THEN** files under `examples/provider/` SHALL NOT produce subtests

#### Scenario: Other directories are not implicitly skipped

- **GIVEN** a new `*.tf` file added under any path beneath `examples/resources/` or `examples/data-sources/`
- **WHEN** the validation harness next runs
- **THEN** that file SHALL be covered by a subtest without further configuration

### Requirement: The harness SHALL run without a live Elastic stack (REQ-006)

The validation harness SHALL be executable as part of the standard `go test ./...` invocation. It SHALL NOT require `TF_ACC=1`, `ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD`, `ELASTICSEARCH_API_KEY`, `KIBANA_ENDPOINT`, `KIBANA_USERNAME`, `KIBANA_PASSWORD`, `KIBANA_API_KEY`, or any other live-stack environment variable consumed by the provider's acceptance-test harness. It SHALL NOT depend on Elasticsearch or Kibana being reachable from the host running the test.

#### Scenario: Test runs in a clean environment

- **GIVEN** a host with no Elastic stack reachable and no acceptance-test environment variables set
- **WHEN** `go test ./...` runs
- **THEN** the validation harness SHALL execute and SHALL pass for all covered, well-formed example files

#### Scenario: Test runs alongside acceptance tests

- **GIVEN** a CI run with `TF_ACC=1` and a live stack configured
- **WHEN** the full test suite runs
- **THEN** the validation harness SHALL still execute exactly once per covered example and SHALL NOT additionally call data sources or apply resources
