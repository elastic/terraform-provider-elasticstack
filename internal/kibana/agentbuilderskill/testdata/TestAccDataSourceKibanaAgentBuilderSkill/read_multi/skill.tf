variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Datasource Skill Multi"
  description = "A skill with multiple tools and references."
  content     = "Multi content."

  tool_ids = ["platform.core.index_explorer", "platform.core.search"]

  referenced_content = [
    {
      name          = "First"
      relative_path = "./first/path.md"
      content       = "First referenced content."
    },
    {
      name          = "Second"
      relative_path = "./second/path.md"
      content       = "Second referenced content."
    },
    {
      name          = "Third"
      relative_path = "./third/path.md"
      content       = "Third referenced content."
    },
  ]
}

data "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id = elasticstack_kibana_agentbuilder_skill.test.skill_id
}
