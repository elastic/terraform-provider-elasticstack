provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test_auto" {
  configuration_yaml = <<-EOT
name: Auto ID Workflow
enabled: true
description: This is a new workflow
triggers:
  - type: manual
inputs:
  - name: message
    type: string
    default: "hello world"
steps:
  - name: hello_world_step
    type: console
    with:
      message: "{{ inputs.message }}"
EOT
}
