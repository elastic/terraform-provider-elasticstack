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
enabled: true
description: A test workflow for acceptance testing
name: Test Workflow
triggers:
  - type: manual
inputs:
  - default: "hello world"
    type: string
    name: message
steps:
  - name: hello_world_step
    with:
      message: "{{ inputs.message }}"
    type: console
EOT
}
