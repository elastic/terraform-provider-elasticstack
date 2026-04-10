variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "range_index" {
  name = var.name

  mappings = jsonencode({
    properties = {
      ip_range    = { type = "ip_range" }
      department  = { type = "keyword" }
      description = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = var.name
  policy_type   = "range"
  indices       = [elasticstack_elasticsearch_index.range_index.name]
  match_field   = "ip_range"
  enrich_fields = ["department", "description"]
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = elasticstack_elasticsearch_enrich_policy.policy.name
}
