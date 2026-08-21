variable "agent_id" {
  description = "The agent ID"
  type        = string
}

variable "esql_tool_id" {
  description = "The ES|QL tool ID"
  type        = string
}

variable "workflow_tool_id" {
  description = "The workflow tool ID"
  type        = string
}

variable "index_search_tool_id" {
  description = "The index_search tool ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
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

resource "elasticstack_kibana_agentbuilder_tool" "esql" {
  tool_id     = var.esql_tool_id
  type        = "esql"
  description = "Test ES|QL tool"
  tags        = ["test", "esql"]
  configuration = jsonencode({
    query  = "FROM logs-* | LIMIT 10"
    params = {}
  })
}

resource "elasticstack_kibana_agentbuilder_tool" "workflow" {
  tool_id     = var.workflow_tool_id
  type        = "workflow"
  description = "Workflow tool"
  configuration = jsonencode({
    workflow_id = elasticstack_kibana_agentbuilder_workflow.test.workflow_id
  })
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name                = "agentbuilder-test-${var.index_search_tool_id}"
  deletion_protection = false
}

resource "elasticstack_kibana_agentbuilder_tool" "index_search" {
  depends_on = [elasticstack_elasticsearch_index.test_index]

  tool_id     = var.index_search_tool_id
  type        = "index_search"
  description = "Test index search tool"
  configuration = jsonencode({
    pattern = "agentbuilder-test-*"
  })
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Test Agent With Tools"
  description  = "Agent with esql, workflow, and index_search tools"
  instructions = "Use the available tools to help answer questions."
  tools = [
    elasticstack_kibana_agentbuilder_tool.esql.tool_id,
    elasticstack_kibana_agentbuilder_tool.workflow.tool_id,
    elasticstack_kibana_agentbuilder_tool.index_search.tool_id,
  ]
}

data "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id             = elasticstack_kibana_agentbuilder_agent.test.agent_id
  include_dependencies = true
}
