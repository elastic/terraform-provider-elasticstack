variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  configuration_yaml = <<-EOT
name: New workflow
enabled: false
description: This is a new workflow
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

resource "elasticstack_kibana_agentbuilder_tool" "test_workflow" {
  tool_id     = var.tool_id
  type        = "workflow"
  description = "Test workflow tool"
  tags        = ["test", "workflow"]
  configuration = jsonencode({
    workflow_id = elasticstack_kibana_agentbuilder_workflow.test.workflow_id
  })
}
