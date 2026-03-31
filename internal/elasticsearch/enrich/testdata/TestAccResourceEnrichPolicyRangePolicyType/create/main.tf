variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = var.name

  mappings = jsonencode({
    properties = {
      range_field = { type = "integer_range" }
      range_label = { type = "keyword" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = var.name
  policy_type   = "range"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "range_field"
  enrich_fields = ["range_label"]
  query = <<-EOD
  {"match_all": {}}
  EOD
}
