variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Updated Test Skill"
  description = "Updated description"
  content     = "Be precise and cite sources."
}
