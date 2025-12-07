variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "index1" {
  name                = var.index_name
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

  write_index = {
    name          = elasticstack_elasticsearch_index.index1.name
    index_routing = "write-routing"
    filter = jsonencode({
      term = {
        status = "published"
      }
    })
  }

  read_indices = [{
    name = elasticstack_elasticsearch_index.index2.name
    filter = jsonencode({
      term = {
        status = "draft"
      }
    })
  }]
}
