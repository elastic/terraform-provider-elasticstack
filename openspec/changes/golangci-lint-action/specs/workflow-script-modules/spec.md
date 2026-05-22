## MODIFIED Requirements

### Requirement: Provider gate evaluates golangci-lint as a distinct job result

The `gateProvider` function in `lib/gate-provider.js` SHALL accept a `golangciLintResult` parameter alongside the existing `lintResult` parameter. Both results SHALL be included in the set of job results that the gate validates and evaluates, together with `buildResult` and `testResult`. The gate SHALL pass only when `classifyResult=true` and all evaluated job results (`buildResult`, `lintResult`, `golangciLintResult`, and `testResult`) are `success`. The gate SHALL fail if any evaluated job result is `failure` or `cancelled`. The gate SHALL treat all four jobs as legitimately skipped (returning `passed: true`) when `classifyResult` is `false` and all job results — including both lint results — are `skipped`.

The `fields` array for the `provider` gate in `lib/runners/gate.js` SHALL include `GOLANGCI_LINT_RESULT` alongside `LINT_RESULT`.

The `gate` job in `provider.yml` SHALL pass `PROVIDER_GATE_GOLANGCI_LINT_RESULT` as an environment variable sourced from the `golangci-lint` job result.

#### Scenario: Both lint jobs succeed

- **GIVEN** `golangciLintResult=success` and `lintResult=success`
- **AND** `classifyResult=true`, `buildResult=success`, `testResult=success`
- **WHEN** `gateProvider` is called
- **THEN** it SHALL return `passed: true`

#### Scenario: golangci-lint job fails

- **GIVEN** `golangciLintResult=failure` and `lintResult=success`
- **AND** all other results are `success`
- **WHEN** `gateProvider` is called
- **THEN** it SHALL return `passed: false`

#### Scenario: Other lint job fails

- **GIVEN** `golangciLintResult=success` and `lintResult=failure`
- **AND** all other results are `success`
- **WHEN** `gateProvider` is called
- **THEN** it SHALL return `passed: false`

#### Scenario: Non-provider change — all jobs skipped

- **GIVEN** `classifyResult=false`
- **AND** `buildResult=skipped`, `golangciLintResult=skipped`, `lintResult=skipped`, `testResult=skipped`
- **WHEN** `gateProvider` is called
- **THEN** it SHALL return `passed: true`

#### Scenario: Provider change but golangci-lint unexpectedly skipped

- **GIVEN** `classifyResult=true`
- **AND** `golangciLintResult=skipped`
- **AND** `lintResult=success`, `buildResult=success`, `testResult=success`
- **WHEN** `gateProvider` is called
- **THEN** it SHALL return `passed: false`
