# Import an agent and all its dependencies into another cluster.
#
# Prerequisites:
#   Run `terraform apply` in the export directory first. This config
#   reads the single "agent" output from that state to recreate the
#   full agent with its tools and workflows.
#
# Resource ordering:
#   1. Workflows are created first (no dependencies).
#   2. Tools are created next. Workflow-type tools reference the new
#      workflow ID via interpolation, so Terraform resolves the
#      dependency automatically.
#   3. The agent is created last with an explicit depends_on for tools,
#      because the tool IDs come from the exported state (not a resource
#      reference Terraform can track).

provider "elasticstack" {
  elasticsearch {}
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
  exported = jsondecode(data.terraform_remote_state.export.outputs.agent)

  agent = jsondecode(local.exported.agent)

  # Readonly tools are built-in and already exist on the target cluster,
  # so we only create the writable ones.
  tools     = [for t in local.exported.tools : t if !t.readonly]
  workflows = local.exported.workflows

  # Map each old workflow ID to its index in local.workflows so we can
  # find the matching new workflow resource by position.
  old_workflow_id_to_index = {
    for i, w in local.workflows : w.id => i
  }

  # For each tool, pre-compute the configuration to use on the new cluster.
  # Workflow-type tools need their workflow_id swapped to the newly created one;
  # all other tool types keep their original configuration as-is.
  tool_configurations = [
    for t in local.tools : (
      t.type == "workflow"
      ? jsonencode({
        workflow_id = elasticstack_kibana_ab_workflow.workflows[
          local.old_workflow_id_to_index[jsondecode(t.configuration).workflow_id]
        ].id
      })
      : t.configuration
    )
  ]
}

# 1. Create workflows

resource "elasticstack_kibana_ab_workflow" "workflows" {
  count         = length(local.workflows)
  configuration = local.workflows[count.index].yaml
}

# 2. Create tools

resource "elasticstack_kibana_ab_tool" "tools" {
  count       = length(local.tools)
  id          = local.tools[count.index].id
  type        = local.tools[count.index].type
  description = local.tools[count.index].description
  tags        = local.tools[count.index].tags

  configuration = local.tool_configurations[count.index]
}

# 3. Create agent

resource "elasticstack_kibana_ab_agent" "agent" {
  id            = local.agent.id
  name          = local.agent.name
  description   = try(local.agent.description, null)
  avatar_color  = try(local.agent.avatar_color, null)
  avatar_symbol = try(local.agent.avatar_symbol, null)
  labels        = try(local.agent.labels, null)
  instructions  = try(local.agent.configuration.instructions, null)

  tools = try(
    flatten([for t in local.agent.configuration.tools : t.tool_ids]),
    null
  )

  depends_on = [elasticstack_kibana_ab_tool.tools]
}
