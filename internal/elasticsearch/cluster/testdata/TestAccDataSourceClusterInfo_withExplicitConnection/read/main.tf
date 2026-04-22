variable "endpoints" {
  description = "List of Elasticsearch endpoints"
  type        = list(string)
}

variable "api_key" {
  description = "Elasticsearch API key (optional)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "username" {
  description = "Elasticsearch username (optional)"
  type        = string
  default     = ""
}

variable "password" {
  description = "Elasticsearch password (optional)"
  type        = string
  default     = ""
  sensitive   = true
}

provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints = var.endpoints
    insecure  = true
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
