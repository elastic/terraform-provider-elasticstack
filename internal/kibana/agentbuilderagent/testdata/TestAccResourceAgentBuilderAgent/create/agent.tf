variable "agent_id" {
  description = "The agent ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Test Agent"
  description  = "A test agent for acceptance testing"
  labels       = ["test", "agent"]
  instructions = "You are a helpful assistant that searches logs. Use the available tools to help answer questions."
  tools        = ["platform.core.index_explorer"]
}
