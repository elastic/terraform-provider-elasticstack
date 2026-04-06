variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Workflow Export"
  description = "Space for testing agent builder workflow exports"
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  space_id           = elasticstack_kibana_space.test.space_id
  configuration_yaml = <<-EOT
name: Space Export Test Workflow
description: A test workflow for export in a non-default space
enabled: true
triggers:
  - type: manual
steps:
  - name: test_step
    type: console
    with:
      message: "hello from space"
EOT
}

data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id = elasticstack_kibana_agentbuilder_workflow.test.id
}
