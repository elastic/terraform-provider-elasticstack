# Export full agent rows (same pattern as terraform_remote_state workflows in import.tf).

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_tool" "exported_tool" {
  tool_id       = "doc-export-example-tool"
  type          = "esql"
  description   = "Tool for the export data source example"
  configuration = jsonencode({ query = "FROM logs-* | LIMIT 10" })
}

resource "elasticstack_kibana_agentbuilder_agent" "source" {
  agent_id     = "doc-export-source-agent"
  name         = "Documentation export agent"
  description  = "Agent whose configuration is exported by the data source below"
  instructions = "You are helpful."
  tools        = [elasticstack_kibana_agentbuilder_tool.exported_tool.tool_id]
}

data "elasticstack_kibana_agentbuilder_agent" "example" {
  agent_id             = elasticstack_kibana_agentbuilder_agent.source.agent_id
  include_dependencies = true
}

output "agent_export" {
  value = data.elasticstack_kibana_agentbuilder_agent.example
}
