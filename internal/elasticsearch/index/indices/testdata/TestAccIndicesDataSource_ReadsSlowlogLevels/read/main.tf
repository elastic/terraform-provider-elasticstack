variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                                   = var.index_name
  search_slowlog_threshold_query_warn    = "10s"
  search_slowlog_threshold_fetch_info    = "800ms"
  indexing_slowlog_threshold_index_debug = "10ms"
  indexing_slowlog_source                = "1000"
  search_slowlog_level                   = "info"
  indexing_slowlog_level                 = "warn"
  deletion_protection                    = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
