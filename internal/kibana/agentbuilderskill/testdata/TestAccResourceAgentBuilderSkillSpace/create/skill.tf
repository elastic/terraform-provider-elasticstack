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
  name        = "Test Space for Skill"
  description = "Space for testing agent builder skills"
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  space_id    = elasticstack_kibana_space.test.space_id
  name        = "Space Skill"
  description = "A skill in a non-default space."
  content     = "Answer questions about this space."
}
