variable "skill_id" {
  description = "The skill ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Invalid Path Skill"
  description = "A skill with an invalid referenced_content relative_path"
  content     = "Always be helpful and accurate."

  referenced_content = [
    {
      name          = "Runbook"
      relative_path = "runbooks/standard.md"
      content       = "Missing the ./ prefix"
    },
  ]
}
