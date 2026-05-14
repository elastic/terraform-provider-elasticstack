variable "workflow_id" {
  description = "The workflow ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  workflow_id = var.workflow_id
  configuration_yaml = <<-EOT
name: No Description Workflow
enabled: true
triggers:
  - type: manual
steps:
  - name: test_step
    type: console
    with:
      message: "workflow with no description field"
EOT
}
