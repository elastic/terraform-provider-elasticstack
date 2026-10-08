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

  auto_expand_replicas                 = "0-1"
  search_idle_after                    = "30s"
  gc_deletes                           = "30s"
  default_pipeline                     = "pipeline-one"
  final_pipeline                       = "pipeline-two"
  unassigned_node_left_delayed_timeout = "30s"
}
