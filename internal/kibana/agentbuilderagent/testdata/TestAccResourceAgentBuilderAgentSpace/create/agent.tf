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
  name        = "Test Space for Agent"
  description = "Space for testing agent builder agents"
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  space_id     = elasticstack_kibana_space.test.space_id
  name         = "Space Agent"
  instructions = "You are a space-scoped agent."
}
