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

variable "connector_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".index"
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" && var.username != "" ? var.username : null
    password  = var.api_key == "" && var.password != "" ? var.password : null
  }
}
