provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "my_workflow" {
  configuration = <<-EOT
name: My Workflow
description: An example workflow
enabled: true
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
