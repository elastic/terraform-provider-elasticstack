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
  name              = "test-export-connector-kbconn"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://example.com"
  })
}

data "elasticstack_kibana_export_saved_objects" "test" {
  objects = [
    {
      type = "action",
      id   = elasticstack_kibana_action_connector.test.connector_id
    }
  ]

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" && var.username != "" ? var.username : null
    password  = var.api_key == "" && var.password != "" ? var.password : null
  }
}
