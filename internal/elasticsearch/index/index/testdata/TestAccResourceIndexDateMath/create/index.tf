variable "index_name" {
  description = "The date math index name expression"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_date_math" {
  name = var.index_name

  alias = [
    {
      name = "date_math_alias_1"
    },
  ]

  mappings = jsonencode({
    properties = {
      "@timestamp" = { type = "date" }
    }
  })

  deletion_protection = false
}
