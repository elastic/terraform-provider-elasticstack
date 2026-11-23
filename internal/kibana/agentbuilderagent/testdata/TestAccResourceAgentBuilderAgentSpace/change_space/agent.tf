variable "agent_id" {
  description = "The agent ID"
  type        = string
}

variable "space_id" {
  description = "The original Kibana space ID"
  type        = string
}

variable "new_space_id" {
  description = "The Kibana space ID the agent is moved to"
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

resource "elasticstack_kibana_space" "changed" {
  space_id    = var.new_space_id
  name        = "Changed Space for Agent"
  description = "Second space used to verify space_id RequiresReplace"
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  space_id     = elasticstack_kibana_space.changed.space_id
  name         = "Space Agent"
  instructions = "You are a space-scoped agent."
}
