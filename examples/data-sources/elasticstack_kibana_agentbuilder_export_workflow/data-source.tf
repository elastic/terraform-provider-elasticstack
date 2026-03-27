provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_agentbuilder_export_workflow" "test" {
  id = "workflow-example"
}
