variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = "${var.name}_renamed"
  description     = ""
  tags            = []
  is_enabled      = true
  definition_yaml = <<-EOT
name: Triage case
severity: high
tags:
  - alpha
fields:
  - name: ticket_ref
    type: keyword
    label: Ticket
    control: INPUT_TEXT
EOT
}
