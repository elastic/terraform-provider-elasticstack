variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings" {
  name = var.index_name

  mappings = jsonencode({
    properties = {
      field1   = { type = "text" }
      sort_key = { type = "keyword" }
    }
  })

  number_of_shards                     = 2
  number_of_routing_shards             = 2
  codec                                = "best_compression"
  routing_partition_size               = 1
  shard_check_on_startup               = "false"
  sort_field                           = ["sort_key"]
  sort_order                           = ["asc"]
  mapping_coerce                       = true
  mapping_total_fields_limit           = 3000
  auto_expand_replicas                 = "0-5"
  search_idle_after                    = "30s"
  refresh_interval                     = "10s"
  max_result_window                    = 5000
  max_inner_result_window              = 2000
  max_rescore_window                   = 1000
  max_docvalue_fields_search           = 1500
  max_script_fields                    = 500
  max_ngram_diff                       = 100
  max_shingle_diff                     = 200
  max_refresh_listeners                = 10
  analyze_max_token_count              = 500000
  highlight_max_analyzed_offset        = 1000
  max_terms_count                      = 10000
  max_regex_length                     = 1000
  query_default_field                  = ["field1"]
  routing_allocation_enable            = "primaries"
  routing_rebalance_enable             = "primaries"
  gc_deletes                           = "30s"
  unassigned_node_left_delayed_timeout = "5m"

  analysis_char_filter = jsonencode({
    zero_width_spaces = {
      type     = "mapping"
      mappings = ["\\u200C=>\\u0020"]
    }
  })
  analysis_filter = jsonencode({
    minimal_english_stemmer = {
      type     = "stemmer"
      language = "minimal_english"
    }
  })
  analysis_analyzer = jsonencode({
    text_en = {
      type        = "custom"
      tokenizer   = "standard"
      char_filter = "zero_width_spaces"
      filter      = ["lowercase", "minimal_english_stemmer"]
    }
  })

  settings {
    setting {
      name  = "number_of_replicas"
      value = "2"
    }
  }

  deletion_protection = false
}
