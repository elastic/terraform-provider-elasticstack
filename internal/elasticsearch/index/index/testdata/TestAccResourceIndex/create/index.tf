variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  alias = [
    {
      name = "test_alias_1"
    },
    {
      name = "test_alias_2"
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

  wait_for_active_shards = "all"
  master_timeout         = "1m"
  timeout                = "1m"
}
