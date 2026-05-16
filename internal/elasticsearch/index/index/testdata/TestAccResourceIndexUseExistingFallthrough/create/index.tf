variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# No pre-existing index: use_existing falls through to normal Create Index API.
resource "elasticstack_elasticsearch_index" "test_use_existing" {
  name             = var.index_name
  use_existing     = true
  number_of_shards = 1

  alias = [
    {
      name = "fallthrough_alias_1"
    },
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
