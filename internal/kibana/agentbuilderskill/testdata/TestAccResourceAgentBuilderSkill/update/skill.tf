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

  tool_ids = ["platform.core.index_explorer"]

  referenced_content = [
    {
      name          = "Runbook"
      relative_path = "./runbooks/standard.md"
      content       = "First entry"
    },
    {
      name          = "Glossary"
      relative_path = "./reference/glossary.md"
      content       = "Second entry"
    },
  ]
}
