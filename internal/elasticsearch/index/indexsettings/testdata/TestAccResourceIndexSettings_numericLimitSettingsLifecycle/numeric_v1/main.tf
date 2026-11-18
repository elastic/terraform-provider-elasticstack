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

  max_result_window             = 12000
  max_inner_result_window       = 150
  max_rescore_window            = 12000
  max_docvalue_fields_search    = 150
  max_script_fields             = 40
  max_ngram_diff                = 2
  max_shingle_diff              = 4
  max_refresh_listeners         = 1200
  analyze_max_token_count       = 12000
  highlight_max_analyzed_offset = 1200000
  max_terms_count               = 70000
  max_regex_length              = 1200
}
