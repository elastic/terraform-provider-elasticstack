provider "elasticstack" {
  kibana {}
}

# ES|QL tool
resource "elasticstack_kibana_agentbuilder_tool" "esql_tool" {
  tool_id     = "my-esql-tool"
  type        = "esql"
  description = "Analyzes trade data with time filtering"
  tags        = ["analytics", "finance"]
  configuration = jsonencode({
    query = "FROM financial_trades | WHERE execution_timestamp >= ?startTime | STATS trade_count=COUNT(*), avg_price=AVG(execution_price) BY symbol | SORT trade_count DESC | LIMIT ?limit"
    params = {
      limit = {
        type        = "integer"
        description = "Maximum number of results to return"
      }
      startTime = {
        type        = "date"
        description = "Start time for the analysis in ISO format"
      }
    }
  })
}

# Workflow tool — references an agent builder workflow
resource "elasticstack_kibana_agentbuilder_workflow" "my_workflow" {
  configuration_yaml = <<-EOT
name: My Workflow
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

resource "elasticstack_kibana_agentbuilder_tool" "workflow_tool" {
  tool_id     = "my-workflow-tool"
  type        = "workflow"
  description = "Exposes a workflow as an agent tool"
  configuration = jsonencode({
    workflow_id = elasticstack_kibana_agentbuilder_workflow.my_workflow.workflow_id
  })
}
