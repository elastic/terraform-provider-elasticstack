# `elasticstack_kibana_agentbuilder_export_workflow` — Schema and Functional Requirements

Data source implementation: `internal/kibana/exportagentbuilder/workflow`

## Purpose

Define the Terraform schema and read behavior for the `elasticstack_kibana_agentbuilder_export_workflow` data source, including accepted identifier forms, space-resolution rules, stack-version gating, canonical state identity, and YAML export mapping from the Kibana Workflow API into Terraform state.

## Schema

```hcl
data "elasticstack_kibana_agentbuilder_export_workflow" "example" {
  id = <required, string> # accepts either bare workflow id or composite "<space_id>/<workflow_id>"; state is normalized to canonical composite id

  space_id = <optional, computed, string> # when omitted, read defaults to "default" unless id supplies a composite space

  workflow_id        = <computed, string>
  configuration_yaml = <computed, string, normalized YAML>
}
```

Notes:

- The data source uses the provider-level Kibana OpenAPI client only; there is no data-source-local connection override block.
- Read enforces a minimum Elastic Stack version of `9.4.0-SNAPSHOT`.
- `configuration_yaml` uses the provider's normalized YAML custom type when stored in state.

## Requirements

### Requirement: Workflow export uses the read API (REQ-001)

The data source SHALL export workflows by calling the Kibana workflow get operation for the resolved `space_id` and `workflow_id`. It SHALL populate Terraform state from the returned workflow object.

#### Scenario: Successful export

- GIVEN a workflow exists in Kibana
- WHEN the data source reads it
- THEN the provider SHALL fetch that workflow through the workflow get API and SHALL write state from the response

### Requirement: API, client, and version errors (REQ-002)

When the provider cannot obtain the Kibana OpenAPI client, the data source SHALL return an error diagnostic. It SHALL also verify that the Elastic Stack version is at least `9.4.0-SNAPSHOT`; if the version is lower, it SHALL fail with an `Unsupported server version` diagnostic. Transport errors and unexpected HTTP statuses from the workflow API SHALL be surfaced as diagnostics.

#### Scenario: Stack below minimum version

- GIVEN a target Elastic Stack version below `9.4.0-SNAPSHOT`
- WHEN the data source read runs
- THEN the read SHALL fail with an unsupported-version diagnostic before calling the workflow API

#### Scenario: Kibana client unavailable

- GIVEN the provider cannot obtain a Kibana OpenAPI client
- WHEN the data source read runs
- THEN the read SHALL fail with an error diagnostic

### Requirement: Input id forms and space resolution (REQ-003)

The data source SHALL accept the required `id` argument either as a bare workflow id or as a composite `<space_id>/<workflow_id>` string. If `id` parses as a composite id, the data source SHALL use the resource-id segment as the workflow id. When `space_id` is omitted or unknown, the data source SHALL use the composite id's space segment; when `space_id` is explicitly set in configuration, that explicit `space_id` SHALL take precedence over any space segment embedded in `id`.

#### Scenario: Composite id supplies the space

- GIVEN `id = "observability/workflow-1234"` and `space_id` is omitted
- WHEN the data source resolves the target workflow
- THEN it SHALL read workflow `workflow-1234` from space `observability`

#### Scenario: Explicit space overrides composite id space

- GIVEN `id = "observability/workflow-1234"` and `space_id = "default"`
- WHEN the data source resolves the target workflow
- THEN it SHALL read workflow `workflow-1234` from space `default`

### Requirement: Default space when none is provided (REQ-004)

When configuration does not provide a known `space_id` and the required `id` is not a composite id, the data source SHALL default the target space to `default`.

#### Scenario: Bare workflow id without space

- GIVEN `id = "workflow-1234"` and `space_id` is omitted
- WHEN the data source reads the workflow
- THEN it SHALL query space `default`

### Requirement: Canonical state identity and exported fields (REQ-005)

After a successful read, the data source SHALL store a canonical composite `id` in the format `<space_id>/<workflow_id>`. It SHALL populate `space_id` from the resolved target space, `workflow_id` from the workflow returned by Kibana, and `configuration_yaml` from the workflow's YAML definition using the normalized YAML custom type.

#### Scenario: State normalization after bare id input

- GIVEN configuration uses bare id `workflow-1234` and the workflow is read from space `default`
- WHEN Terraform state is written
- THEN `id` SHALL be `default/workflow-1234`, `space_id` SHALL be `default`, and `workflow_id` SHALL be `workflow-1234`

### Requirement: Not-found handling (REQ-006)

If the workflow get request returns not found, the data source SHALL fail with a `Workflow not found` diagnostic and SHALL NOT write successful state for that read.

#### Scenario: Missing workflow

- GIVEN the resolved workflow id and space do not exist in Kibana
- WHEN the data source read runs
- THEN the provider SHALL return a `Workflow not found` diagnostic

## Traceability

| Area | Primary files |
|------|----------------|
| Schema | `internal/kibana/exportagentbuilder/workflow/schema.go` |
| Metadata / Configure | `internal/kibana/exportagentbuilder/workflow/data_source.go` |
| Read logic | `internal/kibana/exportagentbuilder/workflow/read.go` |
| Data model | `internal/kibana/exportagentbuilder/workflow/models.go` |
| API status handling | `internal/clients/kibanaoapi/workflows.go` |
| Composite id parsing | `internal/clients/api_client.go` |
| YAML custom type | `internal/utils/customtypes/normalized_yaml_type.go`, `internal/utils/customtypes/normalized_yaml_value.go` |
