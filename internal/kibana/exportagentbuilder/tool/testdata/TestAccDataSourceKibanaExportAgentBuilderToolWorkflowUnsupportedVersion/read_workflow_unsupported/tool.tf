variable "tool_id" {
  description = "The tool ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "test" {
  tool_id     = var.tool_id
  type        = "workflow"
  description = "Workflow tool"
  configuration = jsonencode({
    workflow_id = "test-workflow-id"
  })
}

data "elasticstack_kibana_agentbuilder_export_tool" "test" {
  id               = elasticstack_kibana_agentbuilder_tool.test.id
  include_workflow = true
}
