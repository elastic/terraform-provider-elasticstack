variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name1" {
  description = "The first matching index name"
  type        = string
}

variable "index_name2" {
  description = "The second matching index name"
  type        = string
}

variable "index_name3" {
  description = "The third matching index name"
  type        = string
}

variable "read_indices_pattern" {
  description = "The wildcard expression for read indices"
  type        = string
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

resource "elasticstack_elasticsearch_index" "index3" {
  name                = var.index_name3
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  read_indices = [{
    name = var.read_indices_pattern
  }]

  depends_on = [
    elasticstack_elasticsearch_index.index1,
    elasticstack_elasticsearch_index.index2,
  ]
}
