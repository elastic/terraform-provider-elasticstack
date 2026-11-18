variable "workflow_id" {
  description = "A workflow ID that does not exist"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id = var.workflow_id
}
