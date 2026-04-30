variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# OOB index has legacy_alias; only new_alias is declared — legacy must be removed on adopt.
resource "elasticstack_elasticsearch_index" "test_use_existing" {
  name             = var.index_name
  use_existing     = true
  number_of_shards = 1

  alias = [
    {
      name = "new_alias"
    },
  ]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
