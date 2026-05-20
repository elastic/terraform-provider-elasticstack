provider "elasticstack" {
  kibana {}
}

# Minimal skill: just instructions content.
resource "elasticstack_kibana_agentbuilder_skill" "summarize_incident" {
  skill_id    = "summarize-incident"
  name        = "Summarize incident"
  description = "Summarize an incident from the available signals."
  content     = <<-EOT
    When asked to summarize an incident:

    1. Gather the most relevant alerts and logs.
    2. Identify the affected services and time window.
    3. Produce a concise, factual summary.
  EOT
}

# Skill that references existing tools and includes ordered referenced content.
resource "elasticstack_kibana_agentbuilder_skill" "incident_playbook" {
  skill_id    = "incident-playbook"
  name        = "Incident playbook"
  description = "Run our standard incident response playbook."
  content     = <<-EOT
    Follow the standard incident response playbook step by step.
    Refer to the runbook in referenced content for the exact sequence.
  EOT

  tool_ids = ["platform.core.index_explorer"]

  referenced_content = [
    {
      name          = "Runbook"
      relative_path = "./runbooks/standard.md"
      content       = "## Standard runbook\n\n1. Acknowledge alert.\n2. Page on-call.\n3. Open incident channel."
    },
    {
      name          = "Glossary"
      relative_path = "./reference/glossary.md"
      content       = "## Glossary\n\nSLO: service-level objective.\nSLA: service-level agreement."
    },
  ]
}
