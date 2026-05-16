variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Out-of-band index has 1 shard; this config requests 2 → adopt must error.
resource "elasticstack_elasticsearch_index" "test_mismatch" {
  name             = var.index_name
  use_existing     = true
  number_of_shards = 2

  deletion_protection = false
}
