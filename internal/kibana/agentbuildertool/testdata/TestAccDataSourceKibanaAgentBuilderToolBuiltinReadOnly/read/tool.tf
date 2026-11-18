provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_agentbuilder_tool" "test" {
  id = "platform.core.index_explorer"
}
