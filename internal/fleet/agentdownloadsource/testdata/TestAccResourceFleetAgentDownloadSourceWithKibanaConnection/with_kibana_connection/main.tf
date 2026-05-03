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
  type      = string
  sensitive = true
  default   = ""
}

variable "suffix" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Kibana Connection Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-kbconn-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = true
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" && var.username != "" ? var.username : null
    password  = var.api_key == "" && var.username != "" ? var.password : null
  }
}
