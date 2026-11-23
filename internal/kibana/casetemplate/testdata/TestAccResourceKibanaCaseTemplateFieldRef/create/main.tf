variable "name" {
  type = string
}

variable "field_name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  definition = <<-EOT
name: ${var.field_name}
type: keyword
label: Referenced field
control: INPUT_TEXT
EOT
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = var.name
  definition_yaml = <<-EOT
name: Triage with ref
fields:
  - $ref: ${elasticstack_kibana_case_field_definition.test.name}
EOT
}
