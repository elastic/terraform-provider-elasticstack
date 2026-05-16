## ADDED Requirements

### Requirement: `analysis_config.detectors` internal model type (REQ-033)

The Go field `AnalysisConfigTFModel.Detectors` SHALL be typed `types.List` (not
`[]DetectorTFModel`). The `types.List` type is the Plugin Framework-idiomatic holder
for list attributes that can carry null/unknown state, consistent with
`AnalysisConfigTFModel.CategorizationFilters` and `AnalysisConfigTFModel.Influencers`
in the same struct.

All code that reads `Detectors` as a Go slice SHALL use `ElementsAs(ctx,
&detectorSlice, false)` to convert the `types.List` value. All code that writes
`Detectors` from a `[]DetectorTFModel` slice SHALL use
`types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getDetectorAttrTypes(ctx)},
slice)`.

The Terraform schema declaration (`schema.ListNestedAttribute` for
`analysis_config.detectors`) SHALL NOT change. This requirement applies only to the
internal Go model field type.

#### Scenario: Variable-sourced detectors plan succeeds

- GIVEN a Terraform configuration that assigns `analysis_config.detectors` from a
  `variable` of type `list(object({...}))` (including when the variable has a concrete
  default value that Terraform marks as potentially-unknown at the config phase)
- WHEN `terraform plan` runs
- THEN the provider SHALL NOT return a `Value Conversion Error` and the plan SHALL
  succeed

#### Scenario: Hardcoded detectors continue to work

- GIVEN a Terraform configuration with `analysis_config.detectors` defined as a
  hardcoded list literal in the resource block
- WHEN `terraform plan` and `terraform apply` run
- THEN all existing create, update, read, and delete behaviors SHALL be preserved
  without regression

### Requirement: `ValidateConfig` plan-time unknown guard for detectors (REQ-034)

The `ValidateConfig` implementation SHALL handle the case where
`analysis_config.detectors` is unknown at plan time (for example, when the value flows
from a module input or variable). Specifically:

- When `Detectors.IsUnknown()` is true, validation of `custom_rules` SHALL be skipped
  without returning an error diagnostic. Validation is deferred to apply time when all
  values are concrete.
- When `Detectors` is known and non-null, validation of `custom_rules` SHALL proceed as
  before: each rule that has neither a non-empty `scope` nor at least one `conditions`
  entry SHALL produce an attribute-level error diagnostic.

#### Scenario: Unknown detectors skips custom_rules validation

- GIVEN an `analysis_config.detectors` that is unknown at plan time
- WHEN `ValidateConfig` runs
- THEN no `Invalid detector "custom_rules" entry` error SHALL be emitted and the plan
  SHALL proceed

#### Scenario: Known detectors with empty custom_rules still fails validation

- GIVEN a known `analysis_config.detectors` containing a custom rule with neither a
  `scope` nor any `conditions`
- WHEN `ValidateConfig` runs
- THEN the resource SHALL return an error diagnostic identifying the offending custom
  rule

### Requirement: Acceptance test for variable-sourced detectors (REQ-035)

The acceptance test suite SHALL include at least one test case for
`elasticstack_elasticsearch_ml_anomaly_detection_job` that exercises assigning
`analysis_config { detectors = var.detectors }` where `var.detectors` is a Terraform
`variable` of type `list(object({...}))`. The test SHALL use the minimal repro shape
from issue #2966 (`function` as required string; `field_name`, `by_field_name`,
`detector_description` as optional string). The test SHALL assert that plan and apply
complete without a `Value Conversion Error` and SHALL verify at least one detector
attribute in state (e.g. `analysis_config.detectors.0.function`). The test SHALL be
named consistently with the existing `TestAccResourceAnomalyDetectionJob*` convention.

#### Scenario: Acceptance test — variable-sourced detectors plan and apply

- GIVEN an acceptance test configuration that assigns `analysis_config.detectors` from
  a Terraform `variable` with a concrete default
- WHEN the acceptance test runs `terraform plan` and `terraform apply`
- THEN the resource SHALL be created without a `Value Conversion Error` and state SHALL
  reflect the expected detector function value
