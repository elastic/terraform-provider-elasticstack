provider "elasticstack" {
  kibana {}
}

# Basic agent with tools
resource "elasticstack_kibana_agentbuilder_agent" "my_agent" {
  agent_id      = "my-agent"
  name          = "My Agent"
  description   = "An example agent built with Agent Builder."
  avatar_color  = "#BFDBFF"
  avatar_symbol = "MA"
  labels        = ["example", "demo"]
  instructions  = "You are a helpful assistant."

  tools = [
    elasticstack_kibana_agentbuilder_tool.my_tool.tool_id,
  ]
}

# Agent in a non-default space
resource "elasticstack_kibana_space" "my_space" {
  space_id = "my-space"
  name     = "My Space"
}

resource "elasticstack_kibana_agentbuilder_agent" "space_agent" {
  agent_id = "space-agent"
  space_id = elasticstack_kibana_space.my_space.space_id
  name     = "Space-Scoped Agent"
}
