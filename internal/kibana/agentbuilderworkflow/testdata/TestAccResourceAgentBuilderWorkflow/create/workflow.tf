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
name: Test Workflow
description: A test workflow for acceptance testing
enabled: true
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "hello world"
steps:
  - name: hello_world_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}
