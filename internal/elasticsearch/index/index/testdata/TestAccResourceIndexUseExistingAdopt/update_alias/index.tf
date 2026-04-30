variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Same alias name as create; add a filter to exercise the update/alias reconciliation path.
resource "elasticstack_elasticsearch_index" "test_use_existing" {
  name             = var.index_name
  use_existing     = true
  number_of_shards = 1

  alias = [
    {
      name = "adopt_alias_step1"
      filter = jsonencode({
        term = { "user.id" = "developer" }
      })
    },
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
