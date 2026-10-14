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
# reordered keys, comments, and different indentation
fields:
    - control: INPUT_TEXT
      label:   Ticket   # trailing comment
      type: keyword
      name: ticket_ref

tags:
    - alpha
severity: low
name: Triage case
EOT
}
