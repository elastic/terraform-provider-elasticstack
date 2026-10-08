variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index = elasticstack_elasticsearch_index.test.name

  search_slowlog_threshold_query_warn    = "10s"
  search_slowlog_threshold_query_info    = "5s"
  search_slowlog_threshold_query_debug   = "2s"
  search_slowlog_threshold_query_trace   = "500ms"
  search_slowlog_threshold_fetch_warn    = "10s"
  search_slowlog_threshold_fetch_info    = "5s"
  search_slowlog_threshold_fetch_debug   = "2s"
  search_slowlog_threshold_fetch_trace   = "500ms"
  indexing_slowlog_threshold_index_warn  = "10s"
  indexing_slowlog_threshold_index_info  = "5s"
  indexing_slowlog_threshold_index_debug = "2s"
  indexing_slowlog_threshold_index_trace = "500ms"
  indexing_slowlog_source                = "1000"
}
