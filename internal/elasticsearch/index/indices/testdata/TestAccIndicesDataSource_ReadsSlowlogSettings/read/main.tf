variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                                   = var.index_name
  search_slowlog_threshold_query_warn    = "10s"
  search_slowlog_threshold_query_info    = "5s"
  search_slowlog_threshold_query_debug   = "2s"
  search_slowlog_threshold_query_trace   = "500ms"
  search_slowlog_threshold_fetch_warn    = "1200ms"
  search_slowlog_threshold_fetch_info    = "800ms"
  search_slowlog_threshold_fetch_debug   = "400ms"
  search_slowlog_threshold_fetch_trace   = "200ms"
  indexing_slowlog_threshold_index_warn  = "1s"
  indexing_slowlog_threshold_index_info  = "200ms"
  indexing_slowlog_threshold_index_debug = "10ms"
  indexing_slowlog_threshold_index_trace = "5ms"
  indexing_slowlog_source                = "1000"
  deletion_protection                    = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
