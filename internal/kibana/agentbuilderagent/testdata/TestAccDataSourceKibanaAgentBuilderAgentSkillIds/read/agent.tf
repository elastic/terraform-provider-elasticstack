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
  name        = "Test Skill for Agent Export"
  description = "A skill for testing agent data source skill_ids export."
  content     = "Be helpful."
}

resource "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id     = var.agent_id
  name         = "Test Agent With Skills"
  description  = "An agent with a skill for export testing."
  instructions = "Use the available skills."
  skill_ids    = [elasticstack_kibana_agentbuilder_skill.test.skill_id]
}

data "elasticstack_kibana_agentbuilder_agent" "test" {
  agent_id = elasticstack_kibana_agentbuilder_agent.test.agent_id
}
