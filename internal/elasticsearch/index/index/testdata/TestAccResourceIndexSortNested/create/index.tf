variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  deletion_protection = false

  mappings = jsonencode({
    properties = {
      date     = { type = "date" }
      username = { type = "keyword" }
    }
  })

  sort = [
    {
      field = "date"
      order = "desc"
    },
    {
      field = "username"
      order = "asc"
    },
  ]

  wait_for_active_shards = "all"
  master_timeout         = "1m"
  timeout                = "1m"
}