variable "endpoint" {
  type        = string
  description = "Elasticsearch endpoint for the resource-level connection override"
}

variable "index_template" {
  type        = string
  description = "Name of the index template"
}

variable "password" {
  type        = string
  description = "Elasticsearch password for explicit basic auth coverage"
}

variable "policy_name" {
  type        = string
  description = "Name of the ILM policy"
}

variable "username" {
  type        = string
  description = "Elasticsearch username for explicit basic auth coverage"
}
