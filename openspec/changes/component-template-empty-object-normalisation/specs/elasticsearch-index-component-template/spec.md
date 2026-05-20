# Delta spec: component-template-empty-object-normalisation

Capability: `elasticsearch-index-component-template`  
Base spec: `openspec/specs/elasticsearch-index-component-template/spec.md`

## MODIFIED Requirements

### Requirement: Empty-object normalisation for mappings and settings on read (REQ-022)

The provider SHALL treat a non-nil empty `"mappings": {}` or `"settings": {}` object returned by Elasticsearch as equivalent to an absent field, and SHALL set `template.mappings` or `template.settings` to `null` in Terraform state accordingly. `flattenTemplateBlock` SHALL use `len(t.Mappings) > 0` (instead of `t.Mappings != nil`) and `len(t.Settings) > 0` (instead of `t.Settings != nil`) as the guards for producing non-null values; both nil and empty maps MUST produce null Terraform values.

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
