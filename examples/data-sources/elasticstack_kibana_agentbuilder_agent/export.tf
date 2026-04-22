# Export an agent and its dependencies (full tool rows + workflow YAML on workflow tools).
#
# Set include_dependencies = true so each `tools` entry includes type, configuration,
# readonly, tags, and (for workflow tools) workflow_id + workflow_configuration_yaml.
# Omit it or set false to only get tool id / space_id / tool_id references (no API tool fetch).
#
# The `agent_export` output is the whole data source object (read via
# terraform_remote_state.outputs.agent_export — no jsondecode needed).

provider "elasticstack" {
  kibana {}
}

variable "agent_id" {
  description = "The agent ID to export (plain id, or `<space_id>/<agent_id>` if space_id is not set on the data source)."
  type        = string
}

variable "space_id" {
  description = "Kibana space for the agent. Leave null for the default space."
  type        = string
  default     = null
}

data "elasticstack_kibana_agentbuilder_agent" "example" {
  agent_id             = var.agent_id
  space_id             = var.space_id
  include_dependencies = true
}

output "agent_export" {
  value = data.elasticstack_kibana_agentbuilder_agent.example
}
