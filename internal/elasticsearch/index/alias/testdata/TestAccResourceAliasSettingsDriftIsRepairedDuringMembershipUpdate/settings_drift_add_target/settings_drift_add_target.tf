variable "alias_name" {
  type = string
}

variable "index_name1" {
  type = string
}

variable "index_name2" {
  type = string
}

variable "read_indices_pattern" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "index1" {
  name                = var.index_name1
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index" "index2" {
  name                = var.index_name2
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  read_indices = [{
    name   = var.read_indices_pattern
    filter = jsonencode({ term = { status = "published" } })
  }]

  depends_on = [
    elasticstack_elasticsearch_index.index1,
    elasticstack_elasticsearch_index.index2,
  ]
}
