## ADDED Requirements

### Requirement: Dashboard model package extraction preserves behavior (REQ-043)

The `elasticstack_kibana_dashboard` resource implementation SHALL isolate Terraform model structs in `internal/kibana/dashboard/models` while preserving the resource's externally observable behavior. All Terraform model structs used by the dashboard resource and its panel/config submodels SHALL move to the `models` package and SHALL be exported so conversion, validation, schema, and lifecycle code in `internal/kibana/dashboard` can reference them without import cycles.

This extraction SHALL be mechanical only: it SHALL NOT change the Terraform schema, API payload shapes, state alignment rules, plan semantics, import behavior, or read/write behavior described by existing dashboard requirements.

#### Scenario: Existing dashboard configuration remains behaviorally unchanged after model extraction

- GIVEN a dashboard configuration that was accepted before the model extraction
- WHEN the provider plans, applies, reads, and refreshes that configuration after the extraction
- THEN the resource SHALL expose the same schema and SHALL produce the same observable behavior and state transitions as before

#### Scenario: Dashboard logic references exported models package types

- GIVEN dashboard resource implementation code that performs schema, validation, lifecycle, or API conversion work
- WHEN it references dashboard Terraform model structs
- THEN those structs SHALL be imported from `internal/kibana/dashboard/models`
- AND the extraction SHALL avoid introducing Go import cycles between the models package and dashboard logic packages
