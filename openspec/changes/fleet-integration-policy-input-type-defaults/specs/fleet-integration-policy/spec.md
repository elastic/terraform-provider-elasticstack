## ADDED Requirements

### Requirement: Input-type package defaults extraction (REQ-NEW-INPUT-TYPE)

The defaults extractor SHALL handle both **integration-type** and **input-type** Fleet packages when extracting variable defaults from `PackageInfo`.

Integration-type packages declare `policyTemplates[].inputs[]` with a `type` and `vars` array per entry. Input-type packages declare a single top-level `input` string and a `vars` array at the policy-template level. For an input-type template, the extractor SHALL read `policyTemplate.input` as the input type, call `apiVars.defaults()` on `policyTemplate.vars`, and store the result keyed as `"{policyTemplate.name}-{policyTemplate.input}"` — the same key format used for integration-type packages. The resulting defaults SHALL then be combined with stream-level defaults from `apiDatastreams.defaults()` by the existing `packageInfoToDefaults()` function without further modification.

#### Scenario: Input-type package apply — no inconsistency error

- GIVEN an `elasticstack_fleet_integration_policy` resource targeting an input-type package (e.g. `gcp_pubsub`) where the user configures only a subset of the available variables
- WHEN Terraform applies the configuration
- THEN the provider SHALL NOT produce a `"Provider produced inconsistent result after apply"` error on `.inputs`
- AND the resulting Terraform state SHALL reflect the Kibana API response (including package-default vars) without a follow-up plan diff

#### Scenario: Input-type defaults extraction — known defaults present

- GIVEN a `PackageInfo` for an input-type package with at least one variable that carries a non-null `default` value (e.g. `subscription_type: "shared"`)
- WHEN `packageInfoToDefaults(pkg)` is called
- THEN the defaults map SHALL contain an entry for the expected input ID
- AND the entry's `Vars` JSON object SHALL include the defaulted variable
- AND the entry's `Vars` JSON object SHALL NOT include variables whose `default` is null and `multi` is false

#### Scenario: Input-type defaults extraction — non-defaulted vars omitted

- GIVEN a `PackageInfo` for an input-type package with at least one variable that has no `default` value and `multi: false` (e.g. `project_id`)
- WHEN `packageInfoToDefaults(pkg)` is called
- THEN the extracted defaults JSON SHALL NOT include that variable

### Requirement: Acceptance test coverage for input-type packages (REQ-NEW-INPUT-TYPE-ACC)

The acceptance test suite for `elasticstack_fleet_integration_policy` SHALL include at least one test case targeting an input-type package (e.g. `gcp_pubsub`) that configures a policy with only a subset of available vars, applies it, and then runs a plan that MUST produce no diff.

The test SHALL be skipped when the target Elastic Stack version is strictly below `8.10.0`.

#### Scenario: Acceptance test — apply and re-plan produce no diff

- GIVEN an input-type `gcp_pubsub` integration policy with only the required user-visible vars set (e.g. `project_id`, `subscription_name`, `topic`)
- WHEN the Terraform apply and a subsequent plan run
- THEN no `"inconsistent values for sensitive attribute"` error occurs during apply
- AND the subsequent plan SHALL show no changes
