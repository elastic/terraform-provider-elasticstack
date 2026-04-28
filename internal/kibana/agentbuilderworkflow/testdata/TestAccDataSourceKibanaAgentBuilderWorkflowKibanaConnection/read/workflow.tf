variable "kibana_endpoints" {
  description = "Kibana base URLs for the entity-local connection block"
  type        = list(string)
}

variable "api_key" {
  type    = string
  default = ""
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type    = string
  default = ""
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_agentbuilder_workflow" "test" {
  configuration_yaml = <<-EOT
name: KibanaConnection Test Workflow
description: A test workflow for kibana_connection data source coverage
enabled: true
triggers:
  - type: manual
steps:
  - name: test_step
    type: console
    with:
      message: "hello from kibana_connection test"
EOT
}

data "elasticstack_kibana_agentbuilder_workflow" "test" {
  id = elasticstack_kibana_agentbuilder_workflow.test.id

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
