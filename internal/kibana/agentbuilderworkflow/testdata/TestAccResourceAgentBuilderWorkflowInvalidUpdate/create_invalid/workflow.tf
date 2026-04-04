variable "workflow_id" {
  description = "The workflow ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test_invalid" {
  workflow_id        = var.workflow_id
  configuration_yaml = <<-EOT
not_working: hello_world
EOT
}
