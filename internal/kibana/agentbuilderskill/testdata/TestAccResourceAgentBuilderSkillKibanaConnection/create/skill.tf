variable "skill_id" {
  description = "The skill ID"
  type        = string
}

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

resource "elasticstack_kibana_agentbuilder_skill" "test" {
  skill_id    = var.skill_id
  name        = "Skill kibana_connection"
  description = "A skill created via an entity-local Kibana connection."
  content     = "Be helpful."

  kibana_connection {
    endpoints = var.kibana_endpoints
    insecure  = false
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
