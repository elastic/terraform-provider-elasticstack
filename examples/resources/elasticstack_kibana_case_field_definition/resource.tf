provider "elasticstack" {
  kibana {}
}

# Minimal field: only owner and definition are required. The field name is
# taken from the `name` key in the YAML.
resource "elasticstack_kibana_case_field_definition" "minimal" {
  owner      = "cases"
  definition = <<-EOT
    name: ticket_reference
    type: keyword
    label: Ticket reference
    control: INPUT_TEXT
  EOT
}

# Field with all optional attributes. `name` must match the `name` key in the
# YAML. `name` and the YAML `type` are immutable: changing `name` replaces the
# resource, while changing only the YAML `type` is rejected by the API with a
# 409 `field_identity_immutable` error.
resource "elasticstack_kibana_case_field_definition" "global" {
  owner       = "securitySolution"
  name        = "customer_tier"
  description = "Support tier of the customer affected by the case."
  is_global   = true
  definition  = <<-EOT
    name: customer_tier
    type: keyword
    label: Customer tier
    control: INPUT_TEXT
  EOT
}

# Field in a non-default Kibana space.
resource "elasticstack_kibana_case_field_definition" "in_space" {
  space_id   = "ops"
  owner      = "observability"
  definition = <<-EOT
    name: runbook_url
    type: keyword
    label: Runbook URL
    control: INPUT_TEXT
  EOT
}
