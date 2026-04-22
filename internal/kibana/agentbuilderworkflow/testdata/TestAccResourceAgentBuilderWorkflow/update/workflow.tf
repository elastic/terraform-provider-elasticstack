variable "workflow_id" {
  description = "The workflow ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  workflow_id        = var.workflow_id
  configuration_yaml = <<-EOT
name: Updated Test Workflow
description: An updated test workflow
enabled: false
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "hello world, updated"
steps:
  - name: updated_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}
