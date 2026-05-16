variable "agent_id" {
  description = "The agent ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id      = var.agent_id
  name          = "Avatar Test Agent"
  avatar_color  = "#BFDBFF"
  avatar_symbol = "TA"
}
