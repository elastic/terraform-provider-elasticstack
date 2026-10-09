## ADDED Requirements

### Requirement: Execution metadata is not exposed

The `elasticstack_kibana_alerting_rule` resource SHALL NOT define `last_execution_status` or `last_execution_date`. State written by earlier provider versions that still contains these attributes, including state in the original SDK format that is upgraded to the current schema, SHALL be read without error and the values SHALL be discarded. The resource schema version SHALL NOT change as part of this removal.

#### Scenario: Plan after an execution produces no diff
- GIVEN a managed rule whose configuration is unchanged
- AND Kibana has executed the rule since the last apply
- WHEN `terraform plan` is run
- THEN the plan SHALL show no changes

#### Scenario: Prior state containing the removed attributes
- GIVEN state at the current schema version written by an earlier provider version that includes `last_execution_status` and `last_execution_date`
- WHEN the resource is read or planned with the current provider
- THEN the operation SHALL succeed without error
- AND the removed attributes SHALL NOT appear in the new state

#### Scenario: Original SDK-format state containing the removed attributes
- GIVEN state in the original SDK format that includes `last_execution_status` and `last_execution_date`
- WHEN the state is upgraded to the current schema
- THEN the upgrade SHALL succeed without error
- AND the removed attributes SHALL NOT appear in the upgraded state

#### Scenario: Configuration referencing a removed attribute
- GIVEN a configuration that references `elasticstack_kibana_alerting_rule.x.last_execution_status`
- WHEN the configuration is validated
- THEN validation SHALL fail because the attribute does not exist

### Requirement: Removal is documented with its replacement

The generated documentation for `elasticstack_kibana_alerting_rule` SHALL NOT list `last_execution_status` or `last_execution_date`, and SHALL state that execution status and date are available from the `elasticstack_kibana_alerting_rules` data source. The pull request description that delivers the removal SHALL name both removed attributes, mark the change as breaking, and show a data source example that reads `last_execution_status` for a managed rule, so that generated release notes carry the migration guidance.

#### Scenario: Resource documentation names the replacement
- GIVEN the generated documentation page for the resource
- WHEN a practitioner reads it
- THEN it SHALL NOT list the two removed attributes
- AND it SHALL point to the `elasticstack_kibana_alerting_rules` data source for execution status and date

#### Scenario: Pull request carries migration wording
- GIVEN the pull request that removes the attributes
- WHEN its description is reviewed
- THEN it SHALL name `last_execution_status` and `last_execution_date`
- AND it SHALL mark the change as breaking
- AND it SHALL include a data source example that reads `last_execution_status`

## REMOVED Requirements

### Requirement: State mapping — execution metadata (REQ-025)
**Reason**: `last_execution_status` and `last_execution_date` change on every rule execution. As Computed-only attributes they produced a perpetual `-> (known after apply)` plan diff that `lifecycle { ignore_changes }` cannot suppress (elastic/terraform-provider-elasticstack#4981, #3342). The resource SHALL NOT expose these attributes.
**Migration**: Read the values with the `elasticstack_kibana_alerting_rules` data source, for example `data "elasticstack_kibana_alerting_rules" "r" { space_id = ...  rule_id = elasticstack_kibana_alerting_rule.x.rule_id }`, and reference `data.elasticstack_kibana_alerting_rules.r.rules[0].last_execution_status`. Remove any configuration that references the attributes on the resource. Existing state that still contains the attributes is read without error and the values are dropped.
