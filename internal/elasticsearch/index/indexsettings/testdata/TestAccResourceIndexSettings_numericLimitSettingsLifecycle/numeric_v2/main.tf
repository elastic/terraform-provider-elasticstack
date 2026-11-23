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

  max_result_window             = 15000
  max_inner_result_window       = 200
  max_rescore_window            = 15000
  max_docvalue_fields_search    = 200
  max_script_fields             = 50
  max_ngram_diff                = 3
  max_shingle_diff              = 5
  max_refresh_listeners         = 1500
  analyze_max_token_count       = 15000
  highlight_max_analyzed_offset = 1500000
  max_terms_count               = 80000
  max_regex_length              = 1500
}
