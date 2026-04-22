variable "workflow_id" {
  description = "The workflow ID"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Workflows"
  description = "Space for testing agent builder workflows"
}

resource "elasticstack_kibana_agentbuilder_workflow" "test_space" {
  workflow_id        = var.workflow_id
  space_id           = elasticstack_kibana_space.test.space_id
  configuration_yaml = <<-EOT
name: Space Test Workflow
description: A test workflow in a non-default space
enabled: true
triggers:
  - type: manual
steps:
  - name: hello_step
    type: console
    with:
      message: "hello from space"
EOT
}
