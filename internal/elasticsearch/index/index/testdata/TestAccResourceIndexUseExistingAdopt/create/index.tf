variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Matches PreConfig-created index (1 shard); adopt then reconciles alias + mappings.
resource "elasticstack_elasticsearch_index" "test_use_existing" {
  name             = var.index_name
  use_existing     = true
  number_of_shards = 1

  alias = [
    {
      name = "adopt_alias_step1"
    },
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
