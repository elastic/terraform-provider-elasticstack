provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "not-a-solution"
  name            = "t"
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
