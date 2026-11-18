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

  auto_expand_replicas                 = "0-all"
  search_idle_after                    = "60s"
  gc_deletes                           = "90s"
  default_pipeline                     = "pipeline-three"
  final_pipeline                       = "pipeline-four"
  unassigned_node_left_delayed_timeout = "45s"
}
