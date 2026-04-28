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
  type      = string
  default   = ""
  sensitive = true
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = "kbconn_connector"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://hooks.example.com/kbconn"
  })
}

data "elasticstack_kibana_action_connector" "test" {
  name = elasticstack_kibana_action_connector.test.name

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" && var.username != "" ? var.username : null
    password  = var.api_key == "" && var.password != "" ? var.password : null
  }
}
