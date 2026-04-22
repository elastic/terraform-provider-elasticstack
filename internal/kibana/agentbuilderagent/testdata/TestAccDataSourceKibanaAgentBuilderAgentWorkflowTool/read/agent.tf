variable "agent_id" {
  description = "The agent ID"
  type        = string
}

variable "workflow_tool_id" {
  description = "The workflow tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  configuration_yaml = <<-EOT
name: Test Workflow
description: A test workflow for agent export
enabled: true
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "hello"
steps:
  - name: echo_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}

resource "elasticstack_kibana_agentbuilder_tool" "workflow" {
  tool_id     = var.workflow_tool_id
  type        = "workflow"
  description = "Workflow tool"
  configuration = jsonencode({
    workflow_id = elasticstack_kibana_agentbuilder_workflow.test.workflow_id
  })
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Agent With Workflow Tool"
  instructions = "Use the workflow tool."
  tools        = [elasticstack_kibana_agentbuilder_tool.workflow.tool_id]
}

data "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id             = elasticstack_kibana_agentbuilder_agent.test.agent_id
  include_dependencies = true
}
