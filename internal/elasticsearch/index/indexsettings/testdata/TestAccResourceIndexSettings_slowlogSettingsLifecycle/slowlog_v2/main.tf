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

  search_slowlog_threshold_query_warn    = "20s"
  search_slowlog_threshold_query_info    = "8s"
  search_slowlog_threshold_query_debug   = "3s"
  search_slowlog_threshold_query_trace   = "750ms"
  search_slowlog_threshold_fetch_warn    = "20s"
  search_slowlog_threshold_fetch_info    = "8s"
  search_slowlog_threshold_fetch_debug   = "3s"
  search_slowlog_threshold_fetch_trace   = "750ms"
  indexing_slowlog_threshold_index_warn  = "20s"
  indexing_slowlog_threshold_index_info  = "8s"
  indexing_slowlog_threshold_index_debug = "3s"
  indexing_slowlog_threshold_index_trace = "750ms"
  indexing_slowlog_source                = "false"
}
