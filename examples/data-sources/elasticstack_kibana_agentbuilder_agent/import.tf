# Import an agent and its writable tools into another cluster.
#
# Prerequisites:
#   Run `terraform apply` in the export configuration first. This root reads the
#   `agent_export` output from that state.
#
# Resource ordering:
#   1. Workflows are created first (from workflow-type tools’ exported YAML).
#   2. Writable tools are created next. Workflow-type tools get a new workflow_id
#      pointing at the recreated workflow resources.
#   3. The agent is created last with depends_on on the tools, and its `tools`
#      list combines recreated tool IDs with any readonly (built-in) tool IDs
#      from the export (those are not managed as resources).

provider "elasticstack" {
  kibana {}
}

variable "export_state_path" {
  description = "Path to the terraform.tfstate file produced by the export configuration."
  type        = string
  default     = "../export/terraform.tfstate"
}

data "terraform_remote_state" "export" {
  backend = "local"

  config = {
    path = var.export_state_path
  }
}

locals {
  exported = data.terraform_remote_state.export.outputs.agent_export

  all_tools = coalesce(lookup(local.exported, "tools", []), [])

  # Readonly tools already exist on the target cluster; only create custom/writable tools.
  tools = [for t in local.all_tools : t if !try(t.readonly, false)]

  readonly_tool_ids = [for t in local.all_tools : t.tool_id if try(t.readonly, false)]

  workflows = [
    for t in local.tools : { id = t.workflow_id, yaml = t.workflow_configuration_yaml }
    if t.type == "workflow" && try(t.workflow_id, "") != "" && try(t.workflow_configuration_yaml, "") != ""
  ]

  old_workflow_id_to_index = {
    for i, w in local.workflows : w.id => i
  }

  tool_configurations = [
    for t in local.tools : (
      t.type == "workflow"
      ? jsonencode({
        workflow_id = elasticstack_kibana_agentbuilder_workflow.workflows[
          local.old_workflow_id_to_index[t.workflow_id]
        ].workflow_id
      })
      : t.configuration
    )
  ]
}

# 1. Workflows (from exported workflow-type tools)

resource "elasticstack_kibana_agentbuilder_workflow" "workflows" {
  count              = length(local.workflows)
  configuration_yaml = local.workflows[count.index].yaml
}

# 2. Writable tools

resource "elasticstack_kibana_agentbuilder_tool" "tools" {
  count       = length(local.tools)
  tool_id     = local.tools[count.index].tool_id
  type        = local.tools[count.index].type
  description = try(local.tools[count.index].description, null)
  tags        = try(local.tools[count.index].tags, null)

  configuration = local.tool_configurations[count.index]
  depends_on    = [elasticstack_kibana_agentbuilder_workflow.workflows]
}

# 3. Agent (all tool IDs = recreated writable tools + built-in readonly IDs from export)

resource "elasticstack_kibana_agentbuilder_agent" "agent" {
  agent_id      = local.exported.agent_id
  name          = local.exported.name
  description   = try(local.exported.description, null)
  avatar_color  = try(local.exported.avatar_color, null)
  avatar_symbol = try(local.exported.avatar_symbol, null)
  labels        = try(local.exported.labels, null) == null ? null : tolist(local.exported.labels)
  instructions  = try(local.exported.instructions, null)

  tools = concat(
    [for r in elasticstack_kibana_agentbuilder_tool.tools : r.tool_id],
    local.readonly_tool_ids
  )

  depends_on = [elasticstack_kibana_agentbuilder_tool.tools]
}
