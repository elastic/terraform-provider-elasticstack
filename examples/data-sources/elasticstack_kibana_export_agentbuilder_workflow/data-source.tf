provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_export_ab_workflow" "test" {
  id = "workflow-example"
}
