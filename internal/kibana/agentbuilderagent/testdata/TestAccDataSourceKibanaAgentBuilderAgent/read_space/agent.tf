variable "agent_id" {
  description = "The agent ID"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Agent Export"
  description = "Space for testing agent builder agent data source exports"
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id      = var.agent_id
  space_id      = elasticstack_kibana_space.test.space_id
  name          = "Space Agent"
  description   = "A space-scoped agent for export"
  avatar_color  = "#BFDBFF"
  avatar_symbol = "SA"
  labels        = ["space", "agent"]
  instructions  = "You are a helpful assistant for a Kibana space."
}

data "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id = elasticstack_kibana_agentbuilder_agent.test.agent_id
  space_id = elasticstack_kibana_space.test.space_id
}
