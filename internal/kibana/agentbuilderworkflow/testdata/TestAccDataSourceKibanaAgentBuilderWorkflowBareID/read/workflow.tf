variable "space_id" {
  description = "The Kibana space ID the workflow lives in"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Workflow Bare ID"
  description = "Space for testing agent builder workflow data source bare-ID + explicit space_id input"
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  space_id           = elasticstack_kibana_space.test.space_id
  configuration_yaml = <<-EOT
name: Bare ID Test Workflow
description: A test workflow for verifying bare workflow_id + explicit space_id lookup
enabled: true
triggers:
  - type: manual
steps:
  - name: test_step
    type: console
    with:
      message: "hello from bare id test"
EOT
}

# id is a bare (non-composite) workflow_id, paired with an explicit space_id.
data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id       = elasticstack_kibana_agentbuilder_workflow.test.workflow_id
  space_id = elasticstack_kibana_agentbuilder_workflow.test.space_id
}
