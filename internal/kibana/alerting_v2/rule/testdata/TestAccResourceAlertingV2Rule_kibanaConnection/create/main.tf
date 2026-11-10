variable "name" {
  description = "Unique rule name"
  type        = string
}

variable "rule_id" {
  description = "Rule identifier"
  type        = string
}

variable "kibana_endpoints" {
  description = "Kibana endpoints for the resource-level connection"
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
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  rule_id = var.rule_id
  kind    = "alert"

  metadata = {
    name = var.name
  }

  schedule = {
    every = "1m"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"
  }

  recovery = {
    strategy = "no_breach"
  }

  no_data = {
    strategy = "ignore"
  }

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
