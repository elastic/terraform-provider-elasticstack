provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "lookup" {
  tool_id       = "doc-datasource-example-tool"
  type          = "esql"
  description   = "Example tool for datasource documentation"
  configuration = jsonencode({ query = "FROM logs-* | LIMIT 1" })
}

data "elasticstack_kibana_agentbuilder_tool" "my_tool" {
  id = elasticstack_kibana_agentbuilder_tool.lookup.tool_id
}

data "elasticstack_kibana_agentbuilder_tool" "my_workflow_tool" {
  id               = elasticstack_kibana_agentbuilder_tool.workflow_lookup.tool_id
  include_workflow = true
}

resource "elasticstack_kibana_agentbuilder_workflow" "for_tool_ds" {
  configuration_yaml = <<-EOT
name: Workflow For Tool DS
enabled: true
triggers:
  - type: manual
inputs: []
steps:
  - name: noop
    type: console
    with:
      message: "workflow"
EOT
}

resource "elasticstack_kibana_agentbuilder_tool" "workflow_lookup" {
  tool_id     = "doc-datasource-example-workflow-tool"
  type        = "workflow"
  description = "Workflow tool for datasource example"
  configuration = jsonencode({
    workflow_id = elasticstack_kibana_agentbuilder_workflow.for_tool_ds.workflow_id
  })

  depends_on = [elasticstack_kibana_agentbuilder_workflow.for_tool_ds]
}

output "workflow_yaml" {
  value = data.elasticstack_kibana_agentbuilder_tool.my_workflow_tool.workflow_configuration_yaml
}
