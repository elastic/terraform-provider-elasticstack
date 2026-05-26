# Security Roles Guide

Guide implementation: `templates/guides/security-roles.md.tmpl`
Resource documentation implementation: `templates/resources/kibana_security_role.md.tmpl`, `templates/resources.md.tmpl`

## Purpose

Define the canonical requirements for the provider security roles guide, including guide generation, resource-doc linking, scenario-based examples, privilege-model explanations, embedded Terraform examples, Kibana feature reference coverage, API key composition guidance, and field/document security examples.

## Requirements

### Requirement: Guide page exists and is linked from resource docs

A standalone provider guide SHALL exist at `templates/guides/security-roles.md.tmpl` and be rendered to `docs/guides/security-roles.md` by `make docs-generate`.

Both the Kibana security role resource template (`templates/resources/kibana_security_role.md.tmpl`) and the Elasticsearch security role resource documentation source (`templates/resources.md.tmpl` for the generic fallback template path used by this resource) SHALL include a "See also" link to the guide immediately after the resource description and before the Example Usage section.

#### Scenario: Guide renders without error

- **WHEN** `make docs-generate` is run
- **THEN** `docs/guides/security-roles.md` is created with no generation errors

#### Scenario: Resource pages link to the guide

- **WHEN** the rendered `docs/resources/kibana_security_role.md` is read
- **THEN** it contains a link to the security roles guide page

#### Scenario: make check-docs passes

- **WHEN** the guide template and all referenced example files are committed
- **THEN** `make check-docs` exits with code 0

---

### Requirement: Existing generic examples are replaced by scenario-based examples

The existing example files `resource-with-base.tf`, `resource-with-feature.tf` (for `elasticstack_kibana_security_role`) and `resource.tf` (for `elasticstack_elasticsearch_security_role`) SHALL be removed and replaced with scenario-named files that demonstrate realistic, least-privilege configurations.

The replacement example files SHALL cover the following archetypes, with one `.tf` file per archetype:

- **data-analyst**: read-only access to Discover and Dashboards in a named Kibana space, with read + `view_index_metadata` on the relevant index patterns
- **data-ingest**: write access to data streams with `create_index` and `auto_configure` index privileges; cluster privileges limited to what ingest pipelines and index templates require
- **security-analyst**: full access to SIEM, alerting, cases, and osquery features in the security Kibana space
- **devops-readonly**: read access to Fleet, APM, and infrastructure monitoring features; `monitor` cluster privilege only
- **multi-space**: a single role granting `base = ["all"]` in dev/staging spaces and feature-level read-only in a prod space via two separate `kibana {}` blocks

Each example SHALL use realistic index pattern names (e.g. `logs-*`, `metrics-*`) rather than placeholder values.

#### Scenario: Scenario examples replace generic examples

- **WHEN** `ls examples/resources/elasticstack_kibana_security_role/` is run
- **THEN** files named `resource-with-base.tf` and `resource-with-feature.tf` do not exist
- **THEN** files named `resource-data-analyst.tf`, `resource-data-ingest.tf`, `resource-security-analyst.tf`, `resource-devops-readonly.tf`, and `resource-multi-space.tf` exist

#### Scenario: ES role examples are replaced

- **WHEN** `ls examples/resources/elasticstack_elasticsearch_security_role/` is run
- **THEN** a file named `resource.tf` with generic placeholder content does not exist
- **THEN** scenario-named example files exist

#### Scenario: All example files are valid Terraform

- **WHEN** `terraform validate` is run against each example file
- **THEN** no validation errors are reported

---

### Requirement: Guide explains the privilege model

The guide SHALL include a section explaining the conceptual distinction between `elasticstack_elasticsearch_security_role` and `elasticstack_kibana_security_role`, specifically:

- ES roles control cluster and index access and are used for ES-native access and API keys
- Kibana roles wrap the Kibana Role Management API and control both Kibana application access (spaces, features) and carry an `elasticsearch {}` block for index/cluster privileges
- The `role_descriptors` field on API keys uses the same structure as ES roles and further restricts (never expands) the owning user's privileges

#### Scenario: Guide contains privilege model explanation

- **WHEN** the guide page is read
- **THEN** it contains a section explaining when to use each resource type
- **THEN** it explains the relationship between `role_descriptors` and security roles

---

### Requirement: Guide embeds scenario examples via tffile directives

The guide SHALL embed each scenario example using `{{ tffile "examples/resources/..." }}` directives rather than duplicating the Terraform code inline. This ensures the guide always reflects the committed example files. Small inline explanatory snippets are acceptable only when no provider example file is appropriate; complete reusable Terraform examples SHALL live under `examples/resources/...` and be embedded with `tffile`.

#### Scenario: Guide embeds example files

- **WHEN** the guide template is read
- **THEN** it contains `{{ tffile` directives referencing the scenario example files
- **THEN** no Terraform code blocks are written directly inline in the guide template

---

### Requirement: Guide includes a Kibana feature privilege reference table

The guide SHALL include a reference table of commonly-used Kibana feature names and their available privilege strings. The table SHALL cover at minimum: `discover`, `dashboard`, `visualize`, `ml`, `apm`, `fleet`, `siem`, `securitySolutionCases`, `observabilityCases`, `osquery`, `rulesSettings`, `actions`, `alerting`, `canvas`, `maps`, `infrastructure`. Where Kibana exposes versioned or alias feature IDs in some stack versions, the guide MAY document the canonical/common feature names and SHALL note that deployments can return versioned or alias IDs from `GET /api/features`.

The table SHALL include a note that it covers commonly-used features, is not exhaustive, and that `GET /api/features` returns the full list for a given deployment.

#### Scenario: Table lists feature names and privilege values

- **WHEN** the guide is read
- **THEN** it contains a markdown table with columns for feature name and available privileges
- **THEN** the `discover` feature row lists sub-privileges including `minimal_read`, `url_create`, `store_search_session`

#### Scenario: Table includes exhaustiveness caveat

- **WHEN** the guide is read
- **THEN** it contains a note directing users to `GET /api/features` for a complete list

---

### Requirement: Guide includes multi-resource composition example

The guide SHALL include a complete example showing an API key with `role_descriptors` scoped to a subset of a Kibana security role's privileges, demonstrating that `role_descriptors` further restricts rather than expands.

#### Scenario: Multi-resource example is present

- **WHEN** the guide is read
- **THEN** it contains an example combining `elasticstack_kibana_security_role` and `elasticstack_elasticsearch_security_api_key` with `role_descriptors`

---

### Requirement: Guide includes field security and document-level security examples

The guide SHALL include examples demonstrating `field_security` (grant/except) and document-level security (`query`) within index permissions, explaining the use cases for each.

#### Scenario: Field security example is present

- **WHEN** the guide is read
- **THEN** it contains an example using `field_security` with a `grant` list and an explanation of PII redaction as a use case

#### Scenario: Document security example is present

- **WHEN** the guide is read
- **THEN** it contains an example using `query` with a `jsonencode` expression and an explanation of tenant isolation as a use case
