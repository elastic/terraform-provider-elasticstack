variable "space_id" {
  type = string
}

variable "username" {
  type = string
}

variable "password" {
  type = string
}

variable "role_name" {
  type = string
}

variable "kibana_endpoints" {
  description = "Kibana base URLs for the entity-local connection block"
  type        = list(string)
}
