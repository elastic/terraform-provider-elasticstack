# Export an agent and all its dependencies (tools + workflows) from a cluster.
#
# The single "agent" output captures everything needed to recreate the
# agent in another cluster via the import configuration.

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "agent_id" {
  description = "The ID of the agent to export."
  type        = string
}

data "elasticstack_kibana_export_ab_agent" "this" {
  id                   = var.agent_id
  include_dependencies = true
}

output "agent" {
  value = jsonencode({
    agent     = data.elasticstack_kibana_export_ab_agent.this.agent
    tools     = data.elasticstack_kibana_export_ab_agent.this.tools
    workflows = data.elasticstack_kibana_export_ab_agent.this.workflows
  })
}
