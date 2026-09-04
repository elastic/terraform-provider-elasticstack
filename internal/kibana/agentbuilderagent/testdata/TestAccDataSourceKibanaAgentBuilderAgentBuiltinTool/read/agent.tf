variable "agent_id" {
  description = "The agent ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Agent With Builtin Tool"
  instructions = "Use the available tools to help answer questions."
  tools        = ["platform.core.index_explorer"]
}

data "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id             = elasticstack_kibana_agentbuilder_agent.test.agent_id
  include_dependencies = true
}
