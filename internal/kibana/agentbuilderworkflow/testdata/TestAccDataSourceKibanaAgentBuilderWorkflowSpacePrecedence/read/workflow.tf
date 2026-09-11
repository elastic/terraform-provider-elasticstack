variable "space_id" {
  description = "The Kibana space ID the workflow actually lives in"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Workflow Space Precedence"
  description = "Space for testing agent builder workflow data source space_id precedence"
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  space_id           = elasticstack_kibana_space.test.space_id
  configuration_yaml = <<-EOT
name: Space Precedence Test Workflow
description: A test workflow for verifying explicit space_id precedence
enabled: true
triggers:
  - type: manual
steps:
  - name: test_step
    type: console
    with:
      message: "hello from space precedence test"
EOT
}

# The id embeds a space segment that does not exist. The explicit space_id
# below must win over it per clients.ResolveCompositeSpaceAndID.
data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id       = "space-does-not-exist/${elasticstack_kibana_agentbuilder_workflow.test.workflow_id}"
  space_id = elasticstack_kibana_agentbuilder_workflow.test.space_id
}
