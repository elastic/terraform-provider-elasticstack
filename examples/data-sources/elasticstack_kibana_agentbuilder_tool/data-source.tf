provider "elasticstack" {
  kibana {}
}

# Read a tool by ID
data "elasticstack_kibana_agentbuilder_tool" "my_tool" {
  id = "my-esql-tool"
}

# Read a workflow-type tool and include the referenced workflow
data "elasticstack_kibana_agentbuilder_tool" "my_workflow_tool" {
  id               = "my-workflow-tool"
  include_workflow = true
}

output "workflow_yaml" {
  value = data.elasticstack_kibana_agentbuilder_tool.my_workflow_tool.workflow_configuration_yaml
}
