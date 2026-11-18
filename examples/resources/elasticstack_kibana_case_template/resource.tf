provider "elasticstack" {
  kibana {}
}

# Minimal template: only owner, name and definition_yaml are required. Case
# defaults (severity, category, tags, assignees, connector, settings) and the
# custom fields are declared in the YAML. The Cases templates API is Technical
# Preview, requires Elastic Stack 9.5.2 or later, and is only available when
# the Kibana setting `xpack.cases.templates.enabled` is true (the default).
resource "elasticstack_kibana_case_template" "minimal" {
  owner           = "securitySolution"
  name            = "triage-template"
  definition_yaml = <<-EOT
    name: Triage case
    severity: low
    fields:
      - name: ticket_reference
        type: keyword
        label: Ticket reference
        control: INPUT_TEXT
  EOT
}

# Template with all optional attributes. `tags` preserves order and duplicates.
# Removing `description` or `tags` from the configuration later clears the
# stored value (unlike the field definition resource, which preserves it).
resource "elasticstack_kibana_case_template" "full" {
  owner       = "securitySolution"
  name        = "incident-template"
  description = "Defaults for analyst-created incident cases."
  tags        = ["incident", "tier-1"]
  is_enabled  = false

  definition_yaml = <<-EOT
    name: Incident case
    severity: high
    category: incident
    settings:
      syncAlerts: true
    fields:
      - name: customer_tier
        type: keyword
        label: Customer tier
        control: INPUT_TEXT
  EOT
}

# Template that references a field from the field library by `$ref`
# (requires Elastic Stack 9.6.0 or later for the field library API).
resource "elasticstack_kibana_case_field_definition" "ticket" {
  owner      = "cases"
  definition = <<-EOT
    name: ticket_reference
    type: keyword
    label: Ticket reference
    control: INPUT_TEXT
  EOT
}

resource "elasticstack_kibana_case_template" "with_ref" {
  owner           = "cases"
  name            = "ref-template"
  definition_yaml = <<-EOT
    name: Case with a library field
    fields:
      - $ref: ${elasticstack_kibana_case_field_definition.ticket.name}
  EOT
}

# Template in a non-default Kibana space.
resource "elasticstack_kibana_case_template" "in_space" {
  space_id        = "ops"
  owner           = "observability"
  name            = "ops-template"
  definition_yaml = <<-EOT
    name: Ops case
    fields: []
  EOT
}
