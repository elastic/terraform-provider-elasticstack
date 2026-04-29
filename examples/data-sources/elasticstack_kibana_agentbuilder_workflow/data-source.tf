provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "doc" {
  configuration_yaml = <<-EOT
name: Doc Workflow
description: Example workflow used by the datasource lookup below
enabled: true
triggers:
  - type: manual
inputs: []
steps:
  - name: noop
    type: console
    with:
      message: "ok"
EOT
}

data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id = elasticstack_kibana_agentbuilder_workflow.doc.id
}
