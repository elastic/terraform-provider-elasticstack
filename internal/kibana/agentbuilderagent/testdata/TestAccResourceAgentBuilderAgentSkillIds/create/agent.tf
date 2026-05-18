variable "agent_id" {
  description = "The agent ID"
  type        = string
}

variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Test Skill for Agent"
  description = "A skill for testing agent skill_ids."
  content     = "Be helpful."
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Test Agent With Skills"
  description  = "An agent that references a skill."
  instructions = "Use the available skills."
  skill_ids    = [elasticstack_kibana_agentbuilder_skill.test.skill_id]
}
