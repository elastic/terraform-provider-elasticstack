variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_slowlog" {
  name = var.index_name

  search_slowlog_level                  = "info"
  search_slowlog_threshold_query_warn   = "10s"
  indexing_slowlog_level                = "warn"
  indexing_slowlog_threshold_index_warn = "10s"

  deletion_protection = false
}
