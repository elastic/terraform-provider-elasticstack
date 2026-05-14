## ADDED Requirements

### Requirement: Provider registration (REQ-040)

The `elasticstack_kibana_dashboard` resource SHALL be registered through the provider's standard Plugin Framework resource set returned by `Provider.resources(...)` in `provider/plugin_framework.go`. It SHALL NOT be returned from `Provider.experimentalResources(...)`, and practitioners SHALL NOT be required to set `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` to use the resource.

#### Scenario: Default provider surface includes the resource

- **GIVEN** a released provider build (no `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` override)
- **WHEN** Terraform requests the provider's resource set
- **THEN** `elasticstack_kibana_dashboard` SHALL be present in the resources returned by `Provider.Resources(ctx)`

#### Scenario: Experimental resource set excludes the dashboard resource

- **GIVEN** the provider's experimental Plugin Framework resource set returned by `Provider.experimentalResources(ctx)`
- **WHEN** that set is enumerated
- **THEN** it SHALL NOT contain `dashboard.NewResource`

#### Scenario: Practitioner does not need the experimental opt-in

- **GIVEN** a Terraform configuration declaring `resource "elasticstack_kibana_dashboard" "example" { ... }`
- **WHEN** Terraform plans or applies against a released provider build with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` unset
- **THEN** the provider SHALL recognize and operate the resource without requiring the environment variable
