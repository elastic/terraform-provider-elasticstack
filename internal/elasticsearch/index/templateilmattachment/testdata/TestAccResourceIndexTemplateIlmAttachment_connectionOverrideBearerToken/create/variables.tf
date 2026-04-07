variable "bearer_token" {
  type        = string
  description = "Elasticsearch bearer token for explicit JWT auth coverage"
  sensitive   = true
}

variable "endpoint" {
  type        = string
  description = "Elasticsearch endpoint for the resource-level connection override"
}

variable "index_template" {
  type        = string
  description = "Name of the index template"
}

variable "policy_name" {
  type        = string
  description = "Name of the ILM policy"
}
