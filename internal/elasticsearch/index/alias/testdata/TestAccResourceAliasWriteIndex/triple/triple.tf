variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name1" {
  description = "The first index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

variable "index_name3" {
  description = "The third index name"
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

  write_index = {
    name = elasticstack_elasticsearch_index.index3.name
  }

  read_indices = [
    {
      name = elasticstack_elasticsearch_index.index1.name
    },
    {
      name = elasticstack_elasticsearch_index.index2.name
    }
  ]
}
