provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_ab_agent" "my_agent" {
  id            = "my-agent"
  name          = "My Agent"
  description   = "An example agent built with Agent Builder."
  avatar_color  = "#BFDBFF"
  avatar_symbol = "MA"
  labels        = ["example", "demo"]
  instructions  = "You are a helpful assistant."

  tools = [
    elasticstack_kibana_ab_tool.my_tool.id,
  ]
}
