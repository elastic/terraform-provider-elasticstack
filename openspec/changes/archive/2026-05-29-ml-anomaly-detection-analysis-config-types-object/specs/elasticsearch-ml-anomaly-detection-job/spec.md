# `elasticstack_elasticsearch_ml_anomaly_detection_job` — Fix Value Conversion Error for variable-sourced `analysis_config`

Resource implementation: `internal/elasticsearch/ml/anomalydetectionjob/`

## Purpose

Eliminate the `Value Conversion Error` crash that occurs when `analysis_config` is assigned from a
dynamic source (e.g. `for_each` / `each.value` or a Terraform variable typed as `object({...})`).
The fix changes the internal Go representation from struct-pointers (which cannot hold unknown
values) to `types.Object` (which can), following the established pattern in this resource.

## ADDED Requirements

### Requirement: TFModel.AnalysisConfig uses a framework-native type (REQ-036)
`TFModel.AnalysisConfig` SHALL be typed as `types.Object` so the Plugin Framework can store null,
unknown, and known values for `analysis_config` without panicking during plan evaluation.

Previously `TFModel.AnalysisConfig` was `*AnalysisConfigTFModel`, which cannot hold an unknown
value and produced the error reported in #3403.

#### Scenario: Variable-sourced analysis_config does not produce a Value Conversion Error
- **GIVEN** a resource configuration that assigns `analysis_config = var.analysis_config` where
  `var.analysis_config` is typed as `object({...})` matching the full `analysis_config` schema
- **WHEN** `terraform plan` is run
- **THEN** the plan SHALL succeed without a `Value Conversion Error` at `analysis_config`

#### Scenario: analysis_config sourced from for_each does not produce a Value Conversion Error
- **GIVEN** a resource using `for_each` where `analysis_config = each.value.job.analysis_config`
  is set from a decoded JSON or map value
- **WHEN** `terraform plan` is run
- **THEN** the plan SHALL succeed without a `Value Conversion Error` at `analysis_config`

#### Scenario: Null analysis_config in imported state is handled correctly
- **GIVEN** a resource state where `analysis_config` is null (e.g. after import before the first
  read fills it in)
- **WHEN** the provider reads the resource
- **THEN** the read SHALL not panic and SHALL produce a valid state

#### Scenario: analysis_config round-trips correctly through plan and apply
- **GIVEN** a standard resource configuration with `analysis_config` set directly (not via variable)
- **WHEN** `terraform apply` is run and then `terraform plan` is re-run
- **THEN** the second plan SHALL show no diff and the resource SHALL NOT be replaced

### Requirement: AnalysisConfigTFModel.PerPartitionCategorization uses a framework-native type (REQ-037)
`AnalysisConfigTFModel.PerPartitionCategorization` SHALL be typed as `types.Object` so the Plugin
Framework can store null, unknown, and known values for `per_partition_categorization` without
panicking.

Previously `AnalysisConfigTFModel.PerPartitionCategorization` was
`*PerPartitionCategorizationTFModel`, which carries the same struct-pointer limitation and would
produce the same class of error if `per_partition_categorization` were sourced from a variable.

#### Scenario: Variable-sourced per_partition_categorization does not produce a Value Conversion Error
- **GIVEN** a resource configuration that assigns `analysis_config.per_partition_categorization`
  from a variable or dynamic expression
- **WHEN** `terraform plan` is run
- **THEN** the plan SHALL succeed without a `Value Conversion Error` at
  `analysis_config.per_partition_categorization`

#### Scenario: Absent per_partition_categorization is stored as null
- **GIVEN** a resource where the Elasticsearch API returns `per_partition_categorization` with
  `enabled = false` and the user did not configure it
- **WHEN** the provider reads the resource
- **THEN** `analysis_config.per_partition_categorization` SHALL be `null` in Terraform state,
  matching prior behaviour

#### Scenario: Configured per_partition_categorization round-trips correctly
- **GIVEN** a resource configuration with `per_partition_categorization { enabled = true }`
- **WHEN** `terraform apply` is run and then `terraform plan` is re-run
- **THEN** the second plan SHALL show no diff for `per_partition_categorization`

### Requirement: Schema and user-visible interface are unchanged (REQ-038)
The Terraform schema for `analysis_config` and `per_partition_categorization` SHALL remain
unchanged. No attribute names, types, optionality, computability, or plan modifiers SHALL be
added, removed, or altered as part of this fix.

#### Scenario: Existing configurations continue to work
- **GIVEN** a Terraform configuration that was valid with provider v0.14.x
- **WHEN** the provider is upgraded to the fixed version
- **THEN** the configuration SHALL plan and apply without errors or unexpected diffs

### Requirement: Regression test for variable-sourced analysis_config (REQ-039)
A regression acceptance test SHALL be added that assigns the full `analysis_config` block from a
Terraform variable typed as `object({...})`, runs plan+apply, and then re-runs plan to confirm the
read path does not produce a `Value Conversion Error`.

#### Scenario: TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig — create
- **GIVEN** a test step where `analysis_config` is assigned from a default-valued `object` variable
- **WHEN** the test runs plan and apply
- **THEN** the resource SHALL be created successfully and `job_id` SHALL match the input

#### Scenario: TestAccResourceAnomalyDetectionJobVariableSourcedAnalysisConfig — plan-after-apply
- **GIVEN** the resource created in the previous step with the same variable-sourced config
- **WHEN** a second plan is run
- **THEN** the plan SHALL produce no diff and no `Value Conversion Error`
