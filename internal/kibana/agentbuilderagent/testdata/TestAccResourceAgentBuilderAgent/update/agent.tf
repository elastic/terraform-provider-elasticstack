variable "agent_id" {
  description = "The agent ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Updated Test Agent"
  description  = "An updated test agent"
  labels       = ["test", "agent", "updated"]
  instructions = "You are an updated helpful assistant. Use the available tools wisely."
  tools        = ["platform.core.index_explorer", "platform.core.list_indices"]
}
