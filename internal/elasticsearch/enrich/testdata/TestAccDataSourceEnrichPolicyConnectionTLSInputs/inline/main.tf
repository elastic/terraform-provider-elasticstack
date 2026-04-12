variable "name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "ca_data" {
  type = string
}

variable "cert_data" {
  type = string
}

variable "key_data" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = var.name

  mappings = jsonencode({
    properties = {
      email      = { type = "keyword" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = var.name
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
  query         = jsonencode({ match_all = {} })
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = elasticstack_elasticsearch_enrich_policy.policy.name

  elasticsearch_connection {
    endpoints = [var.endpoint]
    ca_data   = var.ca_data
    cert_data = var.cert_data
    key_data  = var.key_data
  }
}
