variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id = var.skill_id
}
