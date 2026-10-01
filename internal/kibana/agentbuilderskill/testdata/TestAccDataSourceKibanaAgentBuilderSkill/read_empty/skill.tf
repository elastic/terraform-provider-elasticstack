variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Datasource Skill Empty"
  description = "A skill without tools or references."
  content     = "Empty content."
}

data "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id = elasticstack_kibana_agentbuilder_skill.test.skill_id
}
