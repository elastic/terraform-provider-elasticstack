variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  configuration_yaml = <<-EOT
name: Test Workflow
description: A test workflow for tool export
enabled: true
triggers:
  - type: manual
inputs:
  - name: data
    type: string
    default: "test"
steps:
  - name: process_step
    type: console
    with:
      message: "{{ inputs.data }}"
EOT
}

resource "elasticstack_kibana_agentbuilder_tool" "test" {
  tool_id     = var.tool_id
  type        = "workflow"
  description = "Workflow tool"
  configuration = jsonencode({
    workflow_id = elasticstack_kibana_agentbuilder_workflow.test.workflow_id
  })
}

data "elasticstack_kibana_agentbuilder_export_tool" "test" {
  id               = elasticstack_kibana_agentbuilder_tool.test.id
  include_workflow = true
}
