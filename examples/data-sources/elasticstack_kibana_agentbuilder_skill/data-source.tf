provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_skill" "doc" {
  skill_id    = "doc-skill"
  name        = "Doc Skill"
  description = "Example skill used by the data source lookup below."
  content     = "Always be helpful and cite your sources."
}

# Look up the skill we just created. `skill_id` accepts either a bare id or
# a composite "<space_id>/<skill_id>" string.
data "elasticstack_kibana_agentbuilder_skill" "doc" {
  skill_id = elasticstack_kibana_agentbuilder_skill.doc.skill_id
}
