provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id = "workflow-example"
}
