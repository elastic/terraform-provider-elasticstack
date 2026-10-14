variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
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

resource "elasticstack_kibana_case_template" "duplicate" {
  owner           = "cases"
  name            = upper(var.name)
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

  depends_on = [elasticstack_kibana_case_template.test]
}
