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
  name        = "Test Space for Composite Skill"
  description = "Space for testing composite skill_id resolution"
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  space_id    = elasticstack_kibana_space.test.space_id
  name        = "Composite Skill"
  description = "A space-scoped skill read via composite id."
  content     = "Composite content."
}

data "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id = "${elasticstack_kibana_space.test.space_id}/${elasticstack_kibana_agentbuilder_skill.test.skill_id}"
  space_id = elasticstack_kibana_space.test.space_id
}
