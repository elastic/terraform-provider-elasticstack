variable "name" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Case template test space"
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = var.name
  space_id        = elasticstack_kibana_space.test.space_id
  definition_yaml = <<-EOT
name: Triage case
severity: low
tags:
  - alpha
fields:
  - name: ticket_ref
    type: keyword
    label: Ticket
    control: INPUT_TEXT
EOT
}
