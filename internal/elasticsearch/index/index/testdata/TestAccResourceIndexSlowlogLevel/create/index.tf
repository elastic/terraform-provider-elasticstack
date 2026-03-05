variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_slowlog_level" {
  name = var.index_name

  indexing_slowlog_level = "warn"

  deletion_protection = false
}
