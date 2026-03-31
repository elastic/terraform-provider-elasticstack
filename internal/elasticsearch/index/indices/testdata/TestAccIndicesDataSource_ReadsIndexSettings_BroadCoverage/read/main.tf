variable "index_name" {
  type = string
}

variable "pipeline_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test" {
  name        = var.pipeline_name
  description = "Acceptance test pipeline"

  processors = [
    jsonencode({ set = { field = "_pipeline_test", value = "1" } })
  ]
}

resource "elasticstack_elasticsearch_index" "test" {
  name                               = var.index_name
  number_of_shards                   = 2
  number_of_replicas                 = 0
  codec                              = "best_compression"
  mapping_coerce                     = false
  max_inner_result_window            = 250
  max_rescore_window                 = 300
  max_docvalue_fields_search         = 50
  max_script_fields                  = 20
  max_shingle_diff                   = 4
  max_refresh_listeners              = 150
  analyze_max_token_count            = 5000
  highlight_max_analyzed_offset      = 200000
  max_terms_count                    = 2048
  max_regex_length                   = 2000
  routing_rebalance_enable           = "replicas"
  blocks_metadata                    = false
  default_pipeline                   = elasticstack_elasticsearch_ingest_pipeline.test.name
  final_pipeline                     = elasticstack_elasticsearch_ingest_pipeline.test.name
  unassigned_node_left_delayed_timeout = "45s"
  deletion_protection                = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
