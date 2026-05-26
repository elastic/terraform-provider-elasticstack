variable "skill_id" {
  description = "The skill ID"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Skill Export"
  description = "Space for testing agent builder skill data source exports"
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  space_id    = elasticstack_kibana_space.test.space_id
  name        = "Space Skill"
  description = "A space-scoped skill for export."
  content     = "Answer questions about this space."
}

data "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id = elasticstack_kibana_agentbuilder_skill.test.skill_id
  space_id = elasticstack_kibana_space.test.space_id
}
