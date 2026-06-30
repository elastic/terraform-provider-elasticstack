# `elasticstack_kibana_slo` — SLO acceptance test version constraints delta

**Change:** `slo-tests-skip-8-10-4`
**Base spec:** `openspec/specs/kibana-slo/spec.md`

## Purpose

Kibana 8.10.4 returns HTTP 500 (`illegal_argument_exception: must specify at least one document
in [docs]`) when creating an SLO against an empty source index. This appears to be a regression in
8.10.4; whether other 8.10.x patch versions are affected is unconfirmed. The three SLO
acceptance tests that use freshly-created, empty indices must be skipped on 8.10.4, following
the same version-exclusion pattern established for 8.11.x.

## MODIFIED Requirements

### Requirement: SLOKqlAccTestConstraints excludes Kibana 8.10.4

`SLOKqlAccTestConstraints` in `internal/kibana/slo/constants.go` SHALL include `!=8.10.4`
alongside the existing `!=8.11.{0–4}` exclusions so that acceptance tests consuming this
constant are automatically skipped on 8.10.4.

The constraint string SHALL be `>=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4`.

#### Scenario: 8.10.4 cluster — constraint not satisfied
- **WHEN** `SLOKqlAccTestConstraints` is evaluated against a connected cluster reporting version `8.10.4`
- **THEN** the constraint is NOT satisfied and any test step gated on it is skipped

#### Scenario: 8.10.3 cluster — constraint satisfied
- **WHEN** `SLOKqlAccTestConstraints` is evaluated against a connected cluster reporting version `8.10.3`
- **THEN** the constraint is satisfied and the test step proceeds normally

#### Scenario: 8.12.0 cluster — constraint satisfied
- **WHEN** `SLOKqlAccTestConstraints` is evaluated against a connected cluster reporting version `8.12.0`
- **THEN** the constraint is satisfied and the test step proceeds normally

### Requirement: SLOKqlFleetAccTestConstraints excludes Kibana 8.10.4

`SLOKqlFleetAccTestConstraints` in `internal/kibana/slo/constants.go` SHALL include `!=8.10.4`
alongside the existing `!=8.11.{0–4}` exclusions.

The constraint string SHALL be `>=8.10.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4`.

#### Scenario: 8.10.4 cluster — fleet-style step skipped
- **WHEN** `SLOKqlFleetAccTestConstraints` is evaluated against a connected cluster reporting version `8.10.4`
- **THEN** the constraint is NOT satisfied and any fleet-style test step gated on it is skipped

#### Scenario: 8.10.0 cluster — fleet-style step proceeds
- **WHEN** `SLOKqlFleetAccTestConstraints` is evaluated against a connected cluster reporting version `8.10.0`
- **THEN** the constraint is satisfied and the fleet-style test step proceeds normally

### Requirement: TestAccResourceSlo_36_char_slo_id skips on Kibana 8.10.4

The inline `version.NewConstraint` string in `TestAccResourceSlo_36_char_slo_id` SHALL include
`!=8.10.4` so that the test is skipped rather than failing on 8.10.4.

The constraint string SHALL be `>=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4,<8.16.0`.

#### Scenario: 8.10.4 cluster — test skipped
- **WHEN** `TestAccResourceSlo_36_char_slo_id` is executed against Kibana 8.10.4
- **THEN** `SkipIfUnsupportedConstraints` causes the test to be reported as SKIPPED, not FAILED

#### Scenario: 8.13.0 cluster — test runs
- **WHEN** `TestAccResourceSlo_36_char_slo_id` is executed against Kibana 8.13.0
- **THEN** `SkipIfUnsupportedConstraints` is satisfied and the test runs normally

### Requirement: TestAccResourceSloFromSDK skips on Kibana 8.10.4

The inline `version.NewConstraint` string in `TestAccResourceSloFromSDK` SHALL include
`!=8.10.4` so that the test (including its external-provider step 1) is skipped rather than
failing on 8.10.4.

The constraint string SHALL be `>=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4`.

#### Scenario: 8.10.4 cluster — test skipped before step 1
- **WHEN** `TestAccResourceSloFromSDK` is executed against Kibana 8.10.4
- **THEN** `SkipIfUnsupportedConstraints` causes the test to be reported as SKIPPED before the external-provider step 1 runs

#### Scenario: 8.14.0 cluster — both steps run
- **WHEN** `TestAccResourceSloFromSDK` is executed against Kibana 8.14.0
- **THEN** `SkipIfUnsupportedConstraints` is satisfied and both test steps run normally
