variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Full Skill"
  description = "A skill that exercises tool_ids and referenced_content at create."
  content     = "Use the tools and references below to answer questions."

  tool_ids = ["platform.core.index_explorer"]

  referenced_content = [
    {
      name          = "Initial"
      relative_path = "./initial/path.md"
      content       = "Initial referenced content."
    },
  ]
}
