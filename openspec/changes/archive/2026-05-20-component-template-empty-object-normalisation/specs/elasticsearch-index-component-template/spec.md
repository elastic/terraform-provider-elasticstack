# Delta spec: component-template-empty-object-normalisation

Capability: `elasticsearch-index-component-template`  
Base spec: `openspec/specs/elasticsearch-index-component-template/spec.md`

## MODIFIED Requirements

### Requirement: Read state mapping (REQ-022–REQ-026)

The provider SHALL treat an empty `"mappings": {}` or `"settings": {}` object returned by Elasticsearch as semantically equivalent to an absent value (`null`).

- **Flatten-layer normalisation**: `flattenTemplateBlock` SHALL use `len(t.Mappings) > 0` (instead of `t.Mappings != nil`) and `len(t.Settings) > 0` (instead of `t.Settings != nil`) guards so that both nil and empty-map API responses produce `null` Terraform state values.

#### Scenario: Empty-object mappings response is null in state

- GIVEN a component template was created without a `template.mappings` block
- AND Elasticsearch returns `"mappings": {}` in the GET response
- WHEN read runs
- THEN `template.mappings` SHALL be `null` in state
- AND Terraform SHALL NOT report a provider inconsistent-result error or drift

#### Scenario: Empty-object settings response is null in state

- GIVEN a component template was created without a `template.settings` block
- AND Elasticsearch returns `"settings": {}` in the GET response
- WHEN read runs
- THEN `template.settings` SHALL be `null` in state
- AND Terraform SHALL NOT report a provider inconsistent-result error or drift

#### Scenario: No drift after create with alias and short-form settings (issue-609 regression)

- GIVEN a component template configured with:
  - an alias block (`name = "my_template_test"`)
  - `settings = jsonencode({ number_of_shards = "3" })` (short-form, keys not nested under `index`)
  - no `mappings` block
- WHEN `terraform apply` creates the resource
- AND a subsequent `terraform plan` runs
- THEN the plan SHALL be empty (no changes detected)
- AND Terraform SHALL NOT report a provider inconsistent-result error
