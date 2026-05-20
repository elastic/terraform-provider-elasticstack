variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Datasource Skill"
  description = "A skill for data source export."
  content     = "Sample content for export."

  referenced_content = [
    {
      name          = "Exported"
      relative_path = "./exported/path.md"
      content       = "Content available via the data source."
    },
  ]
}

data "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id = elasticstack_kibana_agentbuilder_skill.test.skill_id
}
