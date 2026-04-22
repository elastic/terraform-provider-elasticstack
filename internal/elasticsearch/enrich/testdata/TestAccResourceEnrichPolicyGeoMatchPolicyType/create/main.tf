variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "geo_index" {
  name = var.name

  mappings = jsonencode({
    properties = {
      location    = { type = "geo_shape" }
      name        = { type = "keyword" }
      description = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = var.name
  policy_type   = "geo_match"
  indices       = [elasticstack_elasticsearch_index.geo_index.name]
  match_field   = "location"
  enrich_fields = ["name", "description"]
}
