variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Timeouts Test Skill"
  description = "A test skill for timeouts block coverage"
  content     = "Always be helpful and accurate."

  timeouts = {
    create = "5m"
  }
}
