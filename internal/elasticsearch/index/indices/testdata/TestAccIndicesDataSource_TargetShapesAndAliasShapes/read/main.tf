variable "index_a" {
  type = string
}

variable "index_b" {
  type = string
}

variable "alias_name" {
  type = string
}

variable "miss_target" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "a" {
  name                = var.index_a
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index" "b" {
  name                = var.index_b
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "test" {
  name = var.alias_name

  write_index = {
    name = elasticstack_elasticsearch_index.a.name
  }
}

data "elasticstack_elasticsearch_indices" "multi" {
  target     = "${var.index_a},${var.index_b}"
  depends_on = [elasticstack_elasticsearch_index_alias.test]
}

data "elasticstack_elasticsearch_indices" "alias_target" {
  target     = var.alias_name
  depends_on = [elasticstack_elasticsearch_index_alias.test]
}

data "elasticstack_elasticsearch_indices" "miss" {
  target     = var.miss_target
  depends_on = [elasticstack_elasticsearch_index_alias.test]
}

data "elasticstack_elasticsearch_indices" "with_alias" {
  target     = var.index_a
  depends_on = [elasticstack_elasticsearch_index_alias.test]
}

data "elasticstack_elasticsearch_indices" "without_alias" {
  target     = var.index_b
  depends_on = [elasticstack_elasticsearch_index_alias.test]
}
