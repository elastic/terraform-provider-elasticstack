variable "name" {
  type = string
}

variable "owner" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = var.owner
  name            = var.name
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
