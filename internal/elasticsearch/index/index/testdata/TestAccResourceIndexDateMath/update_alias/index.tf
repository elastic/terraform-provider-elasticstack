variable "index_name" {
  description = "The date math index name expression"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_date_math" {
  name = var.index_name

  # Keep the alias from create step to prove alias is still present after update.
  alias = [
    {
      name = "date_math_alias_1"
    },
  ]

  # Add a new field to the mappings to exercise the update-mappings path targeting
  # the concrete managed index rather than the configured date math expression.
  mappings = jsonencode({
    properties = {
      "@timestamp" = { type = "date" }
      "message"    = { type = "text" }
    }
  })

  deletion_protection = false
}
