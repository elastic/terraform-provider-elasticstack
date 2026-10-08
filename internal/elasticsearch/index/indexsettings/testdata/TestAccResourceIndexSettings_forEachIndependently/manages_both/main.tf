variable "index_names" {
  type = set(string)
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  for_each            = var.index_names
  name                = each.value
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  for_each = var.index_names
  index    = elasticstack_elasticsearch_index.test[each.key].name

  number_of_replicas = 2
}
