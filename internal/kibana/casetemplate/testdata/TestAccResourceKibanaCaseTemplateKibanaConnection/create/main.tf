variable "name" {
  type = string
}

variable "kibana_endpoints" {
  type = list(string)
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

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = var.name
  definition_yaml = <<-EOT
name: Triage case
severity: low
tags:
  - alpha
fields:
  - name: ticket_ref
    type: keyword
    label: Ticket
    control: INPUT_TEXT
EOT

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
