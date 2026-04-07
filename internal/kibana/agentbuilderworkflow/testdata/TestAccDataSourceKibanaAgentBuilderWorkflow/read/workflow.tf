provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  configuration_yaml = <<-EOT
name: Test Workflow
description: A test workflow for export
enabled: true
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "test message"
steps:
  - name: test_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}

data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id = elasticstack_kibana_agentbuilder_workflow.test.id
}
